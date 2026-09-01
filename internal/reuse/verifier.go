package reuse

import "fmt"

type semanticResult struct {
	Plan         Plan
	Evidence     CanonicalEvidence
	CurrentLocks []Lock
	SemanticRoot string
}

func EvaluateSemantics(contract Contract, bundle Bundle, fixtureCase FixtureCase) (semanticResult, error) {
	currentLocks, err := CurrentLocks(bundle.ParentReceipt.Locks, fixtureCase.CurrentLockSet)
	if err != nil {
		return semanticResult{}, err
	}
	parent := ParentForCase(bundle.ParentReceipt, fixtureCase)
	root := ComputeSemanticRoot(currentLocks, fixtureCase.CurrentDependencyDigest)
	plan := Plan{Status: Closed, Reused: []Lock{}, Selected: []Lock{}, Authority: fixtureCase.CurrentAuthority, BlockedBy: []string{}, CacheHit: fixtureCase.CacheHit, ReplayObserved: fixtureCase.ReplayObserved, HashObserved: fixtureCase.HashObserved}
	events := make([]EvidenceEvent, 0, 9)
	add := func(id string, state Status, reason string) {
		events = append(events, EvidenceEvent{Ordinal: len(events) + 1, ID: id, State: state, Reason: reason})
	}
	unknown := func(reason, class, next string, blocked []string) semanticResult {
		record := &UnknownRecord{Stage: "REUSE_ELIGIBILITY", Step: "VERIFY_PARENT_RECEIPT_AND_AUTHORITY", Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: append([]string(nil), blocked...)}
		plan.Status = Unknown
		plan.Reason = reason
		plan.Unknown = record
		plan.NextOperation = next
		plan.BlockedBy = append([]string(nil), blocked...)
		add("eligibility", Unknown, reason)
		return semanticResult{Plan: plan, CurrentLocks: currentLocks, SemanticRoot: root, Evidence: canonicalEvidence(fixtureCase.ID, root, plan, Closed, events)}
	}
	refuted := func(reason, id string) semanticResult {
		plan.Status = Refuted
		plan.Reason = reason
		plan.BlockedBy = []string{}
		add(id, Refuted, reason)
		return semanticResult{Plan: plan, CurrentLocks: currentLocks, SemanticRoot: root, Evidence: canonicalEvidence(fixtureCase.ID, root, plan, Closed, events)}
	}

	if parent == nil {
		return unknown("PARENT_RECEIPT_MISSING", "MISSING_EVIDENCE", "FETCH_PARENT_RECEIPT", []string{"parent_receipt"}), nil
	}
	if parent.Freshness == "STALE" {
		return unknown("PARENT_RECEIPT_STALE", "STALE_EVIDENCE", "REFRESH_PARENT_RECEIPT", []string{"parent_receipt.freshness"}), nil
	}
	if parent.AmbiguityCount != 0 {
		return unknown("PARENT_RECEIPT_AMBIGUOUS", "AMBIGUOUS_EVIDENCE", "DISAMBIGUATE_PARENT_RECEIPT", []string{"parent_receipt.ambiguity_count"}), nil
	}
	if parent.ReceiptID == "" || parent.Status == "" || parent.Freshness == "" {
		return unknown("PARENT_RECEIPT_FIELDS_MISSING", "MISSING_EVIDENCE", "FETCH_COMPLETE_PARENT_RECEIPT", []string{"parent_receipt.receipt_id", "parent_receipt.status"}), nil
	}
	if parent.Status != "PASS" {
		return refuted("PARENT_RECEIPT_NOT_PASS", "parent-receipt"), nil
	}
	add("parent-receipt", Closed, "PARENT_RECEIPT_PRESENT")

	if !parent.Release.ActualImmutable {
		return refuted("PARENT_RELEASE_NOT_ACTUALLY_IMMUTABLE", "parent-release"), nil
	}
	if parent.Release != contract.ParentRelease {
		return refuted("ANNOTATED_TAG_TARGET_OR_SOURCE_ARTIFACT_MISMATCH", "release-coordinates"), nil
	}
	add("release-coordinates", Closed, "ANNOTATED_TAG_TARGET_SOURCE_ARTIFACT_EXACT")

	if parent.Asset.ArtifactID == 0 || parent.Asset.ArtifactSize == 0 || parent.Asset.ObservedDigest == "" {
		return unknown("PARENT_ASSET_DIGEST_MISSING", "MISSING_EVIDENCE", "FETCH_AND_VERIFY_PARENT_ASSET_DIGEST", []string{"parent_asset.digest"}), nil
	}
	if !parent.Asset.Verified || parent.Asset.ArtifactID != contract.ParentAsset.ArtifactID || parent.Asset.ArtifactSize != contract.ParentAsset.ArtifactSize || parent.Asset.ObservedDigest != contract.ParentAsset.ArtifactDigest {
		return refuted("PARENT_ASSET_DIGEST_MISMATCH", "parent-asset"), nil
	}
	add("parent-asset", Closed, "PARENT_ASSET_DIGEST_VERIFIED")

	if parent.RootMaterial == "" || parent.ReceiptRoot == "" || parent.SemanticRoot == "" {
		return unknown("PARENT_RECEIPT_ROOT_MISSING", "MISSING_EVIDENCE", "FETCH_AND_VERIFY_PARENT_RECEIPT_ROOT", []string{"parent_receipt_root"}), nil
	}
	if DigestString(parent.RootMaterial) != parent.ReceiptRoot {
		return refuted("PARENT_RECEIPT_ROOT_DIGEST_MISMATCH", "receipt-root"), nil
	}
	if ComputeSemanticRoot(parent.Locks, parent.DependencyDigest) != parent.SemanticRoot || parent.SemanticRoot != contract.ParentSemanticRoot {
		return refuted("PARENT_SEMANTIC_ROOT_MISMATCH", "semantic-root"), nil
	}
	add("receipt-root", Closed, "PARENT_RECEIPT_ROOT_AND_SEMANTIC_ROOT_VERIFIED")

	if parent.DependencyDigest == "" || fixtureCase.CurrentDependencyDigest == "" {
		return unknown("DEPENDENCY_DIGEST_MISSING", "MISSING_EVIDENCE", "FETCH_CURRENT_DEPENDENCY_DIGEST", []string{"dependency_digest"}), nil
	}
	if parent.DependencyDigest != fixtureCase.CurrentDependencyDigest {
		return refuted("CHANGED_DEPENDENCY", "dependency"), nil
	}
	add("dependency", Closed, "DEPENDENCY_UNAFFECTED")

	if fixtureCase.CurrentAuthority == "" {
		return unknown("CACHE_AUTHORITY_MISSING", "MISSING_EVIDENCE", "PROVIDE_CACHE_AUTHORITY", []string{"cache_authority"}), nil
	}
	if fixtureCase.CurrentAuthority != contract.CacheAuthority {
		return refuted("AUTHORITY_ESCALATION", "authority"), nil
	}
	add("authority", Closed, "CACHE_AUTHORITY_EXACT")

	if err := validateLocks(parent.Locks); err != nil {
		return unknown("PARENT_LOCK_SET_AMBIGUOUS", "AMBIGUOUS_EVIDENCE", "DISAMBIGUATE_PARENT_LOCK_SET", []string{"parent_receipt.locks"}), nil
	}
	if err := validateLocks(currentLocks); err != nil {
		return refuted("CURRENT_LOCK_SET_INVALID", "current-lock-set"), nil
	}
	parentByCoordinate := make(map[string]Lock, len(parent.Locks))
	for _, lock := range parent.Locks {
		parentByCoordinate[lock.Coordinate] = lock
	}
	for _, current := range currentLocks {
		prior, found := parentByCoordinate[current.Coordinate]
		if found && prior.Digest == current.Digest {
			plan.Reused = append(plan.Reused, current)
		} else {
			plan.Selected = append(plan.Selected, current)
		}
	}
	plan.Reason = "CONTENT_ADDRESSED_REUSE_ELIGIBILITY_CLOSED"
	plan.DependencyUnaffected = true
	add("lock-set", Closed, "CURRENT_LOCK_COORDINATE_AND_DIGEST_COMPARISON_COMPLETE")
	add("plan", Closed, "CONTENT_ADDRESSED_REUSE_PLAN_CLOSED")
	return semanticResult{Plan: plan, CurrentLocks: currentLocks, SemanticRoot: root, Evidence: canonicalEvidence(fixtureCase.ID, root, plan, Closed, events)}, nil
}

func ParentForCase(base ParentReceipt, fixtureCase FixtureCase) *ParentReceipt {
	if fixtureCase.ParentState == "missing" {
		return nil
	}
	parent := base
	parent.Locks = append([]Lock(nil), base.Locks...)
	switch fixtureCase.ParentState {
	case "stale":
		parent.Freshness = "STALE"
	case "ambiguous":
		parent.AmbiguityCount = 2
	}
	if fixtureCase.AssetObservedDigest != "" {
		parent.Asset.ObservedDigest = fixtureCase.AssetObservedDigest
	}
	return &parent
}

func CurrentLocks(parent []Lock, lockSet string) ([]Lock, error) {
	locks := append([]Lock(nil), parent...)
	switch lockSet {
	case "parent":
		return locks, nil
	case "one-delta":
		if len(locks) == 0 {
			return nil, fmt.Errorf("one-delta requires a non-empty parent lock set")
		}
		locks[0].Digest = "sha256:changed-lock-001"
		return locks, nil
	case "all-delta":
		for index := range locks {
			locks[index].Digest = fmt.Sprintf("sha256:changed-lock-%03d", index+1)
		}
		return locks, nil
	default:
		return nil, fmt.Errorf("unknown lock set %q", lockSet)
	}
}

func canonicalEvidence(caseID, root string, plan Plan, fullStatus Status, events []EvidenceEvent) CanonicalEvidence {
	refuted := []string{}
	if plan.Status == Refuted {
		refuted = []string{plan.Reason}
	}
	return CanonicalEvidence{Schema: "gooo/content-addressed-proof-reuse/canonical-evidence/v1", CaseID: caseID, Status: plan.Status, ReuseEligibility: plan.Status, ReuseClaim: plan.Status, FullVerificationStatus: fullStatus, SemanticRoot: root, ProofInputs: ProofInputs{CacheHit: plan.CacheHit, ReplayObserved: plan.ReplayObserved, HashObserved: plan.HashObserved}, Unknown: plan.Unknown, Refuted: refuted, Events: events}
}
