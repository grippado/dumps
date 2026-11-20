# macOS host

Use this when a Mac (a Mac mini, an old laptop) is the always-on server and you publish from another machine. If you only want dumps on the Mac you are using, see [../local.md](../local.md) instead.

## 1. Mac prerequisites

- Remote Login on (System Settings, General, Sharing) so the CLI can ssh in
- Homebrew
- `rsync` and Caddy

Install them:

```bash
brew install caddy rsync
```

Stop the Mac from sleeping if it should serve around the clock (Energy settings).

## 2. CLI config on the machine you publish from

```text
DUMPS_HOST=my-mac
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=darwin/arm64
DUMPS_SERVICE=launchd
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=yourteam.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD tag>
```

`DUMPS_TARGET` is `darwin/arm64` for Apple Silicon or `darwin/amd64` for Intel. `my-mac` is an ssh alias in `~/.ssh/config`. With no `DUMPS_ROOT`, the root is the ssh home directory.

## 3. Install dumpsd and the launchd agent

```bash
cli/dumps install-dumpsd
```

This builds `dumpsd` for the Mac, uploads it, and writes `~/Library/LaunchAgents/io.github.grippado.dumpsd.plist` (`RunAtLoad` and `KeepAlive` on, log in `<root>/dumpsd.log`), then loads it with `launchctl bootstrap gui/<uid>`.

It is a LaunchAgent in the user's GUI domain, so it starts when that user logs in. For an unattended Mac, enable automatic login for that user.

Check it on the Mac:

```bash
launchctl print gui/$(id -u)/io.github.grippado.dumpsd
```

## 4. Caddy

Copy `deploy/Caddyfile` to the Mac and replace `/home/you` with the absolute root (for example `/Users/you`). Then run Caddy with that file.

```bash
caddy run --config /path/to/Caddyfile
```

To run Caddy at login, wrap that command in your own launchd agent.

## 5. Expose and publish

Expose the two loopback listeners with a tunnel: [../exposure/cloudflare.md](../exposure/cloudflare.md) or [../exposure/alternatives.md](../exposure/alternatives.md). Then:

```bash
cli/dumps publish ./page.html
```

## Notes

- Thumbnails are rendered on the publishing machine, not on the Mac server.
