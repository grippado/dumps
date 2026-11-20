# Configuration

## Where the CLI reads config

The CLI reads `$DUMPS_CONFIG`, which defaults to `${XDG_CONFIG_HOME:-$HOME/.config}/dumps/config`. Start from the example:

```bash
mkdir -p ~/.config/dumps
```

```bash
cp examples/config ~/.config/dumps/config
```

Rules:

- One `DUMPS_KEY=value` per line. Lines that do not match are ignored (comments, blank lines).
- Values may be wrapped in single or double quotes. `$HOME` and a leading `~/` are expanded.
- Environment wins. A key already set in the environment is not overridden by the file.
- `DUMPS_CONFIG` itself can only come from the environment.

Show the effective configuration:

```bash
cli/dumps config
```

## CLI keys

| Key | Default | Meaning |
|-----|---------|---------|
| `DUMPS_HOST` | `local` | `local` publishes on this machine. Anything else is an ssh host name (from `~/.ssh/config`) |
| `DUMPS_ROOT` | local: `${XDG_DATA_HOME:-~/.local/share}/dumps`; remote: the ssh home | Directory on the server that holds `dumps/`, `dumps-meta/`, `bin/`. Must not contain a single quote |
| `DUMPS_URL` | local: `http://localhost:8090`; remote: none, required | Public base URL, format `http(s)://host[:port]` with no path or trailing slash |
| `DUMPS_LIST_URL` | `$DUMPS_URL/_dumps/` | Where the private listing lives (informational, shown by `dumps config`) |
| `DUMPS_AUTHORS` | empty | Default footer authors, `Name\|URL`, entries separated by `;`. `--author` on publish overrides it |
| `DUMPS_BLOCKED_PATHS` | empty | Colon-separated paths. Publishing anything inside them, or a folder that contains them, exits with code 2 |
| `DUMPS_SITE` | empty | Path to a `site.json` pushed to the server on `install-dumpsd` (see below) |
| `DUMPS_SERVICE` | `none` | Admin service to install: `none`, `termux-runit`, `systemd-user` or `launchd` |
| `DUMPS_TARGET` | `linux/arm64` | `GOOS/GOARCH` of the remote server for `install-dumpsd`. Ignored in local mode (uses this machine) |
| `DUMPS_LISTEN` | `127.0.0.1:8090` | Address of `dumpsd serve`. Must be loopback |
| `DUMPS_ADMIN_ORIGIN` | empty | Exact origin of the admin page (for example `https://admin.example.com`). Required to install a service. `serve-local` falls back to `DUMPS_URL` |
| `DUMPS_AUTH` | `access` | `access`, `tailscale` or `none`. See [auth.md](auth.md) |
| `DUMPS_ACCESS_TEAM` | empty | Cloudflare team domain (for example `yourteam.cloudflareaccess.com`). Required for `access` |
| `DUMPS_ACCESS_AUD` | empty | Application Audience (AUD) tag of the admin Access app. Required for `access` |
| `DUMPS_TAILSCALE_ALLOW` | empty | Space-separated logins allowed in `tailscale` mode. Empty allows any tailnet login |
| `DUMPS_GA_ID` | empty | GA4 measurement ID (`G-XXXXXXX`). Empty disables analytics. Anything else is rejected |
| `DUMPS_MACHINE` | `hostname -s` | Name recorded in the manifest and shown in the listing |
| `DUMPS_CHROME` | auto-detect | Path to the Chrome or Chromium binary used for thumbnails |

The CLI needs `bash`, `jq`, `rsync`, `python3` and `ssh` (for remote hosts). `go` is needed for `install-dumpsd`.

### CLI commands

```text
dumps publish <file.html|folder> [--id <id>] [--title <text>] [--author "Name|URL"]... [--no-bar] [--no-footer] [--hidden]
dumps list
dumps archive <id>
dumps unarchive <id>
dumps edit <id> --title <text>
dumps rm <id>
dumps regen
dumps install-dumpsd
dumps serve-local
dumps config
```

- `--hidden` publishes already archived: the link works, but the dump stays off the home.
- `--title` defaults to the HTML `<title>`, then to the file name. A title set with `--title` or `edit` is kept on later publishes.
- `edit` changes the title on the home and listing. The bar inside the HTML changes only on the next publish.
- `regen` rebuilds home and listing from the manifests.
- `install-dumpsd` runs the Go tests, builds `dumpsd` for the target, installs it to `$DUMPS_ROOT/bin/dumpsd` (the previous one stays if the new one does not run), pushes `DUMPS_SITE` if set, and installs or updates the service.

Exit codes are in [architecture.md](architecture.md#cli-exit-codes).

## site.json

`site.json` customizes the home page. Point `DUMPS_SITE` at it and run `dumps install-dumpsd`. The file is stored as `dumps-meta/.site.json` and the home is re-rendered. See `examples/site.json`.

Unknown fields are rejected. All fields are optional.

| Field | Default | Meaning |
|-------|---------|---------|
| `title` | host of `DUMPS_URL` | Page title and brand |
| `description` | `Artifacts published with dumps.` | Meta description |
| `tagline_html` | `description`, HTML-escaped | Tagline under the title (raw HTML) |
| `host_card_html` | empty | Optional block near the bottom of the home (raw HTML) |
| `footer_html` | empty | Appended to the footer of the home, after the host link (raw HTML) |
| `featured` | empty | Links shown above the cards: `[{ "id": "...", "label": "...", "note": "..." }]`. `id` must be a valid dump id, `label` defaults to `Em destaque`, `note` is optional. Ids with no matching dump are skipped |
| `listing_dir` | `admin` | Directory under the root where the private listing is written |

`listing_dir` must be a relative path inside the root. It cannot be `.`, absolute, contain `..`, or be `dumps`, inside `dumps/`, or start with `dumps-meta`. Keep it outside the public docroot.

Warning: `tagline_html`, `host_card_html` and `footer_html` are inserted into the home as raw, trusted HTML. They are not sanitized. Only put content you wrote.

## dumpsd

`dumpsd` is installed at `$DUMPS_ROOT/bin/dumpsd`. Run it from the root or pass `--root`. Every subcommand defaults `--root` to the current directory. Usage errors exit 1 and failures exit 4.

### version and capabilities

```bash
bin/dumpsd version
```

```bash
bin/dumpsd capabilities
```

`capabilities` prints `render serve apply prepare activate`. The CLI uses it to detect an outdated server binary.

### render

```text
dumpsd render --base-url <url> [--ga] [--no-lock] [--root <dir>]
```

Rebuilds `dumps/index.html` and the listing from the manifests, under the lock.

- `--base-url`: required.
- `--ga`: load `dumps-bar.js` on the home so GA4 also covers it. Also on automatically when the installed bar already contains the GA4 snippet.
- `--no-lock`: skip the lock. For callers that already hold it.

### prepare

```text
dumpsd prepare --dir <folder> [--manifest <file>]
```

Reads a JSON spec from stdin and injects bar, footer and Open Graph tags into every HTML file of `--dir`. With `--manifest`, it also stores the total size in the manifest. All spec fields are required, unknown fields are rejected: `title`, `url`, `version`, `authors` (a JSON array encoded as a string), `bar`, `footer`, `og_image`, `og_authors`. The CLI builds this for you.

### activate

```text
dumpsd activate --base-url <url> [--ga] [--root <dir>] <id>
```

Swaps `dumps-staging/<id>` into `dumps/<id>` with rollback. See [architecture.md](architecture.md#publish-flow).

### apply

```text
dumpsd apply --base-url <url> [--root <dir>] <archive|unarchive|title|delete> <id>
```

Runs an admin action from the shell. `title` reads the new title from stdin. Besides 0, 1 and 4, it exits 5 when the change was applied but home and listing were not regenerated (run `dumps regen`), and 6 when the dump does not exist. `delete` also removes an orphaned folder with no manifest.

### serve

```text
dumpsd serve --base-url <url> --origin <url> [--listen <addr>] [--root <dir>] [--auth access|tailscale|none] [--team <host>] [--aud <tag>] [--allow <login>]... [--static]
```

Runs the admin HTTP API.

- `--base-url`, `--origin`: required. `--origin` is the exact origin of the admin page.
- `--listen`: default `127.0.0.1:8090`. Must start with `127.0.0.1:`, `localhost:` or `[::1]:`.
- `--auth`: default `access`. `access` requires `--team` and `--aud`. `tailscale` takes optional repeated `--allow`. See [auth.md](auth.md).
- `--static`: also serve the public site and the listing (`/_dumps/`) from this process. Only allowed with `--auth none`. This is what `dumps serve-local` uses.

`GET /healthz` answers `ok` without authentication.
