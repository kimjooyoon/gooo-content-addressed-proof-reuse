package reuse

import (
	"encoding/json"
	"fmt"
	"runtime"
	"syscall"
	"time"
)

func RunCase(contract Contract, bundle Bundle, fixtureCase FixtureCase, engine string, fetcher LockFetcher, stats StatsProvider) (CaseRun, error) {
	semantic, err := EvaluateSemantics(contract, bundle, fixtureCase)
	if err != nil {
		return CaseRun{}, err
	}
	if stats != nil {
		stats.ResetStats()
	}
	locks, executionMode := SelectForEngine(engine, semantic)
	if semantic.Plan.Status != Closed {
		semantic.Plan.Reused = []Lock{}
		semantic.Plan.Selected = append([]Lock(nil), locks...)
	}
	start := time.Now()
	beforeRSS := peakRSSKiB()
	fetched := make([]string, 0, len(locks))
	for _, lock := range locks {
		if _, err := fetcher.Fetch(lock); err != nil {
			return CaseRun{}, fmt.Errorf("%s %s fetch %s: %w", engine, fixtureCase.ID, lock.Coordinate, err)
		}
		fetched = append(fetched, lock.Coordinate)
	}
	wallMS := time.Since(start).Milliseconds()
	afterRSS := peakRSSKiB()
	if afterRSS < beforeRSS {
		afterRSS = beforeRSS
	}
	fetchStats := FetchStats{}
	if stats != nil {
		fetchStats = stats.Stats()
	}
	fullStatus := Closed
	fallbackReason, nextOperation, blockedBy := FallbackFields(semantic.Plan, executionMode)
	semantic.Plan.ExecutionMode = executionMode
	semantic.Plan.FallbackReason = fallbackReason
	semantic.Plan.NextOperation = nextOperation
	semantic.Plan.BlockedBy = append([]string(nil), blockedBy...)
	semantic.Evidence.FullVerificationStatus = fullStatus
	metrics := Metrics{WallMS: wallMS, PeakRSSKiB: afterRSS, Requests: fetchStats.Requests, BytesRead: fetchStats.BytesRead, BytesDownloaded: fetchStats.BytesDownloaded, Selected: len(locks), Executed: len(locks), Reused: len(semantic.Plan.Reused), Unknown: boolInt(semantic.Plan.Status == Unknown), Refuted: boolInt(semantic.Plan.Status == Refuted)}
	return CaseRun{Engine: engine, CaseID: fixtureCase.ID, Status: semantic.Plan.Status, ReuseEligibility: semantic.Plan.Status, ReuseClaim: semantic.Plan.Status, FullVerificationStatus: fullStatus, SemanticRoot: semantic.SemanticRoot, ExecutionMode: executionMode, FallbackReason: fallbackReason, NextOperation: nextOperation, BlockedBy: append([]string(nil), blockedBy...), Plan: semantic.Plan, Evidence: semantic.Evidence, Metrics: metrics, Fetched: fetched}, nil
}

func RunPair(contract Contract, bundle Bundle, fixtureCase FixtureCase, fetcher LockFetcher, stats StatsProvider) (CaseRun, CaseRun, CanonicalComparison, error) {
	baseline, err := RunCase(contract, bundle, fixtureCase, "baseline", fetcher, stats)
	if err != nil {
		return CaseRun{}, CaseRun{}, CanonicalComparison{}, err
	}
	candidate, err := RunCase(contract, bundle, fixtureCase, "candidate", fetcher, stats)
	if err != nil {
		return CaseRun{}, CaseRun{}, CanonicalComparison{}, err
	}
	baselineEvidence, err := json.Marshal(baseline.Evidence)
	if err != nil {
		return CaseRun{}, CaseRun{}, CanonicalComparison{}, err
	}
	candidateEvidence, err := json.Marshal(candidate.Evidence)
	if err != nil {
		return CaseRun{}, CaseRun{}, CanonicalComparison{}, err
	}
	comparison := CanonicalComparison{CaseID: fixtureCase.ID, SemanticRoot: candidate.SemanticRoot, Status: candidate.Status, UnknownEqual: equalUnknown(baseline.Evidence.Unknown, candidate.Evidence.Unknown), RefutedEqual: equalStrings(baseline.Evidence.Refuted, candidate.Evidence.Refuted), EvidenceEqual: string(baselineEvidence) == string(candidateEvidence)}
	if baseline.SemanticRoot != candidate.SemanticRoot || baseline.Status != candidate.Status || !comparison.UnknownEqual || !comparison.RefutedEqual || !comparison.EvidenceEqual {
		return baseline, candidate, comparison, fmt.Errorf("baseline/candidate semantic evidence diverged for %s", fixtureCase.ID)
	}
	return baseline, candidate, comparison, nil
}

func equalUnknown(left, right *UnknownRecord) bool {
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return string(leftBytes) == string(rightBytes)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func peakRSSKiB() int64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err == nil {
		value := int64(usage.Maxrss)
		if runtime.GOOS == "darwin" {
			return value / 1024
		}
		return value
	}
	return 0
}
