package reuse

import (
	"fmt"
	"strings"
)

func RenderHumanReport(evidence SuiteEvidence) string {
	var report strings.Builder
	report.WriteString("# Content-addressed proof reuse evidence\n\n")
	fmt.Fprintf(&report, "Protocol: `%s`  \nFixture bundle: `%s`  \n\n", evidence.Protocol, evidence.BundleID)
	report.WriteString("The reuse eligibility claim is separate from full verification. UNKNOWN and REFUTED are retained even when a full fallback verification closes successfully. Cache hits, replay, and hashing alone never close a reuse claim.\n\n")
	fmt.Fprintf(&report, "## Canonical distribution\n\nCLOSED=%d, UNKNOWN=%d, REFUTED=%d across a fixed denominator of %d cases.\n\n", evidence.Tests.Closed, evidence.Tests.Unknown, evidence.Tests.Refuted, evidence.Tests.Denominator)
	report.WriteString("## Execution and exact indicator vectors\n\n")
	report.WriteString("| engine | case | mode | status | wall_ms | peak_rss_kib | requests | bytes_read | bytes_downloaded | selected | executed | reused | unknown | refuted |\n|---|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, run := range evidence.Runs {
		m := run.Metrics
		fmt.Fprintf(&report, "| %s | %s | %s | %s | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d |\n", run.Engine, run.CaseID, run.ExecutionMode, run.Status, m.WallMS, m.PeakRSSKiB, m.Requests, m.BytesRead, m.BytesDownloaded, m.Selected, m.Executed, m.Reused, m.Unknown, m.Refuted)
	}
	report.WriteString("\nFallback cases preserve their eligibility state and are excluded from improvement evidence: ")
	report.WriteString(strings.Join(evidence.Metrics.FallbackCaseIDs, ", "))
	report.WriteString(".\n\n")
	report.WriteString("## Fallback frontier\n\n")
	report.WriteString("| case | reuse eligibility | reuse claim | full verification | reason | next operation | blocked_by |\n|---|---|---|---|---|---|---|\n")
	for _, run := range evidence.Runs {
		if run.Engine != "candidate" || run.ExecutionMode != "FULL_FALLBACK" {
			continue
		}
		fmt.Fprintf(&report, "| %s | %s | %s | %s | %s | %s | %s |\n", run.CaseID, run.ReuseEligibility, run.ReuseClaim, run.FullVerificationStatus, run.FallbackReason, run.NextOperation, strings.Join(run.BlockedBy, ","))
	}
	report.WriteString("\n")
	report.WriteString("## Canonical baseline/candidate equality\n\n")
	report.WriteString("| case | status | semantic root | UNKNOWN equal | REFUTED equal | canonical evidence equal |\n|---|---|---|---|---|---|\n")
	for _, comparison := range evidence.CanonicalComparisons {
		fmt.Fprintf(&report, "| %s | %s | `%s` | %t | %t | %t |\n", comparison.CaseID, comparison.Status, comparison.SemanticRoot, comparison.UnknownEqual, comparison.RefutedEqual, comparison.EvidenceEqual)
	}
	report.WriteString("\n## Authority and utility\n\n")
	fmt.Fprintf(&report, "Verification authority: `%s`; repository writes: %d; local build executions: %d; local test executions: %d; local verification runs: %d; cross-project required gates: %d.\n\n", evidence.Authority.VerificationAuthority, evidence.Authority.RepositoryWrites, evidence.Authority.LocalBuildExecutions, evidence.Authority.LocalTestExecutions, evidence.Authority.LocalVerificationRuns, evidence.Utility.CrossProjectRequiredGates)
	fmt.Fprintf(&report, "Public-network improvement utility: `%s` (`%s`); matched live pair: %t.\n", evidence.Utility.Status, evidence.Utility.Reason, evidence.Utility.MatchedLivePair)
	return report.String()
}
