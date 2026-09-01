package reuse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const protocol = "gooo/content-addressed-proof-reuse/v1"

func ParseContract(path string) (Contract, string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, "", err
	}
	contract, err := parseContractLines(string(contents))
	if err != nil {
		return Contract{}, "", err
	}
	return contract, DigestBytes(contents), nil
}

func parseContractLines(contents string) (Contract, error) {
	contract := Contract{
		Protocol: protocol,
		UnknownFields: append([]string(nil), UnknownFields...),
		IndicatorVector: []string{"wall_ms", "peak_rss_kib", "requests", "bytes_read", "bytes_downloaded", "selected", "executed", "reused", "unknown", "refuted"},
		RequiredCaseStatuses: map[Status]int{Closed: 3, Unknown: 3, Refuted: 3},
		RetryCount: -1,
		TimeoutMS: -1,
	}
	scanner := bufio.NewScanner(strings.NewReader(contents))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 2 {
			return Contract{}, fmt.Errorf(".gooo line %d: expected key|value", lineNumber)
		}
		switch fields[0] {
		case "protocol":
			contract.Protocol = fields[1]
		case "parent_release":
			if len(fields) != 7 {
				return Contract{}, fmt.Errorf(".gooo line %d: parent_release requires seven fields", lineNumber)
			}
			releaseID, parseErr := strconv.ParseInt(fields[3], 10, 64)
			if parseErr != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: release id: %w", lineNumber, parseErr)
			}
			immutable, parseErr := strconv.ParseBool(fields[6])
			if parseErr != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: immutable flag: %w", lineNumber, parseErr)
			}
			contract.ParentRelease = ReleaseIdentity{Project: fields[1], TagName: fields[2], ReleaseID: releaseID, TagObject: fields[4], TargetCommit: fields[5], ActualImmutable: immutable}
		case "parent_asset":
			if len(fields) != 4 {
				return Contract{}, fmt.Errorf(".gooo line %d: parent_asset requires four fields", lineNumber)
			}
			artifactID, parseErr := strconv.ParseInt(fields[1], 10, 64)
			if parseErr != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: asset id: %w", lineNumber, parseErr)
			}
			artifactSize, parseErr := strconv.ParseInt(fields[2], 10, 64)
			if parseErr != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: asset size: %w", lineNumber, parseErr)
			}
			contract.ParentAsset = ReleaseIdentity{ArtifactID: artifactID, ArtifactSize: artifactSize, ArtifactDigest: fields[3]}
		case "parent_semantic_root":
			contract.ParentSemanticRoot = fields[1]
		case "cache_authority":
			contract.CacheAuthority = fields[1]
		case "hash_algorithm":
			contract.HashAlgorithm = fields[1]
		case "replay_obligation":
			contract.ReplayObligations = append(contract.ReplayObligations, fields[1])
		case "canonical_order":
			contract.CanonicalOrder = append(contract.CanonicalOrder, fields[1])
		case "invalidation_edge":
			contract.InvalidationEdges = append(contract.InvalidationEdges, fields[1])
		case "retry":
			parseErr, parseError := strconv.Atoi(fields[1])
			if parseError != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: retry: %w", lineNumber, parseError)
			}
			contract.RetryCount = parseErr
		case "timeout_ms":
			parseValue, parseError := strconv.Atoi(fields[1])
			if parseError != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: timeout: %w", lineNumber, parseError)
			}
			contract.TimeoutMS = parseValue
		case "unknown_fields":
			contract.UnknownFields = strings.Split(fields[1], ",")
		case "indicator_vector":
			contract.IndicatorVector = strings.Split(fields[1], ",")
		case "required_case_status":
			if len(fields) != 3 {
				return Contract{}, fmt.Errorf(".gooo line %d: required_case_status requires three fields", lineNumber)
			}
			count, parseErr := strconv.Atoi(fields[2])
			if parseErr != nil {
				return Contract{}, fmt.Errorf(".gooo line %d: status count: %w", lineNumber, parseErr)
			}
			contract.RequiredCaseStatuses[Status(fields[1])] = count
		case "source_role", "cross_project_required_gates", "current_lock_set", "historical_denominator", "current_denominator", "lock_identity":
			// These are explicit contract declarations consumed by the report.
		default:
			return Contract{}, fmt.Errorf(".gooo line %d: unknown declaration %q", lineNumber, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return Contract{}, err
	}
	if err := ValidateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func ValidateContract(contract Contract) error {
	if contract.Protocol != protocol {
		return fmt.Errorf("unexpected protocol %q", contract.Protocol)
	}
	if contract.ParentRelease.Project == "" || contract.ParentRelease.TagName == "" || contract.ParentRelease.ReleaseID == 0 || contract.ParentRelease.TagObject == "" || contract.ParentRelease.TargetCommit == "" {
		return fmt.Errorf("parent release identity is incomplete")
	}
	if !contract.ParentRelease.ActualImmutable {
		return fmt.Errorf("contract parent release must declare actual immutable=true")
	}
	if contract.ParentAsset.ArtifactID == 0 || contract.ParentAsset.ArtifactSize == 0 || contract.ParentAsset.ArtifactDigest == "" {
		return fmt.Errorf("parent asset identity is incomplete")
	}
	if contract.ParentSemanticRoot == "" || contract.CacheAuthority == "" || contract.HashAlgorithm != "sha256" {
		return fmt.Errorf("semantic root, cache authority, and sha256 are required")
	}
	if contract.RetryCount < 0 || contract.TimeoutMS <= 0 {
		return fmt.Errorf("retry/timeout declaration is invalid")
	}
	if strings.Join(contract.UnknownFields, ",") != strings.Join(UnknownFields, ",") {
		return fmt.Errorf("UNKNOWN must declare the exact six-field frontier")
	}
	if strings.Join(contract.IndicatorVector, ",") != "wall_ms,peak_rss_kib,requests,bytes_read,bytes_downloaded,selected,executed,reused,unknown,refuted" {
		return fmt.Errorf("indicator vector is not exact")
	}
	if contract.RequiredCaseStatuses[Closed] != 3 || contract.RequiredCaseStatuses[Unknown] != 3 || contract.RequiredCaseStatuses[Refuted] != 3 {
		return fmt.Errorf("canonical case distribution must be CLOSED=3 UNKNOWN=3 REFUTED=3")
	}
	if len(contract.ReplayObligations) < 3 || len(contract.CanonicalOrder) < 2 || len(contract.InvalidationEdges) < 3 {
		return fmt.Errorf("replay obligations, canonical order, and invalidation edges are required")
	}
	return nil
}

func LoadBundle(path string) (Bundle, string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Bundle{}, "", err
	}
	var bundle Bundle
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return Bundle{}, "", err
	}
	if err := ValidateBundle(bundle); err != nil {
		return Bundle{}, "", err
	}
	return bundle, DigestBytes(contents), nil
}

func ValidateBundle(bundle Bundle) error {
	if bundle.Schema != "gooo/content-addressed-proof-reuse/fixture-bundle/v1" || bundle.BundleID == "" || bundle.HashAlgorithm != "sha256" {
		return fmt.Errorf("invalid fixture bundle identity")
	}
	if len(bundle.ParentReceipt.Locks) != 57 {
		return fmt.Errorf("fixture parent receipt must contain exactly 57 locks")
	}
	if err := validateLocks(bundle.ParentReceipt.Locks); err != nil {
		return fmt.Errorf("parent locks: %w", err)
	}
	if len(bundle.Cases) != 9 {
		return fmt.Errorf("fixture bundle must contain exactly nine canonical cases")
	}
	counts := map[Status]int{}
	seen := map[string]bool{}
	for _, fixtureCase := range bundle.Cases {
		if fixtureCase.ID == "" || seen[fixtureCase.ID] {
			return fmt.Errorf("case id is missing or duplicated")
		}
		seen[fixtureCase.ID] = true
		counts[fixtureCase.ExpectedStatus]++
		if fixtureCase.CurrentLockSet != "parent" && fixtureCase.CurrentLockSet != "one-delta" && fixtureCase.CurrentLockSet != "all-delta" {
			return fmt.Errorf("case %q has invalid lock set %q", fixtureCase.ID, fixtureCase.CurrentLockSet)
		}
	}
	if counts[Closed] != 3 || counts[Unknown] != 3 || counts[Refuted] != 3 {
		return fmt.Errorf("fixture case distribution must be CLOSED=3 UNKNOWN=3 REFUTED=3")
	}
	return nil
}

func validateLocks(locks []Lock) error {
	seen := map[string]bool{}
	for index, lock := range locks {
		if lock.Ordinal != index+1 || lock.Coordinate == "" || lock.Digest == "" || seen[lock.Coordinate] {
			return fmt.Errorf("invalid lock at index %d", index)
		}
		seen[lock.Coordinate] = true
	}
	return nil
}
