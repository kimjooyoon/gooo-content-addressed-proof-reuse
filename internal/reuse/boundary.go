package reuse

import "fmt"

func ValidateExecutionBoundary(fixtureCase FixtureCase, baseline, candidate CaseRun, currentLockCount int) error {
	if baseline.Metrics.Selected != currentLockCount || baseline.Metrics.Executed != currentLockCount || baseline.Metrics.Reused != 0 {
		return fmt.Errorf("baseline must execute the full current lock set for %s", fixtureCase.ID)
	}
	if candidate.FullVerificationStatus != Closed {
		return fmt.Errorf("full verification must close after deterministic fixture fetch for %s", fixtureCase.ID)
	}
	if candidate.Status != Closed {
		if candidate.ExecutionMode != "FULL_FALLBACK" || candidate.ReuseEligibility == Closed || candidate.ReuseClaim == Closed || candidate.Metrics.Reused != 0 || candidate.Metrics.Selected != currentLockCount || candidate.FallbackReason == "" || candidate.NextOperation == "" {
			return fmt.Errorf("fallback boundary was not preserved for %s", fixtureCase.ID)
		}
	}
	if fixtureCase.ID == "closed-full-reuse" && (candidate.Metrics.Reused != 57 || candidate.Metrics.Selected != 0 || candidate.Metrics.Executed != 0) {
		return fmt.Errorf("57-lock exact pair boundary failed")
	}
	if fixtureCase.ID == "closed-single-delta" && (candidate.Metrics.Reused != 56 || candidate.Metrics.Selected != 1 || candidate.Metrics.Executed != 1) {
		return fmt.Errorf("57-lock one-delta boundary failed")
	}
	return nil
}
