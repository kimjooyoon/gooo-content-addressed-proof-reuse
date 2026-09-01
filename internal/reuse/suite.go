package reuse

import "fmt"

func ValidateConformance(contract Contract, bundle Bundle) error {
	if err := ValidateContract(contract); err != nil {
		return err
	}
	if err := ValidateBundle(bundle); err != nil {
		return err
	}
	counts := map[Status]int{}
	for _, fixtureCase := range bundle.Cases {
		semantic, err := EvaluateSemantics(contract, bundle, fixtureCase)
		if err != nil {
			return fmt.Errorf("case %s: %w", fixtureCase.ID, err)
		}
		if semantic.Plan.Status != fixtureCase.ExpectedStatus {
			return fmt.Errorf("case %s: expected %s, got %s", fixtureCase.ID, fixtureCase.ExpectedStatus, semantic.Plan.Status)
		}
		counts[semantic.Plan.Status]++
		if fixtureCase.ID == "closed-full-reuse" && (len(semantic.Plan.Reused) != 57 || len(semantic.Plan.Selected) != 0) {
			return fmt.Errorf("closed-full-reuse must plan reused=57 selected=0")
		}
		if fixtureCase.ID == "closed-single-delta" && (len(semantic.Plan.Reused) != 56 || len(semantic.Plan.Selected) != 1) {
			return fmt.Errorf("closed-single-delta must plan reused=56 selected=1")
		}
	}
	if counts[Closed] != 3 || counts[Unknown] != 3 || counts[Refuted] != 3 {
		return fmt.Errorf("conformance distribution is CLOSED=%d UNKNOWN=%d REFUTED=%d", counts[Closed], counts[Unknown], counts[Refuted])
	}
	return nil
}

func RunSuite(contract Contract, bundle Bundle, fetcher LockFetcher, stats StatsProvider) (SuiteEvidence, error) {
	if err := ValidateConformance(contract, bundle); err != nil {
		return SuiteEvidence{}, err
	}
	evidence := SuiteEvidence{Schema: "gooo/content-addressed-proof-reuse/suite-evidence/v1", Protocol: contract.Protocol, BundleID: bundle.BundleID, Cases: append([]FixtureCase(nil), bundle.Cases...), Runs: []CaseRun{}, CanonicalComparisons: []CanonicalComparison{}, Summary: map[Status]int{}, Tests: TestRecord{Denominator: len(bundle.Cases), ExpectedDistribution: map[Status]int{Closed: 3, Unknown: 3, Refuted: 3}, LocalExecutions: 0}, Metrics: MetricsRecord{IndicatorVector: append([]string(nil), contract.IndicatorVector...), ByEngine: map[string]MetricsAggregate{"baseline": {Runs: []Metrics{}}, "candidate": {Runs: []Metrics{}}}, Judgments: independentJudgments(), PerformanceCaseIDs: []string{}, FallbackCaseIDs: []string{}}, Authority: AuthorityRecord{RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0, LocalVerificationRuns: 0, OutputLocation: "CALLER_OWNED_TEMP_ONLY", VerificationAuthority: "GITHUB_ACTIONS", CallerOwnedOutputOnly: true, GitHubTokenSource: "github.token", FailedHistoryPreserved: true}, Utility: UtilityRecord{Status: Unknown, MatchedLivePair: false, Reason: "NO_MATCHED_LIVE_PAIR", CrossProjectRequiredGates: 0}, GeneratedArtifacts: GeneratedArtifacts{Count: 0, Bytes: 0}, OptionalExternalInputs: []string{"shared-ledger:v0.48:release:380694027", "tag-object:e18e19e3ff58c4c7643d5028b5fc9bf39e83e769", "target:1985612b1f9eb8184e04626b426c88b91901bdd4", "asset:539891522:54641017:sha256:3792b53a77d88b3471e6f4889baf24679dcdcfba8314abb32d0dd3469ccdc3aa"}}
	for _, fixtureCase := range bundle.Cases {
		baseline, candidate, comparison, err := RunPair(contract, bundle, fixtureCase, fetcher, stats)
		if err != nil {
			return SuiteEvidence{}, err
		}
		evidence.Runs = append(evidence.Runs, baseline, candidate)
		evidence.CanonicalComparisons = append(evidence.CanonicalComparisons, comparison)
		evidence.Summary[candidate.Status]++
		if candidate.ExecutionMode == "FULL_FALLBACK" {
			evidence.Metrics.FallbackCaseIDs = append(evidence.Metrics.FallbackCaseIDs, fixtureCase.ID)
		}
		if candidate.ExecutionMode == "CONTENT_ADDRESSED_REUSE" {
			evidence.Metrics.PerformanceCaseIDs = append(evidence.Metrics.PerformanceCaseIDs, fixtureCase.ID)
			evidence.Metrics.ByEngine["baseline"] = appendMetric(evidence.Metrics.ByEngine["baseline"], baseline.Metrics)
			evidence.Metrics.ByEngine["candidate"] = appendMetric(evidence.Metrics.ByEngine["candidate"], candidate.Metrics)
		}
	}
	evidence.Tests.Closed = evidence.Summary[Closed]
	evidence.Tests.Unknown = evidence.Summary[Unknown]
	evidence.Tests.Refuted = evidence.Summary[Refuted]
	return evidence, nil
}

func appendMetric(aggregate MetricsAggregate, metric Metrics) MetricsAggregate {
	aggregate.Runs = append(aggregate.Runs, metric)
	return aggregate
}

func independentJudgments() map[string]string {
	return map[string]string{
		"wall_ms": "OBSERVED_PER_ELIGIBLE_CLOSED_PAIR_ONLY",
		"peak_rss_kib": "OBSERVED_PER_ELIGIBLE_CLOSED_PAIR_ONLY",
		"requests": "EXACT_CANDIDATE_DELTA_COUNT",
		"bytes_read": "EXACT_CANDIDATE_DELTA_BYTES",
		"bytes_downloaded": "EXACT_CANDIDATE_DELTA_BYTES",
		"selected": "EXACT_LOCK_SELECTION_COUNT",
		"executed": "EXACT_LOCK_EXECUTION_COUNT",
		"reused": "EXACT_LOCK_REUSE_COUNT",
		"unknown": "EXACT_REUSE_ELIGIBILITY_STATUS_COUNT",
		"refuted": "EXACT_REUSE_ELIGIBILITY_STATUS_COUNT",
	}
}
