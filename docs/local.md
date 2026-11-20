# Run it locally

Local mode needs no server, no ssh, no proxy and no tunnel. Everything lives on your machine and is served on loopback. It works on macOS and Linux.

## Requirements

- `bash`, `jq`, `rsync` and `python3`
- Go, at the version in `server/go.mod`, to build `dumpsd`
- Optional, for thumbnails: Chrome or Chromium, plus `sips` (built into macOS) or ImageMagick (Linux)

## Setup

Local mode is the default (`DUMPS_HOST=local`), so no config file is needed. With it, the defaults are:

| Setting | Local default |
|---------|---------------|
| Root | `${XDG_DATA_HOME:-~/.local/share}/dumps` (`DUMPS_ROOT`) |
| Public URL | `http://localhost:8090` (`DUMPS_URL`) |
| Listen address | `127.0.0.1:8090` (`DUMPS_LISTEN`) |

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

Serve everything (public dumps, home and admin) on loopback. It re-renders home and listing first, then runs `dumpsd serve --auth none --static`. Stop with Ctrl-C.

```bash
cli/dumps serve-local
```

Browse:

- http://localhost:8090 for the home
- http://localhost:8090/<id>/ for a dump
- http://localhost:8090/_dumps/ for the admin listing (archive, rename, delete)

To use another port, set `DUMPS_URL` and `DUMPS_LISTEN` together.

```bash
DUMPS_URL=http://localhost:9000 DUMPS_LISTEN=127.0.0.1:9000 cli/dumps serve-local
```

Manage from the terminal:

```bash
cli/dumps list
```

```bash
cli/dumps archive <id>
```

## Thumbnails

`dumps publish` screenshots `index.html` at 1280x800 with headless Chrome, then resizes it to a 640px JPEG.

- Chrome lookup order: `DUMPS_CHROME`, then Google Chrome, Brave and Chromium in `/Applications` (macOS), then `google-chrome`, `chromium` and `chromium-browser` on `PATH`.
- Resize: `sips` on macOS, `magick` or `convert` (ImageMagick) on Linux.

Without a browser or a resizer, publishing still works. The CLI prints a warning and the home shows a placeholder with the first letter of the title. Open Graph tags are then added without an image.

## Notes

- `--auth none` accepts only loopback connections, and `dumpsd serve` refuses to listen on any other address.
- In local mode the public site and the admin API share one origin. Only publish HTML you trust. See [security.md](security.md).
- With the default `DUMPS_SERVICE=none`, no service is installed. Run `dumps serve-local` when you need it.
