# Android (Termux) host

This is where dumps started: an Android phone on the home wifi, serving through a tunnel. The phone runs `dumpsd`, Caddy and the tunnel as runit services. The CLI runs on your laptop.

## 1. Install Termux

Install Termux and Termux:Boot from F-Droid, not from the Play Store (that build is outdated). Then, in Termux:

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

`DUMPS_TARGET=linux/arm64` produces a static binary (`CGO_ENABLED=0`) that runs on Termux. Inside Termux there is no `/etc/resolv.conf` and no system CA bundle, so `dumpsd` reads DNS servers from `$PREFIX/etc/resolv.conf` (falling back to 1.1.1.1 and 8.8.8.8) and CA certificates from `$PREFIX/etc/tls/cert.pem`. It needs them to fetch the Cloudflare Access keys.

The root defaults to the Termux home, so dumps live in `~/dumps`.

## 4. Install dumpsd and its service

```bash
cli/dumps install-dumpsd
```

With `DUMPS_SERVICE=termux-runit` this creates `$PREFIX/var/service/dumpsd/` with a `run` script and a `log/run` script (logs through `svlogd` into `$LOGDIR/sv/dumpsd`). If the service was never enabled, the CLI tells you to enable it. On the phone:

```bash
sv-enable dumpsd
```

When the service is already enabled, `install-dumpsd` restarts it for you.

## 5. Caddy and the tunnel as services

Run Caddy (with `deploy/Caddyfile`, root replaced by the Termux home path) and `cloudflared` (with `deploy/cloudflared.yml`) as their own runit services, then enable them.

```bash
sv-enable caddy
```

```bash
sv-enable cloudflared
```

Each service needs a directory under `$PREFIX/var/service/<name>/` with an executable `run` script that `exec`s the program. Tunnel setup is in [../exposure/cloudflare.md](../exposure/cloudflare.md).

## 6. Using sv over ssh

`sv` finds services through `SVDIR`, which Termux sets only in interactive sessions. Over ssh, set it yourself:

```bash
SVDIR=$PREFIX/var/service sv status dumpsd
```

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
