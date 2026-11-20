# Alternatives to Cloudflare

dumps only needs two things from an exposure layer: the public host reaches Caddy `127.0.0.1:8081`, and the admin host (if you want one) reaches Caddy `127.0.0.1:8080` behind a login that `dumpsd` can verify. Commands and limits of these third-party tools change, so treat the snippets as a starting point and check their docs.

## Tailscale

Tailscale Funnel publishes a local port on a public `*.ts.net` hostname. `tailscale serve` shares a port only inside your tailnet.

- Public: Funnel the public listener (`8081`).
- Admin: `tailscale serve` the admin listener (`8080`), so only tailnet devices can open it.

```bash
tailscale funnel --bg 8081
```

```bash
tailscale serve --bg --https=8443 8080
```

Funnel and serve take different ports on the same hostname, and the exact flags vary by Tailscale version. Set the CLI config to match what you ended up with:

```text
DUMPS_URL=https://box.tailnet-name.ts.net
DUMPS_ADMIN_ORIGIN=https://box.tailnet-name.ts.net:8443
DUMPS_AUTH=tailscale
DUMPS_TAILSCALE_ALLOW="ada@example.com"
```

`tailscale serve` adds the `Tailscale-User-Login` header that `dumpsd` trusts, and `dumpsd` accepts it only from loopback. Funnel serves your `ts.net` name. Limits of the `tailscale` mode ([../auth.md](../auth.md)):

- `dumpsd` trusts the header from any loopback connection, so nothing except `tailscale serve` may be able to reach the admin listener. Never reach it through Funnel, a public hostname or another proxy or tunnel.
- `DUMPS_ADMIN_ORIGIN` must be the exact origin the browser sends, port included.
- Use `DUMPS_TAILSCALE_ALLOW` if the tailnet has other people.

For internal artifacts, do not use Funnel for the dumps listener. Route it through `tailscale serve`
or a private reverse proxy reachable only from the tailnet, and set `DUMPS_URL` to that private
origin. `DUMPS_AUTH=tailscale` protects the admin API; the private route must separately protect
the static artifact host.

## ngrok

ngrok can publish `8081` on a public URL (custom domains depend on your plan). Its login features (OAuth, basic auth) sit in front of the admin host, but `dumpsd` has no mode that reads an ngrok identity, so the admin API would only be guarded by the upstream login and the Origin check. That is not supported natively, see "Fronting with something else" in [../auth.md](../auth.md). For admin work, prefer `dumps archive`, `edit` and `rm` over ssh, and expose only the public listener.

## frp

Run `frps` on a VPS with a public IP and `frpc` on the home box, forwarding `8081` to the VPS. The same caveat as ngrok applies to the admin host: `dumpsd` does not understand the frp login, so expose only the public listener unless you add your own authenticated layer.

## Caddy on a VPS, with a tunnel to home

Run Caddy on a VPS for automatic HTTPS and a custom domain, and reach the home box through a WireGuard tunnel or an ssh reverse tunnel.

```bash
ssh -N -R 127.0.0.1:18081:127.0.0.1:8081 vps.example.com
```

On the VPS, Caddy reverse-proxies `dumps.example.com` to `127.0.0.1:18081`. Keep the admin listener off the VPS, or protect it with a mechanism `dumpsd` supports. You use any domain you like, with no third party in the traffic path. It gives the most control and the most work.

## Admin from the CLI over ssh

If you do not want an admin host at all, do not expose the admin listener. `dumps archive`, `unarchive`, `edit` and `rm` run `dumpsd apply` on the server over ssh, so anyone who already has ssh needs no login layer. Expose only the public listener. Do not use `DUMPS_AUTH=none` for this: the CLI refuses it for a service.

## Tradeoffs

| Option | Cost | Custom domain | Auth for the admin | Setup effort |
|--------|------|---------------|--------------------|--------------|
| Cloudflare Tunnel and Access | A domain on Cloudflare; check current plan limits | Yes | Native (`access`) | Medium: tunnel, DNS, Access app |
| Tailscale Funnel and serve | Check current plan limits | Uses the `ts.net` name | Native (`tailscale`) | Low |
| ngrok | Check current plan limits | Depends on the plan | None native | Low |
| frp on a VPS | A VPS | Yes | None native | Medium |
| Caddy on a VPS and WireGuard or ssh tunnel | A VPS | Yes, automatic HTTPS | None native | Higher: VPS, tunnel, keep-alive |

For a public site with no admin UI, any option works. For the browser admin, use Cloudflare Access or Tailscale, or manage dumps from the CLI.

## Another login provider in front

There is no native support for oauth2-proxy and similar. See [../auth.md](../auth.md) for what that means and the risk.
