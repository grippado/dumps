# Configuration

The CLI reads `DUMPS_*` keys from the environment and from a config file. `dumpsd` reads no file: it receives everything through flags.

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
- There are no trailing comments: everything after the `=` is the value.
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
| `DUMPS_LIST_URL` | local: `http://localhost:<DUMPS_ADMIN_LISTEN port>/_dumps/`; remote: `$DUMPS_URL/_dumps/` | Where the private listing lives (informational, shown by `dumps config`). On a remote host the admin lives on another hostname, so set it |
| `DUMPS_AUTHORS` | empty | Default footer authors, `Name\|URL`, entries separated by `;`. `--author` on publish overrides it |
| `DUMPS_BLOCKED_PATHS` | empty | Colon-separated paths. Publishing anything inside them, or a folder that contains them, exits with code 2. Each item expands `~` on its own: `~` and `~/...` mean your home directory |
| `DUMPS_SITE` | empty | Path to a `site.json` pushed to the server on `install-dumpsd` (see below) |
| `DUMPS_SERVICE` | `none` | Admin service to install: `none`, `termux-runit`, `systemd-user` or `launchd` |
| `DUMPS_TARGET` | `linux/arm64` | `GOOS/GOARCH` of the remote server for `install-dumpsd`. Ignored in local mode (uses this machine) |
| `DUMPS_LISTEN` | `127.0.0.1:8090` | Address of `dumpsd serve`. Must be loopback. In `serve-local` it is the dumps listener |
| `DUMPS_ADMIN_LISTEN` | `127.0.0.1:8091` | Admin listener of `dumps serve-local`. Must be loopback and on a different port than `DUMPS_LISTEN`. Not used by a service |
| `DUMPS_ADMIN_ORIGIN` | empty | Exact origin of the admin page (for example `https://admin.example.com`). Required to install a service. `serve-local` falls back to `http://localhost:<DUMPS_ADMIN_LISTEN port>` |
| `DUMPS_AUTH` | `access` | Login mode of the installed service: `access` or `tailscale`. See [auth.md](auth.md). `none` is refused for a service, and `serve-local` always uses `none` by itself |
| `DUMPS_ACCESS_TEAM` | empty | Cloudflare team domain (for example `yourteam.cloudflareaccess.com`). Required for `access` |
| `DUMPS_ACCESS_AUD` | empty | Application Audience (AUD) tag of the admin Access app. Required for `access` |
| `DUMPS_TAILSCALE_ALLOW` | empty | Space-separated logins allowed in `tailscale` mode. Empty allows any tailnet login |
| `DUMPS_GA_ID` | empty | GA4 measurement ID (`G-XXXXXXX`). Empty disables analytics. Anything else is rejected. Starts with consent denied and shows an accept banner (see security.md) |
| `DUMPS_MACHINE` | `hostname -s` | Name recorded in the manifest and shown in the listing |
| `DUMPS_CHROME` | auto-detect | Path to the Chrome or Chromium binary used for thumbnails |

The values that go into the service (the resolved root, `DUMPS_LISTEN`, `DUMPS_URL`, `DUMPS_ADMIN_ORIGIN`, `DUMPS_ACCESS_TEAM`, `DUMPS_ACCESS_AUD` and each login in `DUMPS_TAILSCALE_ALLOW`) may only contain letters, digits and `: / . _ @ -`. Spaces, quotes and any other character make `install-dumpsd` stop.

The CLI needs `bash`, `jq`, `rsync`, `python3` and `ssh` (for remote hosts). `go` is needed for `install-dumpsd`.

### CLI commands

```text
dumps publish <file.html|folder> [--id <id>] [--title <text>] [--author "Name|URL"]... [--no-bar] [--no-footer] [--hidden] [--source cli|agent]
dumps list
dumps archive <id>
dumps unarchive <id>
dumps edit <id> --title <text>
dumps rm <id>
dumps regen
dumps install-dumpsd
dumps serve-local
dumps config
dumps install-agent [claude|codex|all]
```

| Command | What it does |
|---------|--------------|
| `publish` | Publishes a file or folder. Needs `prepare` and `activate` in the server binary |
| `list` | Lists the dumps by update date |
| `archive`, `unarchive` | Takes a dump off the home without taking it offline, or puts it back |
| `edit` | Changes the title on the home and in the listing. The bar inside the HTML changes only on the next publish |
| `rm` | Deletes the content, manifest and thumbnail |
| `regen` | Rebuilds home and listing from the manifests |
| `install-dumpsd` | Runs the Go tests, builds `dumpsd` for the target, installs it to `$DUMPS_ROOT/bin/dumpsd` (the previous one stays if the new one does not run), pushes `DUMPS_SITE` if set, and installs or updates the service |
| `serve-local` | Serves the dumps on `DUMPS_LISTEN` and the admin on `DUMPS_ADMIN_LISTEN`. Only with `DUMPS_HOST=local`. See [local.md](local.md) |
| `config` | Prints the effective configuration: config file, host, root, URLs, service, target, auth, listeners, admin origin, site, blocked paths, authors, machine and GA |
| `install-agent` | Symlinks the `dump-this` skill and the `/dump-this` command for Claude Code, Codex or both (default `all`). See [agent.md](agent.md) |

Notes on `publish`:

- `--hidden` publishes already archived: the link works, but the dump stays off the home.
- `--source` takes `cli` or `agent` and is stored as `source` in the manifest. Without it, a republish keeps the previous source, and a new dump gets `cli`.
- `--title` defaults to the HTML `<title>`, then to the file name. A title set with `--title` or `edit` is kept on later publishes.
- Ids match `^[a-z0-9][a-z0-9-]{0,62}$`.

Exit codes are in [architecture.md](architecture.md#cli-exit-codes).

## site.json

`site.json` customizes the home page. Point `DUMPS_SITE` at it and run `dumps install-dumpsd`. The file is stored as `dumps-meta/.site.json` and the home is re-rendered. To change the texts later, edit the file and run `install-dumpsd` again: `dumps regen` does not push it. See `examples/site.json`.

The file must be a JSON object. Unknown fields are rejected. All fields are optional.

| Field | Default | Meaning |
|-------|---------|---------|
| `title` | host of `DUMPS_URL` | Page title and brand |
| `description` | `Artifacts published with dumps.` | Meta description |
| `tagline_html` | `description`, HTML-escaped | Tagline under the title (raw HTML) |
| `host_card_html` | empty | Optional block near the bottom of the home (raw HTML) |
| `footer_html` | empty | Appended to the footer of the home, after the host link (raw HTML) |
| `featured` | empty | Highlighted links at the top of the home: `[{ "id": "...", "label": "...", "note": "..." }]` |
| `listing_dir` | `admin` | Directory under the root where the private listing is written |

How `featured` works:

- Each listed id that has a dump becomes one link block at the top of the home, with the label, the dump title and the optional note. The lookup covers every manifest, archived dumps included.
- `label` defaults to `Em destaque`. `note` is optional. Both are HTML-escaped.
- An id with no matching dump is ignored. An id that is not a valid dump id makes the whole file invalid.

How `listing_dir` is validated:

- It must be a relative path inside the root. It cannot be absolute, be `.`, or start with `..`.
- Its first component, compared without case, cannot be `dumps`, `dumps-meta`, `dumps-staging` or `bin`, and cannot start with a dot.
- Keep it outside the public docroot. If you change it, change the `root` of the admin block in the Caddyfile too.

How the file is pushed: `install-dumpsd` checks that it is a JSON object, uploads it and renders. If `dumpsd` refuses it (unknown field, invalid `listing_dir`, invalid `featured` id), the previous `site.json` stays in place, or none if there was none, and the CLI exits 4 with `site.json recusado pelo dumpsd; o anterior foi mantido`.

Warning: `tagline_html`, `host_card_html` and `footer_html` are inserted into the home as raw, trusted HTML. They are not sanitized. Only put content you wrote.

## dumpsd

`dumpsd` is installed at `$DUMPS_ROOT/bin/dumpsd`. Usage errors exit 1 and failures exit 4.

`render`, `apply`, `serve` and `activate` accept `--root <dir>`. Without it, `dumpsd` uses the current directory only if it contains `dumps-meta/`. Otherwise it fails, and asks you to run it inside the root or to pass `--root`. `prepare` works on the folder given in `--dir` and has no `--root`.

### version and capabilities

```bash
bin/dumpsd version
```

```bash
bin/dumpsd capabilities
```

`capabilities` prints `render serve apply prepare activate`. The CLI uses it to detect an outdated server binary: `dumps publish` requires `prepare` and `activate`, and exits 4 with a hint to run `dumps install-dumpsd` when one is missing.

### render

```text
dumpsd render --base-url <url> [--ga] [--no-lock] [--root <dir>]
```

Rebuilds `dumps/index.html` and the listing from the manifests, under the lock.

- `--base-url`: required.
- `--ga`: load `dumps-bar.js` on the home so GA4 also covers it. Also on automatically when the installed bar already contains the GA4 snippet.
- `--no-lock`: skip the lock. For callers that already hold it. Do not use it day to day.

### prepare

```text
dumpsd prepare --dir <folder> [--manifest <file>]
```

Reads a JSON spec from stdin (up to 64 KiB) and injects bar, footer and Open Graph tags into every HTML file of `--dir`. With `--manifest`, it also stores the total size in the manifest. All spec fields are required, unknown fields are rejected: `title`, `url`, `version`, `authors` (a JSON array encoded as a string), `bar`, `footer`, `og_image`, `og_authors`. The CLI builds this for you.

### activate

```text
dumpsd activate --base-url <url> [--ga] [--root <dir>] <id>
```

Activates a staged publish. Under the lock, it backs up the current content, manifest, bar and thumbnail, swaps `dumps-staging/<id>` into `dumps/<id>`, moves manifest, bar and thumbnail into place and re-renders home and listing. If any step fails, it restores the backups and exits 4. It needs `dumps-staging/<id>` and the temporary manifest `dumps-meta/.<id>.json.tmp`, otherwise it fails with `staging incompleto`.

The CLI calls it over ssh (or locally), so the server needs no `flock` command: the lock is taken by `dumpsd` itself, and this works on macOS. See [architecture.md](architecture.md#publish-flow).

### apply

```text
dumpsd apply --base-url <url> [--root <dir>] <archive|unarchive|title|delete> <id>
```

Runs an admin action from the shell. `title` reads the new title from stdin. Besides 0, 1 and 4, it exits 5 when the change was applied but home and listing were not regenerated (run `dumps regen`), and 6 when the dump does not exist. `delete` also removes an orphaned folder with no manifest.

### serve

```text
dumpsd serve --base-url <url> --origin <url> [--listen <addr>] [--admin-listen <addr>] [--root <dir>] [--auth access|tailscale|none] [--team <host>] [--aud <tag>] [--allow <login>]... [--static]
```

Runs the admin HTTP API.

| Flag | Default | Meaning |
|------|---------|---------|
| `--base-url` | required | Public URL |
| `--origin` | required | Exact origin accepted on the `POST`s: the origin of the admin page |
| `--listen` | `127.0.0.1:8090` | Must start with `127.0.0.1:`, `localhost:` or `[::1]:` |
| `--admin-listen` | empty | Admin listener, loopback only. Required with `--static`, on a different port than `--listen` |
| `--root` | current directory with `dumps-meta/` | Root of the dumps |
| `--auth` | `access` | `access` requires `--team` and `--aud`. `tailscale` takes optional repeated `--allow`. `none` is for local mode. See [auth.md](auth.md) |
| `--team` | empty | Access team host (required with `access`) |
| `--aud` | empty | Access application AUD (required with `access`) |
| `--allow` | empty | Login allowed with `tailscale`. Repeat the flag for several |
| `--static` | off | Serve the public site and the admin from this process. Only with `--auth none` and `--admin-listen`. This is what `dumps serve-local` uses |

Without `--static`, one listener serves the API under `/_dumps/api/*` and `/healthz`, and every other path returns 404. In production the proxy serves the files.

With `--static`, there are two listeners. `--listen` serves only `dumps/`, and `--admin-listen` serves the listing at `/_dumps/` and the API. The API accepts only the admin origin.

`GET /healthz` answers `ok` without authentication.

## Makefile

| Target | What it does |
|--------|--------------|
| `make test` | `go test -race ./...` in `server/` |
| `make lint` | gofmt, `go vet` and shellcheck |
| `make build` | Builds `dumpsd` for this machine into `bin/` |
| `make check` | `lint` and `test`, the same as CI |
