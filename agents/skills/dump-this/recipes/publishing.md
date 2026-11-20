# Choose a publisher

`dump-this` creates portable HTML. It can deliver that HTML to a publisher the current harness
actually exposes; it does not assume every harness can log in to every service.

## Select the route

Use an explicitly requested publisher when it is available and allowed by the content's access
policy. Otherwise read `DUMP_THIS_PUBLISHER` from the active harness profile:

| Value | Route |
|---|---|
| `dumps` | Use the configured `dumps` CLI after checking `dumps config` and confirming the returned `DUMPS_URL` is the intended destination. |
| `native` | Use the publisher native to the active, signed-in harness: Claude Artifacts in claude.ai, ChatGPT Sites when the Sites integration is available, or another publisher exposed by that harness. Never fall back to the dumps CLI. |
| `file` | Deliver the HTML in the approved local workspace; do not publish it. |
| unset or `auto` | Use a native publisher exposed in the current harness. Otherwise use a verified dumps CLI configuration. If neither is available, deliver the HTML file. |

The account profile may choose a default route. For example, a personal profile can set
`DUMP_THIS_PUBLISHER=dumps`, while a work profile can set `DUMP_THIS_PUBLISHER=native`. The public
skill does not infer an account from a username, host name, or subscription label. It uses only
the active profile and authenticated session. Do not switch accounts or request credentials.

If the selected publisher is missing or does not expose a publish operation, keep the result as
HTML or a native preview and say that it is not published yet. Never silently route to a different
publisher because it is available.

## Native publisher behavior

- **Claude Artifacts:** In claude.ai, create the page as a native Artifact. Use the active Claude
  account. If sharing is controlled by a UI action the harness does not expose, leave the artifact
  ready for review and direct the user to Claude's Share or Publish controls. Claude for Work
  (Team and Enterprise) artifacts are shared inside the organization and are not public.
- **ChatGPT Sites:** When the active ChatGPT harness exposes Sites, create and deploy the page in
  the current account or workspace. A new Site is private by default. Keep its audience private
  unless the user requested a wider audience and the active account permits it. In Enterprise,
  public publishing is off by default and requires workspace permission.
- **Other publishers:** Use only a native tool or connector available in the current harness and
  authorized for the current account. If no such integration is exposed, return the portable HTML.

Native publisher URLs belong to the account or workspace that is active when the tool runs. Report
the URL and audience the platform confirms; do not describe a preview as published.

For internal material, also follow [`enterprise.md`](enterprise.md). A publisher being native to
Claude or ChatGPT is not enough by itself: confirm the active organization account and ensure the
platform's audience matches the intended readers.

Platform references:

- [Claude Artifacts: publishing and sharing](https://support.claude.com/en/articles/9547008-discover-publish-customize-and-share-artifacts)
- [Creating and using ChatGPT Sites](https://help.openai.com/en/articles/20001339-creating-and-using-chatgpt-sites)
- [Managing ChatGPT Sites for a workspace](https://help.openai.com/en/articles/20001338-managing-chatgpt-sites-for-your-workspace)
