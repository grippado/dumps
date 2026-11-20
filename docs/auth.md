# Admin authentication

Only the admin API needs a login. Published dumps are public. `dumpsd serve` checks every request under `/_dumps/api/` with one of three modes, chosen with `--auth` (or `DUMPS_AUTH` in the CLI config). The default is `access`.

The listing page is static HTML. Its protection is the login in front of the admin host (Access, Tailscale), not `dumpsd`. The public host has no API. `dumpsd serve` only listens on loopback and refuses to start on any other address.

For an installed service, `DUMPS_AUTH` is `access` or `tailscale`. The CLI refuses `DUMPS_AUTH=none` for a service.

The identity check runs first. A failure returns `403 acesso negado`, and the reason is written to the log.

## access (Cloudflare Access)

Config: `DUMPS_AUTH=access`, `DUMPS_ACCESS_TEAM`, `DUMPS_ACCESS_AUD`, `DUMPS_ADMIN_ORIGIN`. Guide: [exposure/cloudflare.md](exposure/cloudflare.md).

`dumpsd` reads the `Cf-Access-Jwt-Assertion` header and verifies it itself, so a request that did not pass through Access cannot forge it. It fails closed on any problem.

What is checked:

- The token is present and at most 16 KiB, with three dot-separated parts.
- The header algorithm is `RS256`. Anything else (including `none` and `HS256`) is rejected.
- The signature verifies against the RSA key with the token's `kid`.
- `iss` equals `https://<team>` exactly.
- `aud` contains the configured AUD tag (a string or a list).
- `exp` is present and not past, `nbf` (if present) is not in the future, both with 60 seconds of clock leeway.

The signing keys come from `https://<team>/cdn-cgi/access/certs` and are cached:

- Keys are refreshed after one hour, with at least a minute between attempts (10 seconds before the first successful fetch).
- If a refresh fails, the cached keys keep working, up to 24 hours after the last successful fetch. After that every request is denied.
- An unknown `kid` is denied.

The `email` claim is only used to label log lines. It is not an authorization check: who may log in is decided by your Access policy. See [exposure/cloudflare.md](exposure/cloudflare.md).

## tailscale

Config: `DUMPS_AUTH=tailscale`, and optionally `DUMPS_TAILSCALE_ALLOW` (logins separated by spaces, one `--allow` flag each).

Uses the identity that `tailscale serve` adds in the `Tailscale-User-Login` header.

- The TCP peer must be loopback. Anything else is denied.
- The header must be present and not empty.
- With one or more `--allow <login>` (`DUMPS_TAILSCALE_ALLOW`), the login must be in the list (case-insensitive). With none, any login that reaches the listener is accepted.

This mode trusts a plain header. It is safe only when `tailscale serve` is the only way to reach the admin listener. `dumpsd` accepts the header from any loopback connection, so do not put any other proxy or tunnel in front of it that could pass a client-supplied `Tailscale-User-Login` through, and never reach it through Funnel or a public hostname. See [exposure/alternatives.md](exposure/alternatives.md).

## none

Config: there is no key to set for a service. `dumps serve-local` starts `dumpsd` with `--auth none` by itself and does not read `DUMPS_AUTH`.

`none` accepts only direct loopback connections and sets no identity (`local` in the logs). It also refuses, with the same `403`, any request that carries one of these headers, which proxies and tunnels add:

- `X-Forwarded-For`
- `Forwarded`
- `Cf-Connecting-Ip`
- `X-Real-Ip`
- `Cf-Access-Jwt-Assertion`

The CLI refuses `DUMPS_AUTH=none` for an installed service. There is one use for this mode: local mode, where only you reach the port.

Do not put `none` behind a proxy or tunnel. The header check catches proxies that add those headers, but a proxy on the same machine that adds none of them still connects from loopback and would pass. On a shared machine, any local user can call the API directly.

`--static` (serving the site from `dumpsd` itself) is allowed only with `--auth none`, and it requires `--admin-listen` on a different loopback port. In that mode the admin listener also serves the listing page without a login, and only that listener has the API. See [local.md](local.md).

## Checks that apply in every mode

After authentication, in this order:

1. Method must be `POST` (otherwise `405`, with `Allow: POST`).
2. The `Origin` header must equal `--origin` (`DUMPS_ADMIN_ORIGIN`) exactly, character by character: scheme, host and port (otherwise `403`). In local mode the origin is `DUMPS_ADMIN_ORIGIN` or `http://localhost:<port of DUMPS_ADMIN_LISTEN>`, so open the admin through `localhost`, not `127.0.0.1`. This blocks cross-site requests from a browser that is logged in. A request with no `Origin` is rejected.
3. `Content-Type` must be `application/json` (otherwise `415`).
4. The body is at most 4 KiB and must be a JSON object with only the fields `id` and `title`. Unknown fields are rejected (`400`).
5. `id` must match `^[a-z0-9][a-z0-9-]{0,62}$` (otherwise `400`).
6. Actions are `archive`, `unarchive`, `title` and `delete`. Others return `404`.
7. For `title`, the value is trimmed and collapsed to single spaces, must be valid UTF-8 without control or invisible format characters, and must be 1 to 200 characters (otherwise `400`).

The `Origin` check and the JSON content type keep a third-party page from firing actions with your login (CSRF). A plain form `POST` does not pass.

Responses are sent with `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`.

## Fronting with something else

There is no built-in mode for other identity-aware proxies such as oauth2-proxy, an OIDC gateway or basic auth. Using one in front of the admin host does not make `dumpsd` verify it, and `--auth none` is not a safe way to combine them (see above). This is not supported natively yet. Use `access` or `tailscale`, or manage dumps from the CLI over ssh (`dumps archive`, `edit`, `rm`), which needs no admin host.

A safe way to add another provider would be a new mode in `dumpsd` that validates the identity inside the binary.
