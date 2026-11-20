# Android (Termux) host

This is where dumps started: an Android phone on the home wifi, serving through a tunnel. The phone runs `dumpsd`, Caddy and the tunnel as runit services. The CLI runs on your laptop. This guide uses the `termux-runit` service, which `dumps install-dumpsd` creates for you.

The Termux steps themselves (packages, runit, boot) follow the Termux ecosystem and are not part of this repository's code, so check the Termux documentation if something changes.

## 1. Install Termux

Install Termux and Termux:Boot from F-Droid, not from the Play Store (that build is outdated). Do not mix install sources between the two. Then, in Termux:

```bash
pkg install openssh rsync caddy cloudflared termux-services
```

Restart Termux once so `termux-services` sets up runit.

## 2. ssh into the phone

Start sshd in Termux (it listens on port 8022 by default) and add your laptop's public key to `~/.ssh/authorized_keys`.

```bash
sshd
```

```bash
passwd
```

On your laptop, in `~/.ssh/config`:

```text
Host phone
  HostName 192.168.0.50
  Port 8022
  User u0_a123
```

Use the real address and the user shown by `whoami` in Termux. Use a key, since the CLI runs ssh with `BatchMode=yes`.

## 3. CLI config

```text
DUMPS_HOST=phone
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=termux-runit
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=yourteam.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD tag>
```

The root defaults to the Termux home, so dumps live in `~/dumps`. To pin it, set `DUMPS_ROOT=/data/data/com.termux/files/home`.

For a service, `DUMPS_AUTH` is `access` or `tailscale`. `none` is refused. Values that go into the service may only contain letters, digits and `: / . _ @ -`.

`DUMPS_TARGET=linux/arm64` produces a static binary (`CGO_ENABLED=0`) that runs on Termux. Inside Termux there is no `/etc/resolv.conf` and no system CA bundle in the usual places, so to fetch the Cloudflare Access keys `dumpsd` reads:

- DNS: `$PREFIX/etc/resolv.conf` (or `/data/data/com.termux/files/usr/etc/resolv.conf` and `/etc/resolv.conf`). With no valid nameserver, it uses `1.1.1.1` and `8.8.8.8`.
- CA certificates: `$PREFIX/etc/tls/cert.pem` (or `/data/data/com.termux/files/usr/etc/tls/cert.pem`).

This only matters with `DUMPS_AUTH=access`.

## 4. Install dumpsd and its service

```bash
cli/dumps install-dumpsd
```

With `DUMPS_SERVICE=termux-runit` this creates `$PREFIX/var/service/dumpsd/` with a `run` script and a `log/run` script (logs through `svlogd` into `$LOGDIR/sv/dumpsd`). If the service was never enabled, the CLI prints `serviço dumpsd criado; para ligar: sv-enable dumpsd`. On the phone:

```bash
sv-enable dumpsd
```

When the service is already enabled, `install-dumpsd` restarts it for you.

## 5. Caddy and the tunnel as services

Run Caddy (with `deploy/Caddyfile`, `/home/you` replaced by `/data/data/com.termux/files/home`) and `cloudflared` (with `deploy/cloudflared.yml`) as their own runit services, then enable them. Each service needs a directory under `$PREFIX/var/service/<name>/` with an executable `run` script that `exec`s the program.

Example for Caddy, in `$PREFIX/var/service/caddy/run`:

```text
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec caddy run --config /data/data/com.termux/files/home/Caddyfile
```

Example for cloudflared, in `$PREFIX/var/service/cloudflared/run`:

```text
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec cloudflared tunnel run dumps
```

```bash
chmod 700 $PREFIX/var/service/caddy/run $PREFIX/var/service/cloudflared/run
```

```bash
sv-enable caddy
```

```bash
sv-enable cloudflared
```

Tunnel setup is in [../exposure/cloudflare.md](../exposure/cloudflare.md).

## 6. Using sv over ssh

`sv` finds services through `SVDIR`, which Termux sets only in interactive sessions. Over ssh, set it yourself:

```bash
SVDIR=$PREFIX/var/service sv status dumpsd
```

`dumps install-dumpsd` already sets `SVDIR` when it restarts the service.

## 7. Keep the phone alive

Android kills background apps aggressively. Do all of these:

- Hold a wake lock from inside Termux:

```bash
termux-wake-lock
```

- Turn battery optimization off for Termux (and Termux:Boot) in Android settings, and allow it to run in the background.
- Put a script in `~/.termux/boot/` (Termux:Boot runs it after reboot) that starts `sshd`, acquires the wake lock and brings the runit services up (see the `termux-services` documentation for the command).
- Keep the phone plugged in, on wifi, and avoid battery saver modes.

## 8. Publish

```bash
cli/dumps publish ./page.html
```

Thumbnails are made on your laptop, so the phone only stores and serves files.
