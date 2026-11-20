# Security

dumps is built for a person or a small group publishing their own HTML. This page describes what it defends, what it does not, and what you must do yourself. To report a problem, see [../SECURITY.md](../SECURITY.md).

## Model

- **The public host serves only `dumps/`.** Manifests (`dumps-meta/`), staging, `bin/` and the listing live outside the public docroot. The Caddy example hides dotfiles and redirects `index.html` and `.html` URLs to clean ones. In local mode, `dumpsd --static` applies the same rules: dotfile segments return 404, `/_dumps` returns 404 on the dumps origin, and only `dumps/` is served there.
- **The admin host is behind a login.** The listing page and `/_dumps/api/*` should only be reachable through Cloudflare Access or Tailscale. `dumpsd` also verifies the identity itself (JWT or Tailscale header), the `Origin` and the request shape. See [auth.md](auth.md).
- **`dumpsd serve` listens on loopback only.** It refuses any other `--listen` or `--admin-listen` address. Nothing opens an inbound port on the internet; a tunnel does the exposure.
- **One writer.** Home, listing and manifests are written only by `dumpsd`, under a file lock, using write-then-rename. Publishing uploads to staging and swaps in atomically with rollback.
- **Local mode has two origins.** The dumps are served on one port and the admin on another. The API accepts only the admin origin, so a dump's JavaScript cannot call it. See [local.md](local.md).

## Path and id rules

- Ids match `^[a-z0-9][a-z0-9-]{0,62}$`, checked in the CLI, in `dumpsd activate` and `dumpsd apply`, and in the API. This keeps ids from containing `/`, `..` or leading dots, and `_` (the folder of shared assets) cannot be an id.
- The CLI refuses sources in or containing `DUMPS_BLOCKED_PATHS` (exit 2). Use it for directories that must never be published, such as work or client folders. Each item expands `~` on its own.
- For folders, symlinks must resolve inside the folder and not into a blocked path. Content is copied with symlinks resolved, and `.git/` and `.DS_Store` are excluded. A loose file must be `.html` or `.htm`.
- `listing_dir` in `site.json` cannot be absolute, be `.`, or start with `..`. Its first component cannot be `dumps`, `dumps-meta`, `dumps-staging` or `bin` (compared without case), and cannot start with a dot.
- `DUMPS_ROOT` cannot contain a single quote. The values written into a service (the resolved root, listen address, URL, admin origin, Access team and AUD, and each Tailscale login) may only contain letters, digits and `: / . _ @ -`, because they pass through a shell.

## What an attacker can and cannot do

| Attacker | Can | Cannot |
|----------|-----|--------|
| Anyone on the internet | Read published dumps and the home. Guess ids (random 10-character ids are not guessable in practice, but a link is not a secret once shared) | Read manifests, staging or the listing, call the admin API, write files |
| A page visited by an admin (CSRF) | Send a cross-site request to the admin host | Pass the `Origin` check or carry a valid Access token |
| Someone who reaches the admin API without a login | Nothing: `403` | Archive, rename or delete |
| Someone who forges a JWT | Nothing | Sign with the Cloudflare key. `alg: none` and `HS256` are rejected |
| A token from another Access app of the same team | Nothing | Pass the `aud` check |
| An expired token, or one from another team | Nothing | Pass the `exp` and `iss` checks |
| A logged-in user who is allowed by your policy | Archive, unarchive, rename, delete any dump | Publish or change content (that needs ssh to the server), run commands |
| JavaScript inside a dump, in local mode | Read the other dumps on the dumps origin | Call the admin API (its `Origin` is not the admin origin, and the dumps listener has no API) |
| An ssh user with your key | Everything | |

Archived dumps are hidden from the home but still public to anyone with the link. Use it for tidiness, not for privacy. Do not publish anything you would not want read. The same applies to dumps published hidden by the agent skill: the link is live as soon as the publish ends. To take a dump offline, use `dumps rm`.

## Third-party HTML shares your origin

Every dump is served from the same origin, for example `https://dumps.example.com`. Scripts in one dump can read the cookies and `localStorage` of every other dump on that host (the home keeps the view preference in `localStorage`, for example), and can fetch their files. The injected bar also sets a small `dumps_tip` cookie there. A malicious dump can also pass as one of your pages (the bar and footer are injected HTML, not a security boundary), and a bad reputation can reach the whole domain.

The admin lives on another hostname and its API requires an exact `Origin`, so a malicious dump cannot trigger the admin by itself. Still, if the admin hostname is a subdomain of the same registrable domain as the dumps, cookies set on the parent domain can interfere.

Publish only HTML you wrote or trust. For content you do not trust (generated by others, from an unknown source), use a separate registrable domain, and a separate `dumps` instance with its own root. The same care applies to pages generated by an agent.

In local mode, the dumps and the admin are on different origins (different ports), and the API accepts only the admin origin. A dump's JavaScript cannot archive, rename or delete other dumps. Two things remain:

- Dumps still share the dumps origin with each other, as above.
- `--auth none` trusts any direct loopback request. It refuses requests that carry `X-Forwarded-For`, `Forwarded`, `Cf-Connecting-Ip`, `X-Real-Ip` or `Cf-Access-Jwt-Assertion`, but a local process that is not a browser can set `Origin` itself and call the API. Treat local mode as single-user.

## Trusted HTML in site.json

`tagline_html`, `host_card_html` and `footer_html` in `site.json` are inserted into the home as raw HTML ([configuration.md](configuration.md)). Treat `site.json` as code: only the server owner should be able to edit it.

## Secrets and files that must never be committed

- Tunnel credentials (`~/.cloudflared/<TUNNEL-UUID>.json`) and the `cert.pem` from `cloudflared tunnel login`.
- ssh private keys.
- Your real CLI config (`~/.config/dumps/config`) and any `site.json` with private text. The Access AUD tag is not a secret, since the JWT still needs a valid signature. It does identify your setup, together with the team domain and the admin hostnames, so keep the config private too.

Commit `examples/` only, with `example.com` hostnames.

## Privacy notes

- With `DUMPS_GA_ID` set, every dump and the home load `googletagmanager.com` and report page views and the `share` and `dumps_clean_mode` events. Leave it empty to disable it. The GA4 ID (`G-...`) appears in clear in the served HTML.
- Open Graph tags include the dump URL and, when present, the author names.
- The share buttons open WhatsApp and Telegram share links on click. Nothing is sent before that.

## Good practices

- Create the Access application before the DNS record of the admin.
- Never point the public hostname at the admin listener.
- Run `curl -i` against the admin host without a login and check that it does not return the listing.
- Reinstall `dumpsd` after you update the repository (`dumps install-dumpsd`). The CLI warns when the server version is behind.

## Reporting a vulnerability

Follow [../SECURITY.md](../SECURITY.md). Do not open a public issue for security problems.
