// Run from the book root: go run scripts/verify-go-examples.go [chapter.md ...]
// Each source block is extracted afresh; checked-in test overlays never replace
// the book's code. An unregistered or newly added block fails the verification.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type manifest struct {
	GoVersion string            `json:"go_version"`
	Requires  map[string]string `json:"requires"`
	Groups    []group           `json:"groups"`
}
type group struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Blocks []block  `json:"blocks"`
	Tests  []string `json:"tests"`
}
type block struct {
	Number int    `json:"number"`
	File   string `json:"file"`
	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}

func main() {
	if err := verify(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Go example verification failed:", err)
		os.Exit(1)
	}
}

func verify(requested []string) error {
	if _, err := os.Stat("SUMMARY.md"); err != nil {
		return fmt.Errorf("run from the book repository root: %w", err)
	}
	raw, err := os.ReadFile("scripts/go-examples/manifest.json")
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var config manifest
	if err := decoder.Decode(&config); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("manifest has trailing data")
	}
	if config.GoVersion == "" {
		return fmt.Errorf("manifest needs a Go version")
	}
	wanted := make(map[string]bool)
	for _, path := range requested {
		wanted[filepath.ToSlash(filepath.Clean(path))] = true
	}
	sources := make(map[string][]string)
	covered := make(map[string]map[int]bool)
	names := make(map[string]bool)
	var selected []group
	for _, current := range config.Groups {
		if !safeRelative(current.Source) || filepath.ToSlash(filepath.Clean(current.Source)) != current.Source ||
			!safeRelative(current.Name) || filepath.Base(current.Name) != current.Name || names[current.Name] {
			return fmt.Errorf("invalid or duplicate group %q", current.Name)
		}
		names[current.Name] = true
		if len(wanted) != 0 && !wanted[current.Source] {
			continue
		}
		if _, ok := sources[current.Source]; !ok {
			data, err := os.ReadFile(current.Source)
			if err != nil {
				return err
			}
			snippets, err := goBlocks(string(data))
			if err != nil {
				return fmt.Errorf("%s: %w", current.Source, err)
			}
			sources[current.Source], covered[current.Source] = snippets, make(map[int]bool)
		}
		for _, spec := range current.Blocks {
			if spec.Number <= 0 || spec.Number > len(sources[current.Source]) || covered[current.Source][spec.Number] {
				return fmt.Errorf("%s: invalid/duplicate block %d", current.Source, spec.Number)
			}
			covered[current.Source][spec.Number] = true
		}
		selected = append(selected, current)
	}
	for path := range wanted {
		if _, ok := sources[path]; !ok {
			return fmt.Errorf("no verification mapping for %s", path)
		}
	}
	for path, snippets := range sources {
		if len(snippets) == 0 || len(covered[path]) != len(snippets) {
			return fmt.Errorf("%s: mapped %d of %d Go blocks; update the manifest", path, len(covered[path]), len(snippets))
		}
	}
	if len(selected) == 0 {
		return fmt.Errorf("no Go examples selected")
	}
	work, err := os.MkdirTemp("", "payment-book-go-examples-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work) // Only the explicit directory created above.
	mod := "module example.com/payment-book-examples\n\ngo " + config.GoVersion + "\n"
	for module, version := range config.Requires {
		mod += "require " + module + " " + version + "\n"
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), []byte(mod), 0600); err != nil {
		return err
	}
	count := 0
	for _, current := range selected {
		directory := filepath.Join(work, current.Name)
		if err := os.Mkdir(directory, 0700); err != nil {
			return err
		}
		files := make(map[string]bool)
		write := func(name string, data []byte) error {
			if !safeRelative(name) || filepath.Base(name) != name || !strings.HasSuffix(name, ".go") || files[name] {
				return fmt.Errorf("invalid/duplicate generated file %q", name)
			}
			files[name] = true
			formatted, err := format.Source(data)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", current.Name, name, err)
			}
			return os.WriteFile(filepath.Join(directory, name), formatted, 0600)
		}
		for _, spec := range current.Blocks {
			data := spec.Prefix + sources[current.Source][spec.Number-1] + spec.Suffix
			if err := write(spec.File, []byte(data)); err != nil {
				return err
			}
			count++
		}
		for _, testPath := range current.Tests {
			if !safeRelative(testPath) {
				return fmt.Errorf("invalid test overlay %q", testPath)
			}
			data, err := os.ReadFile(filepath.Join("scripts/go-examples", testPath))
			if err != nil {
				return err
			}
			if err := write(filepath.Base(testPath), data); err != nil {
				return err
			}
		}
		fmt.Printf("EXTRACT %s: %d Go blocks → %s\n", current.Source, len(current.Blocks), current.Name)
	}
	for _, args := range [][]string{{"mod", "tidy"}, {"test", "-count=1", "-race", "-p", "2", "./..."}, {"vet", "-p", "2", "./..."}} {
		command := exec.Command("go", args...)
		command.Dir, command.Stdout, command.Stderr = work, os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
		}
	}
	fmt.Printf("PASS %d chapters, %d Go blocks, %d packages (compile/race/vet)\n", len(sources), count, len(selected))
	return nil
}

func safeRelative(path string) bool {
	clean := filepath.Clean(path)
	return path != "" && !filepath.IsAbs(path) && clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func goBlocks(markdown string) ([]string, error) {
	var snippets []string
	var lines []string
	inside := false
	for number, line := range strings.Split(markdown, "\n") {
		if !inside && line == "```go" {
			inside = true
			lines = nil
			continue
		}
		if inside && line == "```" {
			snippets = append(snippets, strings.Join(lines, "\n")+"\n")
			inside = false
			continue
		}
		if inside && strings.HasPrefix(line, "```") {
			return nil, fmt.Errorf("unexpected fence at line %s", strconv.Itoa(number+1))
		}
		if inside {
			lines = append(lines, line)
		}
	}
	if inside {
		return nil, fmt.Errorf("unclosed Go fence")
	}
	return snippets, nil
}
