package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/gooo-content-addressed-proof-reuse/internal/fixture"
	"github.com/kimjooyoon/gooo-content-addressed-proof-reuse/internal/reuse"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: gooo-content-addressed-proof-reuse <conformance|suite>")
	}
	switch os.Args[1] {
	case "conformance":
		runConformance(os.Args[2:])
	case "suite":
		runSuite(os.Args[2:])
	default:
		fail("unknown command %q", os.Args[1])
	}
}

func runConformance(arguments []string) {
	flags := flag.NewFlagSet("conformance", flag.ExitOnError)
	contractPath := flags.String("contract", ".gooo/content-addressed-proof-reuse.gooo", "path to the .gooo contract")
	bundlePath := flags.String("bundle", "fixtures/fixture-bundle-v1.json", "path to the immutable fixture bundle")
	_ = flags.Parse(arguments)
	contract, _, err := reuse.ParseContract(*contractPath)
	if err != nil {
		fail("parse contract: %v", err)
	}
	bundle, _, err := reuse.LoadBundle(*bundlePath)
	if err != nil {
		fail("load bundle: %v", err)
	}
	if err := reuse.ValidateConformance(contract, bundle); err != nil {
		fail("conformance: %v", err)
	}
	fmt.Println("conformance: CLOSED=3 UNKNOWN=3 REFUTED=3; 57-lock pairs: reused=57 selected=0 and reused=56 selected=1")
}

func runSuite(arguments []string) {
	flags := flag.NewFlagSet("suite", flag.ExitOnError)
	contractPath := flags.String("contract", ".gooo/content-addressed-proof-reuse.gooo", "path to the .gooo contract")
	bundlePath := flags.String("bundle", "fixtures/fixture-bundle-v1.json", "path to the immutable fixture bundle")
	repoRoot := flags.String("repo-root", ".", "repository root for exact inventory")
	outputDir := flags.String("output-dir", "", "caller-owned output directory")
	_ = flags.Parse(arguments)
	if *outputDir == "" {
		fail("suite requires --output-dir beneath a caller-owned temporary directory")
	}
	contract, contractDigest, err := reuse.ParseContract(*contractPath)
	if err != nil {
		fail("parse contract: %v", err)
	}
	bundle, fixtureDigest, err := reuse.LoadBundle(*bundlePath)
	if err != nil {
		fail("load bundle: %v", err)
	}
	server := fixture.NewServer(bundle)
	defer server.Close()
	evidence, err := reuse.RunSuite(contract, bundle, server, server)
	if err != nil {
		fail("suite: %v", err)
	}
	evidence.ContractDigest = contractDigest
	evidence.FixtureDigest = fixtureDigest
	evidence.Inventory, err = reuse.ComputeInventory(*repoRoot)
	if err != nil {
		fail("inventory: %v", err)
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fail("create caller-owned output directory: %v", err)
	}
	if err := writeJSON(filepath.Join(*outputDir, "suite-evidence.json"), evidence); err != nil {
		fail("write suite evidence: %v", err)
	}
	if err := writeJSON(filepath.Join(*outputDir, "metrics-vector.json"), evidence.Metrics); err != nil {
		fail("write metrics vector: %v", err)
	}
	for _, run := range evidence.Runs {
		name := fmt.Sprintf("receipt-%s-%s.json", run.Engine, safeName(run.CaseID))
		if err := writeJSON(filepath.Join(*outputDir, name), run); err != nil {
			fail("write %s: %v", name, err)
		}
	}
	for _, comparison := range evidence.CanonicalComparisons {
		name := fmt.Sprintf("canonical-comparison-%s.json", safeName(comparison.CaseID))
		if err := writeJSON(filepath.Join(*outputDir, name), comparison); err != nil {
			fail("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(*outputDir, "human-report.md"), []byte(reuse.RenderHumanReport(evidence)), 0o644); err != nil {
		fail("write human report: %v", err)
	}
	fmt.Printf("suite: cases=%d output=%s\n", len(evidence.Cases), *outputDir)
}

func writeJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(path, encoded, 0o644)
}

func safeName(value string) string {
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(1)
}
