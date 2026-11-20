# dumps

Self-hosted place to publish static HTML artifacts. Publish a single `.html` file or a folder with an `index.html`, and get a link at `https://<host>/<id>/` that you can share.

It was born on an Android phone in a living room (Termux, Caddy and a Cloudflare Tunnel), but it runs on any Linux, macOS or Android box, or fully locally on your laptop.

## Features

- Publish a `.html` file or a folder with `index.html`; each dump gets a short random id (or one you choose with `--id`).
- Public home page with cards and thumbnails, a grid/list toggle and optional featured links.
- Top bar on every dump with a share button (WhatsApp, Telegram, copy link, native share on touch devices) and a clean mode (press Esc) that hides the bar and footer.
- Footer with the dump URL and authors.
- Open Graph and Twitter card tags injected into every published HTML file, with a thumbnail as preview image.
- Optional GA4 (`DUMPS_GA_ID`).
- Private admin listing behind a login: archive, unarchive, rename and delete.
- Atomic publish with rollback: a failed publish leaves the previous version online.
- Thumbnails rendered locally with headless Chrome or Chromium.
- One small Go binary on the server (`dumpsd`, no CGO), one bash CLI on your machine.

The UI strings (home, bar, footer, admin and CLI messages) are currently in Portuguese. i18n is planned.

## Architecture

```text
laptop                         server (any Linux, macOS or Android box)
------                         ----------------------------------------
cli/dumps  --- ssh/rsync --->  DUMPS_ROOT/
 - renders thumbnails            dumps/          public files, one folder per id
 - ships raw HTML                dumps-meta/     one JSON manifest per dump, .site.json
                                 dumps-staging/  uploads waiting for activation
                                 bin/dumpsd      prepare, activate, render, apply, serve
                                 admin/          private listing (listing_dir)

                                 Caddy :8081 (public)  serves dumps/ only
                                 Caddy :8080 (admin)   listing + proxy to dumpsd :8090
                                          |
                                 tunnel (cloudflared, Tailscale, ...)
                                          |
                                      internet
                                 public host: dumps.example.com
                                 admin host:  admin.example.com (behind login)
```

In local mode (`DUMPS_HOST=local`) there is no ssh, no Caddy and no tunnel: `dumps serve-local` makes `dumpsd` serve everything on loopback. Details in [docs/architecture.md](docs/architecture.md).

## Quick start (local mode)

Local mode publishes and serves on your own machine. You need bash, jq, rsync, python3 and Go (the version in `server/go.mod`). Headless Chrome or Chromium is optional and only used for thumbnails.

Clone the repo:

```bash
git clone https://github.com/grippado/dumps.git
```

```bash
cd dumps
```

Build and install `dumpsd` into the local root (`~/.local/share/dumps`). There is no need to run `make build`.

```bash
cli/dumps install-dumpsd
```

Publish a page. The command prints its URL.

```bash
cli/dumps publish ./page.html
```

Start the local server:

```bash
cli/dumps serve-local
```

Open http://localhost:8090 for the home page and http://localhost:8090/_dumps/ for the admin listing.

More in [docs/local.md](docs/local.md).

## Documentation

- [Architecture](docs/architecture.md): components, directory layout, publish flow, exit codes
- [Configuration](docs/configuration.md): every `DUMPS_*` key, `site.json`, `dumpsd` subcommands
- [Run it locally](docs/local.md)
- Hosts: [Linux](docs/hosts/linux.md), [macOS](docs/hosts/macos.md), [Android (Termux)](docs/hosts/android-termux.md)
- Exposure: [Cloudflare Tunnel and Access](docs/exposure/cloudflare.md), [alternatives](docs/exposure/alternatives.md)
- [Admin authentication](docs/auth.md)
- [Security model](docs/security.md)
- [Contributing](CONTRIBUTING.md) and [security policy](SECURITY.md)

Portuguese version: [README.pt-BR.md](README.pt-BR.md).

## License

MIT, see [LICENSE](LICENSE).
