# Architecture

dumps has three parts: a bash CLI that runs on your machine, a Go binary (`dumpsd`) that runs on the server, and two browser scripts that `dumpsd` injects into published pages. The CLI only ships raw HTML. Everything that changes the state of the server (inject the bar, activate, build home and listing) is done by `dumpsd`, the single writer.

## Components

| Part | Where | Job |
|------|-------|-----|
| `cli/dumps` | your machine | Validates the source, renders the thumbnail, ships raw HTML over ssh/rsync (or copies locally), calls `dumpsd` |
| `dumpsd` | server | `prepare` injects bar, footer and Open Graph; `activate` swaps a dump in atomically; `render` writes home and listing; `apply` runs admin actions from the CLI; `serve` is the admin HTTP API (see [configuration.md](configuration.md)) |
| `web/dumps-bar.js` | browser | Defines `<dumps-bar>` (share button, clean mode) and `<dumps-footer>` (authors) |
| `web/dumps-ga.js` | browser | Appended to the bar when `DUMPS_GA_ID` is set; loads GA4 and reports `share` and `dumps_clean_mode` |
| `agents/` | your machine | The `dump-this` skill and the `/dump-this` command, linked by `dumps install-agent` ([agent.md](agent.md)) |
| `site/` | GitHub Pages | The project landing page, published at https://grippado.github.io/dumps/ |
| Reverse proxy | server | Serves `dumps/` to the public (Caddy example in `deploy/Caddyfile`), and the admin listener |
| Tunnel | server | Exposes the proxy to the internet (cloudflared example in `deploy/cloudflared.yml`) |

In production the proxy serves the files, and `dumpsd serve` only answers the admin API (`/_dumps/api/*`) and `/healthz`. In local mode, `dumps serve-local` turns on `--static` and `dumpsd` serves everything on loopback, with the dumps and the admin on two different origins ([local.md](local.md)).

The CLI only uploads into `dumps-staging/` and temporary files in `dumps-meta/`. `dumpsd` is the only writer of the public `dumps/` folder, the manifests and the listing.

## Server directory layout

All paths are relative to `DUMPS_ROOT` (default: the ssh home directory; in local mode, `~/.local/share/dumps`).

```text
bin/dumpsd                  the server binary
dumps/                      public docroot
  index.html                home page (written by render)
  <id>/                     one folder per dump
  _/dumps-bar.js            bar, footer and optional GA4 script, shared by all dumps
  _/thumbs/<id>.jpg         thumbnails
dumps-meta/
  <id>.json                 manifest of each dump
  .site.json                optional site config (see configuration.md)
  .render.lock              flock file
dumps-staging/              uploads before activation
admin/                      private listing (index.html), name set by listing_dir
```

The public proxy serves only `dumps/`. The manifests, staging area and listing directory are never under it. The folder `_` holds shared assets and cannot be a dump id, because ids must start with a letter or a digit.

A manifest looks like this:

```json
{
  "v": 1,
  "id": "abc",
  "title": "My page",
  "kind": "file",
  "entry": "index.html",
  "source_name": "page.html",
  "bytes": 2048,
  "created": "2026-01-01T10:00:00-0300",
  "updated": "2026-01-01T10:00:00-0300",
  "url": "https://dumps.example.com/abc/",
  "machine": "laptop",
  "bar": true,
  "footer": true,
  "authors": [{ "name": "Ada Lovelace", "url": "https://example.com" }],
  "thumb": true,
  "archived": false,
  "title_locked": false,
  "source": "cli"
}
```

`kind` is `file` or `dir`. `source` is `cli` or `agent`: `dumps publish --source agent` stores `agent`, and a republish without the flag keeps the previous value. Manifests written before this field existed have no `source`.

## Publish flow

`dumps publish <file|folder>` does the following.

1. **Validate locally.** The source must exist, be a `.html`/`.htm` file or a folder with `index.html`, must not be inside (or contain) a path in `DUMPS_BLOCKED_PATHS`, and symlinks in a folder must stay inside it. The id must match `^[a-z0-9][a-z0-9-]{0,62}$`.
2. **Preflight.** Checks `DUMPS_URL`, reaches the server (exit 3 if not), runs `bin/dumpsd version` (exit 4 if missing, a warning if it differs from the repo), and runs `bin/dumpsd capabilities`. `prepare` and `activate` are required, otherwise exit 4 with a hint to run `dumps install-dumpsd`.
3. **Id.** Uses `--id`, or generates 10 random characters from `a-z2-7` and retries up to 3 times if the id is taken. Republishing the same id replaces the dump and keeps `created`, the archived state, a renamed title and the `source`.
4. **Build the bundle.** Copies the content to a temp dir (`rsync -aL`, skipping `.DS_Store` and `.git/`), builds `dumps-bar.js` (plus the GA4 snippet when `DUMPS_GA_ID` is set) and renders a 640px JPEG thumbnail with headless Chrome. If the thumbnail fails, a warning is printed and the home uses a placeholder.
5. **Stage.** Clears `dumps-staging/<id>` on the server, then uploads the content to `dumps-staging/<id>/`, the bar to `dumps-staging/<id>.bar.tmp`, the manifest to `dumps-meta/.<id>.json.tmp` and the thumbnail to `dumps-staging/<id>.thumb.tmp`. Nothing public changes yet.
6. **Prepare.** `dumpsd prepare` reads a JSON spec on stdin and edits the staged files: it injects `<dumps-bar>` and the script tag after `<body>`, `<dumps-footer>` before `</body>`, and Open Graph and Twitter card meta tags in the `<head>`, in every `.html`/`.htm` file (UTF-16 files are skipped, existing tags are respected). It also writes the final byte count into the manifest.
7. **Activate.** The CLI runs `dumpsd activate <id>`. Under the lock, it backs up the current content, manifest, bar and thumbnail (`dumps-staging/<id>.old` and friends), then renames staging into `dumps/<id>`, moves the manifest, bar and thumbnail into place and re-renders home and listing. The lock is taken by `dumpsd` itself, so the server needs no `flock` command and a macOS server works.
8. **Rollback.** If any step of the swap or the render fails, `activate` restores the backups, removes the staging leftovers, re-renders and exits 4 ("previous version restored"). On success the backups are deleted.
9. **Result.** The CLI prints the dump URL: `<DUMPS_URL>/<id>/`.

If the CLI fails or is interrupted before activation, a trap removes the staging files and prints that nothing changed. If the connection drops during activation, the CLI exits 4 with an uncertain state and asks you to check with `dumps list`.

## Single writer and the lock

`render`, `activate`, and every admin action (`archive`, `unarchive`, `title`, `delete`) take an exclusive `flock` on `dumps-meta/.render.lock` and wait up to 60 seconds. This serializes concurrent publishes and admin clicks. Home and listing are written to a temporary file in the target directory, synced and renamed, so a reader never sees a half-written page. `dumpsd` ignores SIGHUP so a dropped ssh session does not kill it mid-write.

`render --no-lock` exists for callers that already hold the lock. Do not use it day to day.

Because `flock` is used, `dumpsd` targets Unix systems (Linux, macOS, Android).

## Admin flow

The admin page is the private listing at `<DUMPS_LIST_URL>`. In production the default is `<DUMPS_URL>/_dumps/` unless you set `DUMPS_LIST_URL` to the admin host. In local mode it is `http://localhost:8091/_dumps/`. Its buttons call `POST /_dumps/api/<archive|unarchive|title|delete>` with `{"id": "..."}` (plus `"title"` for `title`).

1. The browser reaches the admin host. Your login layer (Cloudflare Access, Tailscale) authenticates the user.
2. The proxy sends `/_dumps/api/*` to `dumpsd serve` on `127.0.0.1:8090` and serves `/_dumps/*` from the listing directory.
3. `dumpsd` verifies the identity, the `Origin` header, the content type and the body ([auth.md](auth.md)), then changes the manifest under the lock and re-renders. If the render fails, the previous manifest is restored.

In local mode there is no proxy: `dumpsd serve --auth none --static` has two listeners, the dumps on `DUMPS_LISTEN` and the admin (listing and API) on `DUMPS_ADMIN_LISTEN`. The API accepts only the admin origin, so JavaScript in a dump cannot call it.

From the terminal, `dumps archive|unarchive|edit|rm <id>` runs `dumpsd apply` on the server over ssh instead of the HTTP API. It does not go through the login, because whoever gets there already has ssh.

`archive` removes a dump from the home but keeps its link alive. `delete` removes the content, manifest and thumbnail.

In the listing, a dump published with `--source agent` shows a **do agente** badge. If it is also archived, its toggle button reads **Publicar** (highlighted) instead of **Desarquivar**, and it calls the same `unarchive` action. See [agent.md](agent.md).

## What runs where

| Where | What |
|-------|------|
| Your machine | `cli/dumps`, headless Chrome (thumbnails), the assembly of the bar and the JSON spec, `go build` for `dumpsd` (`dumps install-dumpsd` cross-compiles for `DUMPS_TARGET`) |
| Your machine to the server | The upload (rsync over ssh) |
| Server | `dumpsd` (`prepare`, `activate`, `render`, `apply` as short-lived commands over ssh, plus the `serve` service), reverse proxy, tunnel |
| Browser | `dumps-bar.js`, and the admin page script |
| Local mode | All of the above on one machine; `dumpsd serve --auth none --static` replaces the proxy and the tunnel |

The server needs `rsync` and `bin/dumpsd`. It does not need Go, Chrome or jq.

## CLI exit codes

| Code | Meaning |
|------|---------|
| 0 | OK |
| 1 | Usage or validation error |
| 2 | Blocked path (`DUMPS_BLOCKED_PATHS`) |
| 3 | Server unreachable |
| 4 | Remote failure |

`dumpsd apply` also uses 5 (action applied, but home and listing were not regenerated; run `dumps regen`) and 6 (the dump does not exist). The CLI turns both into messages.
