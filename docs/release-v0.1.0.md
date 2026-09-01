# v0.1.0 release evidence

This release packages the content-addressed proof-reuse protocol, its strict
`.gooo` contract, the immutable 57-lock fixture bundle, the deterministic
local fixture server, and the GitHub Actions conformance job.

The release is valid only when the published GitHub release reports
`immutable=true` and its tag and assets are inspected through the GitHub API.
The release asset is evidence packaging; it does not replace the Actions run
artifact that contains observed wall time and RSS.
