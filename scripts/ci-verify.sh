#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: ci-verify.sh SUITE_EVIDENCE_JSON" >&2
  exit 64
fi

evidence=$1
vector='["wall_ms","peak_rss_kib","requests","bytes_read","bytes_downloaded","selected","executed","reused","unknown","refuted"]'

check() {
  label=$1
  filter=$2
  if ! jq -e --argjson vector "$vector" "$filter" "$evidence" >/dev/null; then
    echo "CI invariant failed: $label" >&2
    exit 1
  fi
}

check denominator '.tests.denominator == 9 and .tests.closed == 3 and .tests.unknown == 3 and .tests.refuted == 3 and .tests.local_executions == 0'
check indicator-vector '.metrics.indicator_vector == $vector'
check authority '.authority.repository_writes == 0 and .authority.local_build_executions == 0 and .authority.local_test_executions == 0 and .authority.local_verification_runs == 0 and .authority.caller_owned_output_only == true and .authority.verification_authority == "GITHUB_ACTIONS" and .authority.github_token_source == "github.token" and .authority.failed_history_preserved == true'
check utility '.utility.status == "UNKNOWN" and .utility.matched_live_pair == false and .utility.cross_project_required_gates == 0'
check inventory '.inventory.root_readme_excluded == true'
check canonical-equality '([.canonical_comparisons[] | (.unknown_equal and .refuted_equal and .canonical_evidence_equal and .semantic_root != "")] | all)'
check fallback-boundary '([.runs[] | select(.engine == "candidate" and (.case_id == "unknown-parent-missing" or .case_id == "unknown-parent-stale" or .case_id == "unknown-parent-ambiguous" or .case_id == "refuted-parent-asset-digest" or .case_id == "refuted-changed-dependency" or .case_id == "refuted-authority-escalation")) | (.execution_mode == "FULL_FALLBACK" and .reuse_eligibility != "CLOSED" and .reuse_claim != "CLOSED" and .full_verification_status == "CLOSED" and .metrics.reused == 0 and .metrics.selected == 57 and .fallback_reason != "" and .next_operation != "" and (.blocked_by | type == "array"))] | all)'
check exact-57-lock-reuse '([.runs[] | select(.engine == "candidate" and .case_id == "closed-full-reuse") | (.status == "CLOSED" and .metrics.reused == 57 and .metrics.selected == 0 and .metrics.executed == 0 and .metrics.requests == 0 and .metrics.bytes_read == 0 and .metrics.bytes_downloaded == 0)] | all)'
check exact-one-lock-delta '([.runs[] | select(.engine == "candidate" and .case_id == "closed-single-delta") | (.status == "CLOSED" and .metrics.reused == 56 and .metrics.selected == 1 and .metrics.executed == 1 and .metrics.requests == 1)] | all)'
check baseline-full-fetch '([.runs[] | select(.engine == "baseline" and (.case_id == "closed-full-reuse" or .case_id == "closed-single-delta")) | (.metrics.selected == 57 and .metrics.executed == 57 and .metrics.reused == 0 and .metrics.requests == 57)] | all)'

echo "CI invariants: exact denominator, fallback boundary, canonical equality, and 57-lock pairs verified"
