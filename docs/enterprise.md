# Enterprise and internal publishing

`dump-this` has an enterprise delivery mode for organizations that want artifacts at an internal
`dumps.<org-domain>` address. It is a security and delivery overlay: the requested specialist and
style still determine the page. The mode uses the organization's existing dumps server and CLI
configuration; it does not create an organization, identity provider, or publishing endpoint.

Ask for it explicitly, or say that the material must stay inside the company:

```text
/dump-this enterprise incident report for the security team, publish on our internal dumps host
```

## How publishing works

The `dumps` CLI prepares the HTML locally and sends it to the configured server over SSH and
rsync. `DUMPS_HOST` selects the SSH host and `DUMPS_URL` is the URL given to readers. The agent
machine must have the CLI, the organization's approved config and network access to the server.
See a [host guide](hosts/linux.md) for installation details.

Use a managed config on each publishing machine. For example:

```text
DUMPS_HOST=org-dumps
DUMPS_URL=https://dumps.example.org
DUMPS_LIST_URL=https://admin-dumps.example.org/_dumps/
DUMPS_TARGET=linux/amd64
DUMPS_SERVICE=systemd-user
DUMPS_ADMIN_ORIGIN=https://admin-dumps.example.org
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=org.cloudflareaccess.com
DUMPS_ACCESS_AUD=<admin Access application AUD>
```

`org-dumps` is an SSH alias configured for the approved server. Keep organization-specific
hostnames, Access values, keys and policies in managed local configuration, not in this public
repository. Select a dedicated config file with `DUMPS_CONFIG=/path/to/org/dumps.conf` when the
machine has more than one dumps destination. Check `dumps config` before publishing.

The server owner must configure the artifact and admin origins separately:

- `DUMPS_URL` serves the home and static HTML artifacts.
- `DUMPS_LIST_URL` serves the admin listing and API.
- `DUMPS_AUTH=access` makes `dumpsd` verify the Cloudflare Access JWT for the admin API. Its AUD
  must match the admin Access application.
- The SSH route used by the CLI must be reachable only through the organization's approved
  network or access policy.

## Protect the artifacts, not only the admin

Static dumps are public by default. `--hidden` keeps a dump off the home page but does not add
authentication; anyone who can reach the artifact host and knows the link can read it. For
company material, enforce corporate SSO or VPN access in front of the entire `DUMPS_URL` origin,
including the home and every artifact, before publishing. A private hostname without a public
route is another option when the organization has private DNS and a VPN.

The admin is a separate access boundary. `DUMPS_AUTH=access` and `DUMPS_AUTH=tailscale` are the
identity modes `dumpsd` can verify for the admin API. `DUMPS_AUTH=none` is only for local mode and
must not be placed behind a proxy. Other identity-aware proxies are not natively verified by
`dumpsd`; in that case, keep the admin origin unexposed and use the CLI over SSH. See the
[authentication guide](auth.md), [Cloudflare guide](exposure/cloudflare.md) and
[exposure alternatives](exposure/alternatives.md).

Each dump runs in the same browser origin as the other dumps. A page's JavaScript can read other
pages on that host, so SSO or VPN is not a boundary between teams. Publish only trusted HTML,
restrict publishing rights, and use separate instances and artifact origins for groups that must
not share that trust boundary. Keep `DUMPS_GA_ID` empty and avoid external fonts, scripts, analytics
and other third-party requests unless the organization's security policy approves them.

Before publishing, check that the active `DUMPS_URL` is the exact organization-approved origin
and that its access policy covers the artifact host itself. If the profile, domain or policy is
unknown, do not publish internal material. Keep the HTML in the approved workspace and ask the
service owner for the managed destination.

## Distribute the skill

For Claude Code and Codex, install `dump-this` alongside the CLI using
[`cli/dumps install-agent`](agent.md). The skill package includes the enterprise recipe. An
organization may distribute that package from an internal mirror and add a private policy
overlay, but should keep domains, credentials, customer data and identity-provider settings out
of the public repository.

claude.ai Skills and ChatGPT instructions can generate the HTML, but they cannot run the local
CLI or reach a private network by themselves. Direct publication from those hosted harnesses
requires an organization-approved connector. Without one, deliver the HTML only through a
destination the organization's policy approves; do not present a hosted artifact as an internal
dumps publication.
