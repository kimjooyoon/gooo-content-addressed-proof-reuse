# Content-addressed proof reuse protocol v1

## Safety boundary

`CLOSED` for reuse eligibility means that the evidence identity is complete,
not merely that a cache or replay was found. The decision requires all of:

- the parent release is actually immutable;
- the annotated tag object and target commit match the declared release;
- the source artifact ID, size, and digest match;
- the parent asset digest and receipt root are verified;
- the parent semantic root is the SHA-256 root of the canonical ordered lock
  coordinates and digests plus the unchanged dependency digest;
- the cache authority is exact; and
- every reused current lock has an exact coordinate+digest match and no
  invalidation edge has fired.

The candidate can then verify only the selected delta. A changed lock is not a
refutation: it is selected for current verification. A changed dependency,
authority escalation, or digest contradiction is a `REFUTED` reuse claim.

## Fallback semantics

Missing, stale, or ambiguous parent evidence is `UNKNOWN`. A fallback full
verification can continue to validate the current 57-lock set and can be
`CLOSED` as a full-verification result. It cannot change the reuse claim. The
receipt therefore carries both statuses:

```text
reuse_eligibility=UNKNOWN|REFUTED
reuse_claim=UNKNOWN|REFUTED
full_verification_status=CLOSED
execution_mode=FULL_FALLBACK
reused=0
selected=57
fallback_reason=<original reason>
next_operation=<frontier operation>
blocked_by=<frontier>
```

Fallback cases are not performance comparisons. Only canonical closed cases
where content-addressed reuse actually executes are in the improvement set.

## Denominators and evidence

The 50 semantic denominator from v0.46 and its 48 release-lock unit are
preserved as historical fields. The current fixture uses 57 release locks;
these are different units and are never silently substituted. Baseline and
candidate run against the same deterministic local server in one Actions job.
Semantic root, status, UNKNOWN record, REFUTED reasons, and canonical evidence
must compare byte-for-byte. Runtime/resource observations remain independent
indicator fields.
