package reuse

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestString(value string) string { return DigestBytes([]byte(value)) }

func DigestCanonical(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(encoded), nil
}

func FileDigest(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return DigestBytes(contents), nil
}

func CanonicalLockMaterial(locks []Lock, dependencyDigest string) string {
	ordered := append([]Lock(nil), locks...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Ordinal != ordered[j].Ordinal {
			return ordered[i].Ordinal < ordered[j].Ordinal
		}
		return ordered[i].Coordinate < ordered[j].Coordinate
	})
	var builder strings.Builder
	builder.WriteString("gooo/content-addressed-proof-reuse/semantic-root/v1|")
	builder.WriteString(dependencyDigest)
	for _, lock := range ordered {
		fmt.Fprintf(&builder, "|%03d|%s|%s", lock.Ordinal, lock.Coordinate, lock.Digest)
	}
	return builder.String()
}

func ComputeSemanticRoot(locks []Lock, dependencyDigest string) string {
	return DigestString(CanonicalLockMaterial(locks, dependencyDigest))
}

func ComputeInventory(root string) (Inventory, error) {
	var inventory Inventory
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "README.md" {
			inventory.RootReadmeExcluded = true
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		inventory.RegularFiles++
		inventory.TreeBytes += int64(len(contents))
		switch {
		case strings.HasSuffix(relative, ".go"):
			inventory.GoFiles++
			inventory.GoLines += int64(countLines(contents))
			if strings.HasSuffix(relative, "_test.go") {
				inventory.TestFiles++
			}
		case strings.HasSuffix(relative, ".gooo"):
			inventory.GoooFiles++
			inventory.GoooLines += int64(countLines(contents))
		case strings.HasPrefix(relative, "fixtures/"):
			inventory.FixtureFiles++
		case strings.HasPrefix(relative, "contracts/"):
			inventory.ContractFiles++
		}
		return nil
	})
	return inventory, err
}

func countLines(contents []byte) int {
	if len(contents) == 0 {
		return 0
	}
	lines := 1
	for _, character := range contents {
		if character == '\n' {
			lines++
		}
	}
	return lines
}
