# Enterprise delivery mode

Enterprise is a delivery and audience overlay, not a page type or style level. Use it alongside
the specialist recipe and style dial when an organization wants `dump-this` to create artifacts
for an internal `dumps.<org-domain>` service.

## Keep the organization boundary

- Use only company material the requester is authorized to use and the intended audience is
  allowed to read. Include only the internal or customer details needed for the page and follow the
  organization's privacy policy. Keep secrets, passwords, tokens, private keys and session cookies
  out of the HTML, links, filenames, metadata and source code.
- Do not send internal material to a personal dumps host, a public home, or a third-party artifact
  service as a fallback.
- Do not ask the user to paste credentials into chat. Authentication and network access must
  already be provisioned for the agent's machine or harness.
- Keep third-party requests out of internal pages by default: use system or locally hosted fonts,
  inline assets, and no analytics unless the organization approves them. Keep `DUMPS_GA_ID` empty
  unless its data flow is approved.

## Resolve the internal destination before publishing

The CLI publishes through SSH and rsync using its active `DUMPS_HOST`; `DUMPS_URL` is the link it
prints for readers. This is not an HTTP upload API. The CLI must be installed and the agent must
run in an environment that can reach the approved server.

1. Run `dumps config` and check that `host` and `url` point to the organization's provisioned
   service. If IT provided a dedicated config file, use it explicitly for both commands:
   `DUMPS_CONFIG=/path/to/org/dumps.conf dumps config` and
   `DUMPS_CONFIG=/path/to/org/dumps.conf dumps publish ...`.
2. Confirm that the configured `DUMPS_URL` is the exact organization-owned artifact origin and
   that the organization's SSO or VPN policy protects that entire origin. Do not infer the target
   from the company name, repository, or a previous personal setup.
3. If the company profile is missing, the domain is ambiguous, or access protection is unknown,
   do not publish. Keep the HTML in the approved local workspace and ask for the approved internal
   destination or its managed CLI profile.

## Respect the access boundary

Published files are static and public by default. `--hidden` keeps a dump off the home page; it
does not add authentication. Anyone who can reach the artifact host and knows the link can read
it. Sensitive company artifacts require SSO or VPN enforcement in front of the whole
`DUMPS_URL` origin, including its home and every dump. Verify the organization has applied that
policy before publishing.

Dumps share one browser origin. JavaScript in one dump can read other dumps on that host, so SSO
or VPN does not isolate teams from each other. Publish only trusted HTML and restrict who can
publish. Use separate dumps instances and artifact origins for groups that should not share this
trust boundary.

The admin host is a separate origin. `DUMPS_AUTH=access` and `DUMPS_AUTH=tailscale` protect the
admin API when configured as documented; they do not protect the artifact host. Other identity
providers are not natively verified by `dumpsd`. Do not combine an unsupported admin proxy with
`DUMPS_AUTH=none`. If the organization cannot use a supported admin mode, manage the service over
SSH and keep the admin origin unexposed. See the project's [auth guide](https://github.com/grippado/dumps/blob/main/docs/auth.md),
[Cloudflare guide](https://github.com/grippado/dumps/blob/main/docs/exposure/cloudflare.md) and
[alternatives](https://github.com/grippado/dumps/blob/main/docs/exposure/alternatives.md).

## Publish and report

When the configured destination and audience are confirmed, publish for review:

```bash
dumps publish <page> --title "<title>" --hidden --source agent
```

Check that the returned URL uses the approved organization origin. Say that the link requires
the company's network or login, and that hidden means absent from the home page only. Do not
unarchive an internal dump unless the organization's home has the same access boundary.

In browser-only harnesses such as claude.ai or ChatGPT, the local CLI is unavailable unless the
organization has provided an approved connector. Do not claim to have published to the internal
service. If the harness cannot save to an approved internal workspace, ask the user to use an
approved local harness or provide an internal connector before generating company material there.
Use the platform's hosted artifact only if company policy explicitly allows that destination for
this material.
