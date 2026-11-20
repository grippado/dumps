# dumps

Self-hosted place to publish static HTML artifacts. Publish a single `.html` file or a folder with an `index.html`, and get a link at `https://<host>/<id>/` that you can share.

It was born on an Android phone in a living room (Termux, Caddy and a Cloudflare Tunnel), but it runs on any Linux, macOS or Android box, or fully locally on your laptop.

English | [Português](README.pt-BR.md)

Project page: https://grippado.github.io/dumps/

## Just want better artifacts? Use the skill

`dump-this` is a skill that makes a page the way a specialist would (an analyst for reports, an
editor for stories, a presentation coach for slides, a data visualization specialist for
dashboards, and more). It works in your favorite harness, with or without a dumps server:

- **Claude Code:** unpack [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) into
  `~/.claude/skills` (no clone needed), then type `/dump-this <what you want>`. Codex: same zip
  into `~/.codex/skills`. Commands in [docs/agent.md](docs/agent.md).
- **claude.ai:** upload [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) in
  Settings, Capabilities, Skills, and get the page as an artifact.
- **ChatGPT:** paste [dump-this-chatgpt.txt](https://grippado.github.io/dumps/dump-this-chatgpt.txt)
  into a custom GPT or a Project, and get the page in canvas.

Its own page, with the eight specialists: [grippado.github.io/dumps/skill](https://grippado.github.io/dumps/skill/).
It is opinionated on purpose: a style dial read from your own words goes from "conservative"
to "master innovation", and it refuses the default artifact look. See one example per type at
[grippado.github.io/dumps/examples](https://grippado.github.io/dumps/examples/).

Full guide: [docs/agent.md](docs/agent.md).
For organization-owned internal hosts, see the [enterprise publishing guide](docs/enterprise.md).

## Features

- Publish a `.html` file or a folder with `index.html`; each dump gets a short random id (or one you choose with `--id`).
- Public home page with cards and thumbnails, a grid/list toggle and optional featured links.
- Top bar on every dump with a share button (WhatsApp, Telegram, copy link, native share on touch devices) and a clean mode (press Esc) that hides the bar and footer.
- Footer with the dump URL and authors.
- Open Graph and Twitter card tags injected into every published HTML file, with a thumbnail as preview image.
- Optional GA4 (`DUMPS_GA_ID`).
- Private admin listing behind a login (Cloudflare Access or Tailscale): archive, unarchive, rename and delete.
- Atomic publish with rollback: a failed publish leaves the previous version online.
- Thumbnails rendered locally with headless Chrome or Chromium.
- An agent skill, `dump-this`, that lets Claude Code and Codex turn a request into a page and publish it hidden for your review.
- An enterprise delivery recipe for organization-managed CLI profiles and internal `dumps.<org-domain>` hosts protected by corporate SSO or VPN.
- Local mode with no server, no tunnel and no login, with the dumps and the admin on separate origins.
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

In local mode (`DUMPS_HOST=local`) there is no ssh, no Caddy and no tunnel: `dumps serve-local` makes `dumpsd` serve the dumps on `127.0.0.1:8090` and the admin on `127.0.0.1:8091`, two origins on loopback. Details in [docs/architecture.md](docs/architecture.md).

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

Open two URLs:

- http://localhost:8090 for the home and the dumps
- http://localhost:8091/_dumps/ for the admin listing (archive, rename, delete)

Use `localhost`, not `127.0.0.1`: the admin API accepts only the exact origin `http://localhost:8091`.

Optional, still inside the repo folder: install the agent skill and the `/dump-this` command for Claude Code and Codex.

```bash
cli/dumps install-agent all
```

The agent runs the `dumps` command, so put `cli/dumps` on your `PATH`. More in [docs/local.md](docs/local.md) and [docs/agent.md](docs/agent.md).

## Documentation

- [Architecture](docs/architecture.md): components, directory layout, publish flow, exit codes
- [Configuration](docs/configuration.md): every `DUMPS_*` key, `site.json`, `dumpsd` subcommands
- [Run it locally](docs/local.md)
- [Agent skill](docs/agent.md): `dump-this`, its recipes and the review flow
- [Enterprise publishing](docs/enterprise.md): internal domains, access boundaries and skill distribution
- [Enterprise roadmap proposal](docs/enterprise-roadmap.md): proposed S3-compatible storage and Google Workspace access
- Hosts: [Linux](docs/hosts/linux.md), [macOS](docs/hosts/macos.md), [Android (Termux)](docs/hosts/android-termux.md)
- Exposure: [Cloudflare Tunnel and Access](docs/exposure/cloudflare.md), [alternatives](docs/exposure/alternatives.md)
- [Admin authentication](docs/auth.md)
- [Security model](docs/security.md)
- [Contributing](CONTRIBUTING.md) and [security policy](SECURITY.md)

## Development

```bash
make check
```

`make check` runs gofmt, `go vet`, shellcheck and the Go tests, the same as CI.

## Built with flux

dumps was built end to end with [flux](https://grippado.github.io/flux/), a family of commands
for AI agents: every slice was refined, built, reviewed and iterated with it, and the story of
the build is itself a dump at [dumps.gripp.link/historia](https://dumps.gripp.link/historia/).

## License

MIT, see [LICENSE](LICENSE).
