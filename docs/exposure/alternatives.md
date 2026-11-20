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

`tailscale serve` adds the `Tailscale-User-Login` header that `dumpsd` trusts. Read the trust rules in [../auth.md](../auth.md) before using this mode. Funnel serves your `ts.net` name.

## ngrok

ngrok can publish `8081` on a public URL (custom domains depend on your plan). Its login features (OAuth, basic auth) sit in front of the admin host, but `dumpsd` has no mode that reads an ngrok identity, so the admin API would only be guarded by the upstream login and the Origin check. That is not supported natively, see "Fronting with something else" in [../auth.md](../auth.md). For admin work, prefer `dumps archive`, `edit` and `rm` over ssh, and expose only the public listener.

## frp

Run `frps` on a VPS with a public IP and `frpc` on the home box, forwarding `8081` to the VPS. The same caveat as ngrok applies to the admin host: expose only the public listener unless you add your own authenticated layer.

## Caddy on a VPS, with a tunnel to home

Run Caddy on a VPS for automatic HTTPS and a custom domain, and reach the home box through a WireGuard tunnel or an ssh reverse tunnel.

```bash
ssh -N -R 127.0.0.1:18081:127.0.0.1:8081 vps.example.com
```

On the VPS, Caddy reverse-proxies `dumps.example.com` to `127.0.0.1:18081`. Keep the admin listener off the VPS, or protect it with a mechanism `dumpsd` supports.

## Tradeoffs

| Option | Cost | Custom domain | Auth for the admin | Setup effort |
|--------|------|---------------|--------------------|--------------|
| Cloudflare Tunnel and Access | A domain on Cloudflare; check current plan limits | Yes | Native (`access`) | Medium: tunnel, DNS, Access app |
| Tailscale Funnel and serve | Check current plan limits | Uses the `ts.net` name | Native (`tailscale`) | Low |
| ngrok | Check current plan limits | Depends on the plan | None native | Low |
| frp on a VPS | A VPS | Yes | None native | Medium |
| Caddy on a VPS and WireGuard or ssh tunnel | A VPS | Yes, automatic HTTPS | None native | Higher: VPS, tunnel, keep-alive |

For a public site with no admin UI, any option works. For the browser admin, use Cloudflare Access or Tailscale, or manage dumps from the CLI.
