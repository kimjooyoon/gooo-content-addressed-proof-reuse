package reuse_test

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-content-addressed-proof-reuse/internal/fixture"
	"github.com/kimjooyoon/gooo-content-addressed-proof-reuse/internal/reuse"
)

func loadInputs(t *testing.T) (reuse.Contract, reuse.Bundle) {
	t.Helper()
	root := filepath.Join("..", "..")
	contract, _, err := reuse.ParseContract(filepath.Join(root, ".gooo", "content-addressed-proof-reuse.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	bundle, _, err := reuse.LoadBundle(filepath.Join(root, "fixtures", "fixture-bundle-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return contract, bundle
}

func TestCanonicalDistributionAndExactLockDeltas(t *testing.T) {
	contract, bundle := loadInputs(t)
	if err := reuse.ValidateConformance(contract, bundle); err != nil {
		t.Fatal(err)
	}
	server := fixture.NewServer(bundle)
	defer server.Close()
	for _, fixtureCase := range bundle.Cases[:2] {
		baseline, candidate, comparison, err := reuse.RunPair(contract, bundle, fixtureCase, server, server)
		if err != nil {
			t.Fatal(err)
		}
		if !comparison.EvidenceEqual || baseline.SemanticRoot != candidate.SemanticRoot {
			t.Fatalf("semantic pair mismatch for %s", fixtureCase.ID)
		}
		if fixtureCase.ID == "closed-full-reuse" && (candidate.Metrics.Reused != 57 || candidate.Metrics.Selected != 0 || candidate.Metrics.Requests != 0) {
			t.Fatalf("unexpected full reuse metrics: %+v", candidate.Metrics)
		}
		if fixtureCase.ID == "closed-single-delta" && (candidate.Metrics.Reused != 56 || candidate.Metrics.Selected != 1 || candidate.Metrics.Requests != 1) {
			t.Fatalf("unexpected delta metrics: %+v", candidate.Metrics)
		}
	}
}

func TestFallbackDoesNotPromoteEligibility(t *testing.T) {
	contract, bundle := loadInputs(t)
	server := fixture.NewServer(bundle)
	defer server.Close()
	fixtureCase := bundle.Cases[3]
	_, candidate, _, err := reuse.RunPair(contract, bundle, fixtureCase, server, server)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.Status != reuse.Unknown || candidate.ReuseClaim != reuse.Unknown || candidate.ExecutionMode != "FULL_FALLBACK" {
		t.Fatalf("fallback changed eligibility: %+v", candidate)
	}
	if candidate.FullVerificationStatus != reuse.Closed || candidate.Metrics.Reused != 0 || candidate.Metrics.Selected != 57 || candidate.NextOperation != "FETCH_PARENT_RECEIPT" || len(candidate.BlockedBy) != 1 {
		t.Fatalf("fallback accounting is incomplete: %+v", candidate)
	}
}

func TestParentSemanticRootIsContentAddressed(t *testing.T) {
	contract, bundle := loadInputs(t)
	root := reuse.ComputeSemanticRoot(bundle.ParentReceipt.Locks, bundle.ParentReceipt.DependencyDigest)
	if root != contract.ParentSemanticRoot || root != bundle.ParentReceipt.SemanticRoot {
		t.Fatalf("parent semantic root mismatch: %s %s %s", root, contract.ParentSemanticRoot, bundle.ParentReceipt.SemanticRoot)
	}
}
