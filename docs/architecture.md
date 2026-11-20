# Architecture

dumps has three parts: a bash CLI that runs on your machine, a Go binary (`dumpsd`) that runs on the server, and two browser scripts that `dumpsd` injects into published pages.

## Components

| Part | Where | Job |
|------|-------|-----|
| `cli/dumps` | your machine | Validates the source, renders the thumbnail, ships raw HTML over ssh/rsync (or copies locally), calls `dumpsd` |
| `dumpsd` | server | `prepare` injects bar, footer and Open Graph; `activate` swaps a dump in atomically; `render` writes home and listing; `apply` runs admin actions from the CLI; `serve` is the admin HTTP API |
| `web/dumps-bar.js` | browser | Defines `<dumps-bar>` (share button, clean mode) and `<dumps-footer>`; optionally bundles `web/dumps-ga.js` for GA4 |
| Reverse proxy | server | Serves `dumps/` to the public (Caddy example in `deploy/Caddyfile`), and the admin listener |
| Tunnel | server | Exposes the proxy to the internet (cloudflared example in `deploy/cloudflared.yml`) |

The CLI only uploads into `dumps-staging/` and temporary files in `dumps-meta/`. `dumpsd` is the only writer of the public `dumps/` folder, the manifests and the listing.

## Server directory layout

All paths are relative to `DUMPS_ROOT`.

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

The public proxy serves only `dumps/`. The manifests, staging area and listing directory are never under it.

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
  "title_locked": false
}
```

## Publish flow

`dumps publish <file|folder>` does the following.

1. **Validate locally.** The source must exist, be a `.html`/`.htm` file or a folder with `index.html`, must not be inside (or contain) a path in `DUMPS_BLOCKED_PATHS`, and symlinks in a folder must stay inside it. The id must match `^[a-z0-9][a-z0-9-]{0,62}$`. Without `--id`, the CLI generates 10 random characters from `a-z2-7` and retries up to 3 times if the id is taken.
2. **Preflight.** Checks `DUMPS_URL`, reaches the server (exit 3 if not), runs `bin/dumpsd version` (exit 4 if missing, a warning if it differs from the repo), and requires the `prepare` and `activate` capabilities.
3. **Build the bundle.** Copies the content to a temp dir (`rsync -aL`, skipping `.DS_Store` and `.git/`), builds `dumps-bar.js` (plus the GA4 snippet when `DUMPS_GA_ID` is set) and renders a 640px JPEG thumbnail with headless Chrome. If the thumbnail fails, a warning is printed and the home uses a placeholder.
4. **Stage.** Clears `dumps-staging/<id>` on the server, then uploads the content to `dumps-staging/<id>/`, the bar to `dumps-staging/<id>.bar.tmp`, the manifest to `dumps-meta/.<id>.json.tmp` and the thumbnail to `dumps-staging/<id>.thumb.tmp`. Nothing public changes yet. The server keeps the `created` date, the `archived` flag and a renamed title of an existing dump.
5. **Prepare.** `dumpsd prepare` reads a JSON spec on stdin and edits the staged files: it injects `<dumps-bar>` and the script tag after `<body>`, `<dumps-footer>` before `</body>`, and Open Graph and Twitter card meta tags in the `<head>`, in every `.html`/`.htm` file (UTF-16 files are skipped, existing tags are respected). It also writes the final byte count into the manifest.
6. **Activate.** `dumpsd activate <id>` takes the lock, backs up the current content, manifest, bar and thumbnail (`dumps-staging/<id>.old` and friends), then renames staging into `dumps/<id>`, moves the manifest, bar and thumbnail into place and re-renders home and listing.
7. **Rollback.** If any step of the swap or the render fails, `activate` restores the backups, removes the staging leftovers, re-renders and exits 4 ("previous version restored"). On success the backups are deleted.

If the CLI fails or is interrupted before activation, a trap removes the staging files and prints "nothing changed". If the connection drops during activation, the CLI exits 4 with an uncertain state and asks you to check with `dumps list`.

The final output is the dump URL: `<DUMPS_URL>/<id>/`.

## Single writer and the lock

`render`, `activate`, and every admin action (`archive`, `unarchive`, `title`, `delete`) take an exclusive `flock` on `dumps-meta/.render.lock` and wait up to 60 seconds. This serializes concurrent publishes and admin clicks. Home and listing are written to a temporary file in the target directory and renamed, so a reader never sees a half-written page. `dumpsd` ignores SIGHUP so a dropped ssh session does not kill it mid-write.

Because `flock` is used, `dumpsd` targets Unix systems (Linux, macOS, Android).

## Admin flow

The admin page is the private listing at `<DUMPS_LIST_URL>` (default `<DUMPS_URL>/_dumps/`). Its buttons call `POST /_dumps/api/<archive|unarchive|title|delete>` with a JSON body on the admin host.

1. The browser reaches the admin host. Your login layer (Cloudflare Access, Tailscale) authenticates the user.
2. The proxy sends `/_dumps/api/*` to `dumpsd serve` on `127.0.0.1:8090` and serves `/_dumps/*` from the listing directory.
3. `dumpsd` verifies the identity, the `Origin` header, the content type and the body, then changes the manifest under the lock and re-renders.

From the terminal, `dumps archive|unarchive|edit|rm <id>` runs `dumpsd apply` on the server instead of the HTTP API. See [auth.md](auth.md) for what `serve` checks.

`archive` removes a dump from the home but keeps its link alive. `delete` removes the content, manifest and thumbnail.

## What runs where

| Where | What |
|-------|------|
| Your machine | `cli/dumps`, headless Chrome, `go build` for `dumpsd` (`dumps install-dumpsd` cross-compiles for `DUMPS_TARGET`) |
| Server | `dumpsd` (short-lived commands over ssh, plus the `serve` service), reverse proxy, tunnel |
| Browser | `dumps-bar.js`, and the admin page script |
| Local mode | All of the above on one machine; `dumpsd serve --auth none --static` replaces the proxy and the tunnel |

## CLI exit codes

| Code | Meaning |
|------|---------|
| 0 | OK |
| 1 | Usage or validation error |
| 2 | Blocked path (`DUMPS_BLOCKED_PATHS`) |
| 3 | Server unreachable |
| 4 | Remote failure |
