# gooo-content-addressed-proof-reuse

This repository defines a fail-closed, content-addressed proof-reuse protocol
for release-lock verification. Go has four deliberately narrow roles:

1. verify the parent release and receipt;
2. plan exact content-addressed reuse;
3. verify the current delta;
4. emit canonical machine evidence and a human report.

The `.gooo` contract declares the immutable parent release and asset
coordinates, parent semantic root, current lock set, lock identity and digest,
invalidation edges, cache authority, hash algorithm, replay obligations,
canonical order, retry, and timeout. Reuse is authorized only when the actual
parent release is immutable, the annotated tag target and source artifact match
exactly, the parent asset digest and receipt root verify, each current lock
coordinate+digest matches, and dependencies are unaffected.

The deterministic local fixture server contains 57 ordered locks. The baseline
verifier fetches every current lock. The candidate reuses only the verified
parent evidence and fetches the delta. Therefore the canonical 57-lock pair is
`reused=57, selected=0, requests=0` for an exact match, and a one-lock change
is `reused=56, selected=1, requests=1`.

Parent receipt missing, stale, or ambiguous is `UNKNOWN` with the exact six
fields `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
`blocked_by`. Digest mismatch, changed dependency, and authority escalation
are `REFUTED`. A full fallback may still produce
`full_verification_status=CLOSED`, but it preserves the original
`reuse_eligibility` and `reuse_claim`, emits `execution_mode=FULL_FALLBACK`,
`reused=0`, the full current `selected` count, and the fallback frontier.
Fallback cases are excluded from improvement evidence.

The fixed denominator is nine cases: three `CLOSED`, three `UNKNOWN`, and
three `REFUTED`. The v0.46 semantic denominator of 50 and its 48 release-lock
unit are retained as distinct fields from the current 57-lock fixture unit.
No cache hit, replay, or hash result can close semantic reuse eligibility by
itself.

## CI authority

GitHub Actions is the only verification authority. One job uses the same
checkout, `.gooo` contract, immutable fixture bundle, Go toolchain, local
fixture server, and runner for baseline and candidate. It records the exact
indicator vector:

`wall_ms, peak_rss_kib, requests, bytes_read, bytes_downloaded, selected, executed, reused, unknown, refuted`

Each indicator is judged independently. No aggregate improvement score is
computed. All generated outputs are written only below the caller-owned
runner temporary directory. Local build/test/verification executions and
repository writes are recorded as zero in the evidence dossier. Failed CI
runs upload uniquely named retained artifacts so history is not overwritten.

The optional public shared-ledger v0.48 identifiers are recorded as read-only
integration inputs. They are not required gates (`cross_project_required_gates=0`);
without a matched live pair, public-network improvement utility remains
`UNKNOWN`.

## Commands

The Actions job runs these commands from the repository root:

```text
gooo-content-addressed-proof-reuse conformance \
  --contract .gooo/content-addressed-proof-reuse.gooo \
  --bundle fixtures/fixture-bundle-v1.json

gooo-content-addressed-proof-reuse suite \
  --contract .gooo/content-addressed-proof-reuse.gooo \
  --bundle fixtures/fixture-bundle-v1.json \
  --repo-root . \
  --output-dir "$RUNNER_TEMP/gooo-content-addressed-proof-reuse/evidence"
```

The root README is intentionally excluded from inventory and tree accounting;
the implementation and fixture files are not.
