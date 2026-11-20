# Enterprise roadmap proposal: object storage and Google Workspace

**Status:** proposal. Neither an S3 backend nor native Google sign-in is implemented in dumps today.

This proposal extends the current self-hosted Enterprise workflow with durable object storage and
organization-controlled access. The two capabilities can be delivered independently.

## What works today

- The CLI publishes to a local filesystem or to a remote filesystem over SSH and rsync. `dumpsd`
  renders the home, listing and admin from filesystem-backed metadata under a single-writer lock.
- The artifact origin serves static HTML and is public unless the organization puts SSO, VPN or
  private routing in front of the complete origin.
- The admin can verify Cloudflare Access or Tailscale identity. `DUMPS_AUTH` does not protect the
  static artifact origin.
- An organization can use Google Workspace as an identity provider through Cloudflare Access
  today. This puts Google login and the Access policy at the network edge; dumps itself does not
  validate Google tokens. Apply Access separately to the artifact and admin origins.

See the [current Enterprise guide](enterprise.md) for configuration and the
[Google Workspace integration in Cloudflare Access](https://developers.cloudflare.com/cloudflare-one/integrations/identity-providers/google-workspace/).

## Proposal: S3-compatible artifact storage

Let organizations store artifact bodies and generated static assets in a private S3 bucket while
keeping `DUMPS_URL` as the stable address readers use. A CloudFront S3 origin can remain private
with Origin Access Control (OAC); another provider needs an equivalent private-origin mechanism.
The bucket must not be directly public.

The backend must preserve dumps behavior, including generated home and listing, hidden and archived
states, updates, removal, activation rollback and the single-writer guarantee. The current design
depends on local filesystem operations and `flock`; mounting an object bucket as a filesystem is
not a sufficient implementation plan. The storage contract and activation model need a threat
model and failure analysis before implementation.

Design requirements:

- Keep the filesystem backend as the default and preserve its golden output.
- Define object keys, metadata ownership, staging, activation and cleanup for interrupted publishes.
- Serialize metadata changes or use a proven conditional-write strategy so concurrent publishers
  cannot lose home or listing updates.
- Keep the bucket private behind the approved artifact origin. Apply SSO, VPN or private routing
  at the reader-facing boundary, not only at the admin.
- Prefer workload or role-based short-lived credentials with a bucket-prefix scope. Do not require
  long-lived access keys in agent configuration.
- Cover uploads, updates, archive, unarchive, delete, rollback, cache behavior and recovery from
  partial network failure.

Open design choices include AWS S3 first versus a wider S3-compatible contract, whether the CLI
uploads directly or sends a prepared bundle to a service, and how the metadata writer coordinates
activation. Amazon S3 supports conditional writes with `If-Match` and `If-None-Match`, which may be
one building block; that alone does not define dumps' complete transaction or lock semantics.
See [S3 conditional writes](https://docs.aws.amazon.com/AmazonS3/latest/userguide/conditional-writes.html)
and [CloudFront access to a private S3 origin](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html).

## Proposal: native Google Workspace access

The existing Cloudflare Access integration is the near-term Google login option. A native dumps
integration would remove the requirement to place that specific identity proxy in front of the
service and make access policy portable across self-hosted deployments.

Keep these permissions separate:

1. **Reader:** may open the protected home and artifact pages.
2. **Publisher:** may create or update artifacts through an approved publishing path.
3. **Admin:** may archive, rename or remove artifacts and change service settings.

The design should decide whether Google Workspace identity gates artifact reads, admin actions, or
both. It should support explicit organization-controlled domain policy and evaluate group policy as
a separate authorization requirement. Google authentication is not the same as authorization: a
successful Google login must not grant access by itself.

Security requirements:

- Verify signed identity tokens on the server, including signature, issuer, audience and expiry.
- When restricting access to Workspace domains, use Google's verified hosted-domain (`hd`) claim;
  do not treat an email suffix alone as proof of Workspace membership.
- Keep the artifact boundary in front of every page and ensure there is no direct bucket or origin
  route that bypasses authentication.
- Use secure, short-lived sessions with explicit sign-out, revocation and membership-change behavior.
- Deny by default, separate read and write roles, and avoid sending credentials or artifact contents
  to Google beyond the identity flow the organization approves.

Google's [backend authentication guide](https://developers.google.com/identity/sign-in/web/backend-auth)
describes server-side ID-token verification and the hosted-domain claim. If group-based authorization
is included, determine the required Workspace API permissions and administrator consent before
choosing an implementation.

## Suggested slices and acceptance criteria

1. Threat model and design review for bucket access, writer coordination, Google sessions, roles,
   cross-origin deployment and recovery. Decide the first supported storage provider and identity
   boundary.
2. Add a storage interface and S3 implementation while keeping the filesystem backend and its
   golden output unchanged.
3. Implement atomic or conditional activation, metadata serialization, rollback, deletion and
   recovery behavior for the object backend.
4. Add organization-managed credential configuration and least-privilege examples. Never put cloud
   credentials or organization hostnames in the public skill package.
5. Add Google Workspace authentication and reader, publisher and admin authorization at the chosen
   boundary, or explicitly keep Cloudflare Access as the supported identity gateway for the first
   S3 release.
6. Document and verify that anonymous readers, unauthorized domains or groups, and read-only users
   cannot access or modify protected artifacts. Test concurrent publishers and interrupted writes.

The S3 origin access layer and the user login layer solve different problems: keep the bucket
private from direct access, then enforce the organization's viewer policy at the public or private
reader-facing origin.
