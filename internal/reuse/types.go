package reuse

type Status string

const (
	Closed  Status = "CLOSED"
	Unknown Status = "UNKNOWN"
	Refuted Status = "REFUTED"
)

var UnknownFields = []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}

type UnknownRecord struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type ReleaseIdentity struct {
	Project        string `json:"project"`
	TagName        string `json:"tag_name"`
	ReleaseID      int64  `json:"release_id"`
	TagObject      string `json:"tag_object"`
	TargetCommit   string `json:"target_commit"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactSize   int64  `json:"artifact_size"`
	ArtifactDigest string `json:"artifact_digest"`
	ActualImmutable bool  `json:"actual_immutable"`
}

type AssetProof struct {
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactSize   int64  `json:"artifact_size"`
	ObservedDigest string `json:"observed_digest"`
	Verified       bool   `json:"verified"`
}

type Lock struct {
	Ordinal    int    `json:"ordinal"`
	Coordinate string `json:"coordinate"`
	Digest     string `json:"digest"`
}

type ParentReceipt struct {
	ReceiptID        string         `json:"receipt_id"`
	Status           string         `json:"status"`
	Freshness        string         `json:"freshness"`
	AmbiguityCount   int            `json:"ambiguity_count"`
	Release          ReleaseIdentity `json:"release"`
	Asset            AssetProof     `json:"asset"`
	SemanticRoot     string         `json:"semantic_root"`
	ReceiptRoot      string         `json:"receipt_root"`
	RootMaterial     string         `json:"root_material"`
	DependencyDigest string         `json:"dependency_digest"`
	Locks            []Lock         `json:"locks"`
}

type FixtureCase struct {
	ID                     string `json:"id"`
	Description            string `json:"description"`
	ParentState            string `json:"parent_state"`
	CurrentLockSet         string `json:"current_lock_set"`
	CurrentDependencyDigest string `json:"current_dependency_digest"`
	CurrentAuthority       string `json:"current_authority"`
	AssetObservedDigest    string `json:"asset_observed_digest,omitempty"`
	CacheHit               bool   `json:"cache_hit"`
	ReplayObserved         bool   `json:"replay_observed"`
	HashObserved           bool   `json:"hash_observed"`
	ExpectedStatus         Status `json:"expected_status"`
}

type Bundle struct {
	Schema        string        `json:"schema"`
	BundleID      string        `json:"bundle_id"`
	HashAlgorithm string        `json:"hash_algorithm"`
	ParentReceipt ParentReceipt  `json:"parent_receipt"`
	Cases         []FixtureCase  `json:"cases"`
}

type Contract struct {
	Protocol              string
	ParentRelease         ReleaseIdentity
	ParentAsset           ReleaseIdentity
	ParentSemanticRoot    string
	CacheAuthority        string
	HashAlgorithm         string
	ReplayObligations     []string
	CanonicalOrder        []string
	InvalidationEdges     []string
	RetryCount            int
	TimeoutMS             int
	UnknownFields         []string
	IndicatorVector       []string
	RequiredCaseStatuses  map[Status]int
}

type Plan struct {
	Status               Status         `json:"status"`
	Reason               string         `json:"reason"`
	Unknown              *UnknownRecord `json:"unknown,omitempty"`
	Reused               []Lock         `json:"reused"`
	Selected             []Lock         `json:"selected"`
	DependencyUnaffected bool         `json:"dependency_unaffected"`
	Authority             string         `json:"cache_authority"`
	ExecutionMode         string         `json:"execution_mode"`
	FallbackReason        string         `json:"fallback_reason"`
	NextOperation         string         `json:"next_operation"`
	BlockedBy             []string       `json:"blocked_by"`
	CacheHit              bool           `json:"cache_hit"`
	ReplayObserved        bool           `json:"replay_observed"`
	HashObserved          bool           `json:"hash_observed"`
}

type ProofInputs struct {
	CacheHit       bool `json:"cache_hit"`
	ReplayObserved bool `json:"replay_observed"`
	HashObserved   bool `json:"hash_observed"`
}

type EvidenceEvent struct {
	Ordinal int    `json:"ordinal"`
	ID      string `json:"id"`
	State   Status `json:"state"`
	Reason  string `json:"reason"`
}

type CanonicalEvidence struct {
	Schema               string          `json:"schema"`
	CaseID               string          `json:"case_id"`
	Status               Status          `json:"status"`
	ReuseEligibility     Status          `json:"reuse_eligibility"`
	ReuseClaim           Status          `json:"reuse_claim"`
	FullVerificationStatus Status        `json:"full_verification_status"`
	SemanticRoot         string          `json:"semantic_root"`
	ProofInputs          ProofInputs     `json:"proof_inputs"`
	Unknown              *UnknownRecord  `json:"unknown,omitempty"`
	Refuted              []string        `json:"refuted"`
	Events               []EvidenceEvent `json:"events"`
}

type Metrics struct {
	WallMS          int64 `json:"wall_ms"`
	PeakRSSKiB      int64 `json:"peak_rss_kib"`
	Requests        int64 `json:"requests"`
	BytesRead       int64 `json:"bytes_read"`
	BytesDownloaded int64 `json:"bytes_downloaded"`
	Selected        int   `json:"selected"`
	Executed        int   `json:"executed"`
	Reused          int   `json:"reused"`
	Unknown         int   `json:"unknown"`
	Refuted         int   `json:"refuted"`
}

type FetchStats struct {
	Requests        int64 `json:"requests"`
	BytesRead       int64 `json:"bytes_read"`
	BytesDownloaded int64 `json:"bytes_downloaded"`
}

type LockFetcher interface {
	Fetch(Lock) ([]byte, error)
}

type StatsProvider interface {
	ResetStats()
	Stats() FetchStats
}

type CaseRun struct {
	Engine                 string            `json:"engine"`
	CaseID                 string            `json:"case_id"`
	Status                 Status            `json:"status"`
	ReuseEligibility       Status            `json:"reuse_eligibility"`
	ReuseClaim             Status            `json:"reuse_claim"`
	FullVerificationStatus Status            `json:"full_verification_status"`
	SemanticRoot           string            `json:"semantic_root"`
	ExecutionMode          string            `json:"execution_mode"`
	FallbackReason         string            `json:"fallback_reason"`
	NextOperation          string            `json:"next_operation"`
	BlockedBy              []string          `json:"blocked_by"`
	Plan                   Plan              `json:"plan"`
	Evidence               CanonicalEvidence `json:"evidence"`
	Metrics                Metrics           `json:"metrics"`
	Fetched                []string          `json:"fetched"`
}

type Inventory struct {
	RegularFiles       int64 `json:"regular_files"`
	TreeBytes          int64 `json:"tree_bytes"`
	GoFiles            int64 `json:"go_files"`
	GoLines            int64 `json:"go_lines"`
	GoooFiles          int64 `json:"gooo_files"`
	GoooLines          int64 `json:"gooo_lines"`
	FixtureFiles       int64 `json:"fixture_files"`
	ContractFiles      int64 `json:"contract_files"`
	TestFiles          int64 `json:"test_files"`
	RootReadmeExcluded bool  `json:"root_readme_excluded"`
}

type UtilityRecord struct {
	Status                Status `json:"status"`
	MatchedLivePair       bool   `json:"matched_live_pair"`
	Reason                string `json:"reason"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
}

type AuthorityRecord struct {
	RepositoryWrites       int    `json:"repository_writes"`
	LocalBuildExecutions   int    `json:"local_build_executions"`
	LocalTestExecutions    int    `json:"local_test_executions"`
	LocalVerificationRuns  int    `json:"local_verification_runs"`
	OutputLocation         string `json:"output_location"`
	VerificationAuthority  string `json:"verification_authority"`
	CallerOwnedOutputOnly  bool   `json:"caller_owned_output_only"`
	GitHubTokenSource      string `json:"github_token_source"`
	FailedHistoryPreserved bool   `json:"failed_history_preserved"`
}

type SuiteEvidence struct {
	Schema             string             `json:"schema"`
	Protocol           string             `json:"protocol"`
	BundleID           string             `json:"bundle_id"`
	ContractDigest     string             `json:"contract_digest"`
	FixtureDigest      string             `json:"fixture_digest"`
	Cases              []FixtureCase      `json:"cases"`
	Runs               []CaseRun          `json:"runs"`
	CanonicalComparisons []CanonicalComparison `json:"canonical_comparisons"`
	Summary            map[Status]int    `json:"summary"`
	Inventory          Inventory          `json:"inventory"`
	Tests              TestRecord        `json:"tests"`
	Metrics            MetricsRecord     `json:"metrics"`
	Authority          AuthorityRecord   `json:"authority"`
	Utility            UtilityRecord     `json:"utility"`
	GeneratedArtifacts GeneratedArtifacts `json:"generated_artifacts"`
	OptionalExternalInputs []string       `json:"optional_external_inputs"`
}

type CanonicalComparison struct {
	CaseID       string `json:"case_id"`
	SemanticRoot string `json:"semantic_root"`
	Status       Status `json:"status"`
	UnknownEqual bool   `json:"unknown_equal"`
	RefutedEqual bool   `json:"refuted_equal"`
	EvidenceEqual bool  `json:"canonical_evidence_equal"`
}

type TestRecord struct {
	Denominator      int            `json:"denominator"`
	Closed           int            `json:"closed"`
	Unknown          int            `json:"unknown"`
	Refuted          int            `json:"refuted"`
	ExpectedDistribution map[Status]int `json:"expected_distribution"`
	LocalExecutions  int            `json:"local_executions"`
}

type MetricsRecord struct {
	IndicatorVector []string `json:"indicator_vector"`
	ByEngine        map[string]MetricsAggregate `json:"by_engine"`
	Judgments       map[string]string `json:"independent_judgments"`
	PerformanceCaseIDs []string `json:"performance_case_ids"`
	FallbackCaseIDs   []string `json:"fallback_case_ids"`
}

type MetricsAggregate struct {
	Runs []Metrics `json:"runs"`
}

type GeneratedArtifacts struct {
	Count int   `json:"count"`
	Bytes int64 `json:"bytes"`
}
