# Admin authentication

Only the admin API needs a login. Published dumps are public. `dumpsd serve` checks every request under `/_dumps/api/` with one of three modes, chosen with `--auth` (or `DUMPS_AUTH` in the CLI config). The default is `access`.

The identity check runs first. A failure returns `403 acesso negado`, and the reason is written to the log.

## access (Cloudflare Access)

Needs `--team` (`DUMPS_ACCESS_TEAM`) and `--aud` (`DUMPS_ACCESS_AUD`). `dumpsd` reads the `Cf-Access-Jwt-Assertion` header and verifies it itself, so a request that did not pass through Access cannot forge it. It fails closed on any problem.

What is checked:

- The token is present and at most 16 KiB, with three dot-separated parts.
- The header algorithm is `RS256`. Anything else is rejected.
- The signature verifies against the key with the token's `kid`.
- `iss` equals `https://<team>` exactly.
- `aud` contains the configured AUD tag (a string or a list).
- `exp` is present and not past, `nbf` (if present) is not in the future, both with 60 seconds of clock leeway.

The signing keys come from `https://<team>/cdn-cgi/access/certs` and are cached:

- Keys are refreshed after one hour, with at least a minute between attempts (10 seconds before the first successful fetch).
- If a refresh fails, the cached keys keep working, up to 24 hours after the last successful fetch. After that every request is denied.
- An unknown `kid` is denied.

The `email` claim is only used to label log lines. It is not an authorization check: who may log in is decided by your Access policy. See [exposure/cloudflare.md](exposure/cloudflare.md).

## tailscale

Uses the identity that `tailscale serve` adds in the `Tailscale-User-Login` header.

- The TCP peer must be loopback. Anything else is denied.
- The header must be present.
- With one or more `--allow <login>` (`DUMPS_TAILSCALE_ALLOW`), the login must be in the list (case-insensitive). With none, any login that reaches the listener is accepted.

This mode trusts a plain header. It is safe only when `tailscale serve` is the way to reach the admin listener. Do not put any other proxy or tunnel in front of it that could pass a client-supplied `Tailscale-User-Login` through, and never reach it through Funnel or a public hostname.

## none

Accepts only loopback connections and sets no identity (`local` in the logs). It exists for local mode (`dumps serve-local`). Do not use it behind a proxy or tunnel: the proxy connects from loopback, so every request would pass.

`--static` (serving the site from `dumpsd` itself) is allowed only with `--auth none`.

## Checks that apply in every mode

After authentication, in this order:

1. Method must be `POST` (otherwise `405`).
2. The `Origin` header must equal `--origin` (`DUMPS_ADMIN_ORIGIN`) exactly (otherwise `403`). This blocks cross-site requests from a browser that is logged in. A request with no `Origin` is rejected.
3. `Content-Type` must be `application/json` (otherwise `415`).
4. The body is at most 4 KiB and must be a JSON object with only the fields `id` and `title`. Unknown fields are rejected (`400`).
5. `id` must match `^[a-z0-9][a-z0-9-]{0,62}$`.
6. Actions are `archive`, `unarchive`, `title` and `delete`. Others return `404`.
7. For `title`, the value is trimmed and collapsed to single spaces, must be valid UTF-8 without control or invisible format characters, and must be 1 to 200 characters.

Responses are sent with `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`. `dumpsd serve` refuses to listen on anything but loopback.

## Fronting with something else

There is no built-in mode for other identity-aware proxies such as oauth2-proxy, an OIDC gateway or basic auth. Using one in front of the admin host does not make `dumpsd` verify it, and `--auth none` is not a safe way to combine them (see above). This is not supported natively yet. Use `access` or `tailscale`, or manage dumps from the CLI over ssh (`dumps archive`, `edit`, `rm`), which needs no admin host.
