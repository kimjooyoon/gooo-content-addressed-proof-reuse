package reuse

func SelectForEngine(engine string, semantic semanticResult) ([]Lock, string) {
	if semantic.Plan.Status != Closed {
		return append([]Lock(nil), semantic.CurrentLocks...), "FULL_FALLBACK"
	}
	if engine == "baseline" {
		return append([]Lock(nil), semantic.CurrentLocks...), "FULL_BASELINE"
	}
	if len(semantic.Plan.Reused) == 0 {
		return append([]Lock(nil), semantic.Plan.Selected...), "FULL_EXECUTION"
	}
	return append([]Lock(nil), semantic.Plan.Selected...), "CONTENT_ADDRESSED_REUSE"
}

func FallbackFields(plan Plan, mode string) (string, string, []string) {
	if mode != "FULL_FALLBACK" {
		return "", "", []string{}
	}
	if plan.Unknown != nil {
		return plan.Reason, plan.Unknown.NextOperation, append([]string(nil), plan.Unknown.BlockedBy...)
	}
	return plan.Reason, "FULL_VERIFY_CURRENT_LOCK_SET", []string{}
}
