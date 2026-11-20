# Linux host

Use this for a Raspberry Pi, a VPS, a NAS or any box with systemd and ssh. Your laptop runs the CLI, the server runs `dumpsd`, Caddy and a tunnel.

## 1. Server prerequisites

- An ssh login (key-based, the CLI uses `BatchMode`)
- `rsync`
- Caddy, from your distribution or from the Caddy docs
- systemd with user services

## 2. ssh alias on your laptop

In `~/.ssh/config`:

```text
Host dumps-box
  HostName box.example.com
  User you
```

## 3. CLI config

In `~/.config/dumps/config` on your laptop:

```text
DUMPS_HOST=dumps-box
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=systemd-user
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=yourteam.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD tag>
```

`DUMPS_TARGET` is the server architecture: `linux/arm64` (Raspberry Pi 4 and 5 on a 64-bit OS, most ARM VPS) or `linux/amd64` (x86 VPS and NAS). With no `DUMPS_ROOT`, the root is the ssh home directory, so dumps live in `~/dumps`.

For a service, `DUMPS_AUTH` is `access` or `tailscale`. `none` is refused. Values that go into the service (root, listen address, URL, admin origin, team, AUD and Tailscale logins) may only contain letters, digits and `: / . _ @ -`. For the auth modes and the Access setup, see [../auth.md](../auth.md) and [../exposure/cloudflare.md](../exposure/cloudflare.md).

## 4. Install dumpsd and the service

On your laptop (needs Go):

```bash
cli/dumps install-dumpsd
```

This cross-compiles `dumpsd`, uploads it to `~/bin/dumpsd`, and writes `~/.config/systemd/user/dumpsd.service` with `Restart=on-failure` and `NoNewPrivileges=true`, then runs `daemon-reload`, enables and restarts it. The unit runs `dumpsd serve` on `DUMPS_LISTEN` (default `127.0.0.1:8090`). It also creates `bin/`, `dumps/`, `dumps-meta/` and `dumps-staging/` in the root.

Check it on the server:

```bash
systemctl --user status dumpsd
```

## 5. Keep user services running after logout

User services stop when your last session ends unless lingering is on. To have them start at boot and survive logout, run on the server:

```bash
loginctl enable-linger $USER
```

## 6. Caddy

Copy `deploy/Caddyfile` to the server (for example `/etc/caddy/Caddyfile`) and replace `/home/you` with the absolute root. It defines two loopback listeners:

- `127.0.0.1:8081`: public, serves only `dumps/`
- `127.0.0.1:8080`: admin, proxies `/_dumps/api/*` to `dumpsd` on `:8090` and serves the listing from `admin/`

The file turns off `auto_https` and the Caddy admin API (`admin off`), because TLS lives in the tunnel. Reload Caddy.

```bash
sudo systemctl reload caddy
```

The user that runs Caddy needs read access to `dumps/` and `admin/`. If you installed Caddy as a system service, it runs as another user, so check the folder permissions in your home.

An alternative is to run Caddy as a user service next to `dumpsd`. Example `~/.config/systemd/user/caddy.service`:

```text
[Unit]
Description=Caddy (dumps)
After=network-online.target

[Service]
ExecStart=/usr/bin/caddy run --config %h/Caddyfile
Restart=on-failure

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now caddy
```

Adjust the binary path (`command -v caddy`). With lingering on, both start at boot.

## 7. Expose it to the internet

Expose both listeners with a tunnel: [../exposure/cloudflare.md](../exposure/cloudflare.md) or [../exposure/alternatives.md](../exposure/alternatives.md).

## 8. Publish

```bash
cli/dumps publish ./page.html
```

The home and the listing are written by the first publish. To create them earlier:

```bash
cli/dumps regen
```
