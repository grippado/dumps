# Cloudflare Tunnel and Access

This setup exposes the server with no open inbound port: a tunnel carries two hostnames to two loopback listeners, and Cloudflare Access protects the admin one. Dashboard menu names change over time, so check the Cloudflare docs if a label differs.

You need a domain on Cloudflare and the Caddy setup from your host guide ([Linux](../hosts/linux.md), [macOS](../hosts/macos.md), [Android](../hosts/android-termux.md)).

## Two hostnames, two listeners

| Hostname | Goes to | Who can reach it |
|----------|---------|------------------|
| `dumps.example.com` (public) | Caddy `127.0.0.1:8081`, serves only `dumps/` | Anyone |
| `admin.example.com` (private) | Caddy `127.0.0.1:8080`, listing and admin API | Only users allowed by Access |

Never point the public hostname at the admin listener (`:8080`). The listing page is a static file with no login of its own, so anyone could read your ids, titles, sizes and machine names, including archived ones. Only the API actions are checked by `dumpsd`.

## 1. Create the tunnel

On the server, after installing `cloudflared`:

```bash
cloudflared tunnel login
```

```bash
cloudflared tunnel create dumps
```

This prints the tunnel UUID and writes a credentials file `~/.cloudflared/<TUNNEL-UUID>.json`. Treat it as a secret ([../security.md](../security.md)).

## 2. Ingress

Copy `deploy/cloudflared.yml` to `~/.cloudflared/config.yml` and fill in the tunnel UUID and your user:

```yaml
tunnel: <TUNNEL-UUID>
credentials-file: /home/you/.cloudflared/<TUNNEL-UUID>.json
ingress:
  - hostname: dumps.example.com
    service: http://127.0.0.1:8081
  - hostname: admin.example.com
    service: http://127.0.0.1:8080
  - service: http_status:404
```

The last rule returns 404 for any other hostname.

## 3. Create the Access application first

Create the Access application for `admin.example.com` before its DNS record exists. If the hostname resolves first, the admin listener is briefly public.

1. In the Cloudflare Zero Trust dashboard, go to Access, Applications, and add a self-hosted application.
2. Set the application domain to `admin.example.com`.
3. Add a policy with action Allow and an include rule for yourself (for example your email address, or an identity provider group). Everyone else is denied by default.

Do not create an Access app for `dumps.example.com`. It stays public.

## 4. Find the team domain and the AUD tag

- Team domain: in Zero Trust settings, General. It looks like `yourteam.cloudflareaccess.com`. This is `DUMPS_ACCESS_TEAM`, without `https://`.
- Application Audience (AUD) tag: open the application you created and copy the AUD tag from its overview (or its token settings). This is `DUMPS_ACCESS_AUD`.

The AUD tag is not a secret, but it identifies your setup, so keep the config out of public repos.

## 5. DNS records

Now route both hostnames to the tunnel.

```bash
cloudflared tunnel route dns dumps dumps.example.com
```

```bash
cloudflared tunnel route dns dumps admin.example.com
```

Run the tunnel.

```bash
cloudflared tunnel run dumps
```

On a long-lived box, run it as a service (systemd, runit or launchd; see the [host guides](../hosts/linux.md)).

## 6. Tell dumps about Access

In the CLI config:

```text
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=yourteam.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD tag>
DUMPS_ADMIN_ORIGIN=https://admin.example.com
```

`DUMPS_ADMIN_ORIGIN` must match the browser `Origin` exactly: scheme and host, with no trailing slash (add the port only if it is not the default).

Reinstall so the service picks up the arguments:

```bash
cli/dumps install-dumpsd
```

Open `https://admin.example.com/_dumps/`, log in through Access, and try archiving a dump. What `dumpsd` verifies is described in [../auth.md](../auth.md).

## Option: ssh to the server through Access

If you do not want to open ssh to the internet, you can put ssh behind Access too. This is optional and not needed when the CLI runs on the same network. It uses a third hostname, which is not part of `deploy/cloudflared.yml`.

1. Add an ingress rule before the 404 rule, for example `hostname: ssh.example.com` with `service: ssh://localhost:22`, and a DNS route for it.
2. Create an Access application for `ssh.example.com` with its own policy.
3. On your laptop, in `~/.ssh/config`:

```text
Host dumps-box
  HostName ssh.example.com
  User you
  ProxyCommand cloudflared access ssh --hostname %h
```

Log in once with `cloudflared access login https://ssh.example.com` so that non-interactive runs work. The CLI calls ssh with `BatchMode=yes`, so it cannot answer prompts. If the login has expired, `dumps` exits with code 3 ("server unreachable").

## Verify

```bash
curl -i https://admin.example.com/_dumps/
```

Without a login, the response must be a redirect to the Access login, never the listing.
