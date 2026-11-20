# Run it locally

Local mode needs no server, no ssh, no proxy and no tunnel. Everything lives on your machine and is served on loopback. It works on macOS and Linux. With `DUMPS_HOST=local` (the default), the CLI copies files with `rsync` into a local root and runs `dumpsd` there.

## Requirements

- `bash`, `jq`, `rsync` and `python3`
- Go, at the version in `server/go.mod`, to build `dumpsd`
- Optional, for thumbnails: Chrome or Chromium, plus `sips` (built into macOS) or ImageMagick (Linux)

## Setup

Local mode is the default, so no config file is needed. The defaults are:

| Setting | Local default |
|---------|---------------|
| Root | `${XDG_DATA_HOME:-~/.local/share}/dumps` (`DUMPS_ROOT`) |
| Public URL (dumps origin) | `http://localhost:8090` (`DUMPS_URL`) |
| Dumps listener | `127.0.0.1:8090` (`DUMPS_LISTEN`) |
| Admin listener | `127.0.0.1:8091` (`DUMPS_ADMIN_LISTEN`) |
| Admin origin | `http://localhost:8091`, from the port of `DUMPS_ADMIN_LISTEN` (or `DUMPS_ADMIN_ORIGIN` if set) |
| Admin page | `http://localhost:8091/_dumps/` (`DUMPS_LIST_URL`) |

Build and install `dumpsd` into the root. It runs the Go tests first and builds for the current machine.

```bash
cli/dumps install-dumpsd
```

Publish.

```bash
cli/dumps publish ./page.html
```

```bash
cli/dumps publish ./my-folder --title "My folder" --id my-folder
```

Serve the dumps and the admin on loopback. The command re-renders home and listing, then runs `dumpsd serve --auth none --static` with both listeners. Stop with Ctrl-C.

```bash
cli/dumps serve-local
```

Browse:

- http://localhost:8090 for the home
- http://localhost:8090/<id>/ for a dump
- http://localhost:8091/_dumps/ for the admin listing (archive, rename, delete)

Open the admin with `localhost`, not `127.0.0.1`. The API compares the `Origin` header as exact text, so on `http://127.0.0.1:8091` the admin actions fail with 403.

Manage from the terminal:

```bash
cli/dumps list
```

```bash
cli/dumps archive <id>
```

Show the effective configuration:

```bash
cli/dumps config
```

## Two origins

`dumps serve-local` starts one `dumpsd` process with two listeners, so the dumps and the admin are different origins (different ports):

| Origin | Address | What it serves |
|--------|---------|----------------|
| Dumps | `DUMPS_LISTEN` (default `127.0.0.1:8090`) | Only `dumps/`: the home, every dump and the thumbnails. `/_dumps` returns 404 and dotfile paths return 404 |
| Admin | `DUMPS_ADMIN_LISTEN` (default `127.0.0.1:8091`) | `/` redirects to `/_dumps/`, which is the listing. The API lives at `/_dumps/api/*`. Anything else returns 404 |

The API accepts only the admin origin: the `Origin` header of every `POST` must equal the admin origin exactly. The admin origin is `DUMPS_ADMIN_ORIGIN`, or `http://localhost:<port of DUMPS_ADMIN_LISTEN>` when that is empty. JavaScript inside a dump runs on the dumps origin, which has no API, and its `Origin` does not match, so a dump cannot archive, rename or delete anything.

Rules enforced by `dumpsd serve --static`:

- It only works with `--auth none`.
- It requires `--admin-listen`, on a different port than `--listen`.
- Both addresses must be loopback (`127.0.0.1:`, `localhost:` or `[::1]:`).

To use other ports, set the three values together. Because `DUMPS_ADMIN_ORIGIN` is empty, the admin origin follows the new admin port.

```bash
DUMPS_URL=http://localhost:9000 DUMPS_LISTEN=127.0.0.1:9000 DUMPS_ADMIN_LISTEN=127.0.0.1:9001 cli/dumps serve-local
```

If `DUMPS_URL` and `DUMPS_LISTEN` disagree, the links break. The `Origin` check breaks when `DUMPS_ADMIN_ORIGIN` is set to something other than the origin you open the admin from.

## Where the data lives

The default root is `~/.local/share/dumps` (or `$XDG_DATA_HOME/dumps`). Change it with `DUMPS_ROOT`. The layout is described in [architecture.md](architecture.md).

## Thumbnails

`dumps publish` screenshots `index.html` at 1280x800 with headless Chrome, then resizes it to a 640px JPEG.

- Chrome lookup order: `DUMPS_CHROME`, then Google Chrome, Brave and Chromium in `/Applications` (macOS), then `google-chrome`, `chromium` and `chromium-browser` on `PATH`.
- Resize: `sips` on macOS, `magick` or `convert` (ImageMagick) on Linux.

Without a browser or a resizer, publishing still works. The CLI prints a warning and the home shows a placeholder with the first letter of the title. Open Graph tags are then added without an image.

The page is opened through `file://`, so references to absolute paths (`/assets/x.css`) do not load in the thumbnail. Use relative paths in the folders you publish.

## Notes

- `--auth none` accepts only direct loopback connections. It also refuses any request that carries `X-Forwarded-For`, `Forwarded`, `Cf-Connecting-Ip`, `X-Real-Ip` or `Cf-Access-Jwt-Assertion`. `dumpsd serve` refuses to listen on a non-loopback address. See [auth.md](auth.md).
- `--auth none` is only for `serve-local`. The CLI refuses `DUMPS_AUTH=none` when it installs a service.
- A dump's JavaScript cannot call the admin API. Dumps still share one origin with each other, so publish only HTML you trust. See [security.md](security.md).
- With the default `DUMPS_SERVICE=none`, no service is installed. Run `dumps serve-local` when you need it.
