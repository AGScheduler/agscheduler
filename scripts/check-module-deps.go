// Command check-module-deps enforces the dependency boundaries of the split
// modules. It is intentionally dependency-free so it can run before any module
// graph is loaded.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	modulesFile    = "scripts/modules.txt"
	moduleDepsFile = "scripts/module-deps.txt"
)

type editJSON struct {
	Module struct {
		Path string
	}
	Require []struct {
		Path     string
		Version  string
		Indirect bool
	}
}

type workJSON struct {
	Use []struct {
		DiskPath string
	}
	Replace []struct {
		Old struct {
			Path    string
			Version string
		}
		New struct {
			Path    string
			Version string
		}
	}
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	modules, err := readLines(filepath.Join(root, modulesFile))
	if err != nil {
		fatal(err)
	}
	sort.Strings(modules)

	allowedByModule, err := readModuleDeps(filepath.Join(root, moduleDepsFile))
	if err != nil {
		fatal(err)
	}

	registered := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		registered[module] = struct{}{}
		allowed, ok := allowedByModule[module]
		if !ok {
			fatal(fmt.Errorf("%s: missing dependency allowlist; add it to %s", module, moduleDepsFile))
		}
		checkDirectRequirements(root, module, allowed)
	}
	for module := range allowedByModule {
		if _, ok := registered[module]; !ok {
			fatal(fmt.Errorf("%s: dependency allowlist refers to unregistered module %s", moduleDepsFile, module))
		}
	}
	checkCoreRuntime(root)
	checkWorkspace(root, modules)
	fmt.Println("module dependency boundaries are valid")
}

func checkDirectRequirements(root, module string, allowedDirect []string) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = filepath.Join(root, module)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		fatal(fmt.Errorf("%s: go mod edit: %w", module, err))
	}

	var mod editJSON
	if err := json.Unmarshal(output, &mod); err != nil {
		fatal(fmt.Errorf("%s: decode go.mod: %w", module, err))
	}

	allowed := make(map[string]struct{}, len(allowedDirect))
	for _, path := range allowedDirect {
		allowed[path] = struct{}{}
	}

	for _, requirement := range mod.Require {
		if requirement.Indirect || strings.HasPrefix(requirement.Path, "github.com/agscheduler/agscheduler") {
			continue
		}
		if _, ok := allowed[requirement.Path]; !ok {
			fatal(fmt.Errorf("%s: unexpected direct dependency %s", module, requirement.Path))
		}
	}
}

func readModuleDeps(path string) (map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		module := fields[0]
		if _, exists := result[module]; exists {
			return nil, fmt.Errorf("%s:%d: duplicate module %s", path, lineNumber, module)
		}
		result[module] = append([]string(nil), fields[1:]...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, exists := seen[line]; exists {
			return nil, fmt.Errorf("%s:%d: duplicate module %s", path, lineNumber, line)
		}
		seen[line] = struct{}{}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func checkWorkspace(root string, modules []string) {
	cmd := exec.Command("go", "work", "edit", "-json")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		fatal(fmt.Errorf("go work edit: %w", err))
	}

	var work workJSON
	if err := json.Unmarshal(output, &work); err != nil {
		fatal(fmt.Errorf("decode go.work: %w", err))
	}

	rootModule, err := currentModulePath(root)
	if err != nil {
		fatal(err)
	}
	currentVersion, err := currentModuleVersion(root)
	if err != nil {
		fatal(err)
	}

	registered := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		registered[normalizeModuleDir(module)] = struct{}{}
	}
	for _, use := range work.Use {
		if _, ok := registered[normalizeModuleDir(use.DiskPath)]; !ok {
			fatal(fmt.Errorf("go.work uses unregistered module %s", use.DiskPath))
		}
	}

	replaced := make(map[string]string, len(work.Replace))
	for _, replacement := range work.Replace {
		key := replacement.Old.Path + "@" + replacement.Old.Version
		replaced[key] = normalizeModuleDir(replacement.New.Path)
	}

	for _, module := range modules {
		module = normalizeModuleDir(module)
		modulePath := rootModule
		if module != "." {
			modulePath += "/" + module
		}

		replacementDir, ok := replaced[modulePath+"@v"+currentVersion]
		if !ok {
			fatal(fmt.Errorf("go.work is missing replacement for %s@v%s", modulePath, currentVersion))
		}
		if replacementDir != module {
			fatal(fmt.Errorf("go.work replacement for %s points to %s, want %s", modulePath, replacementDir, module))
		}
	}
}

func currentModulePath(root string) (string, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go mod edit: %w", err)
	}

	var mod editJSON
	if err := json.Unmarshal(output, &mod); err != nil {
		return "", fmt.Errorf("decode go.mod: %w", err)
	}
	if mod.Module.Path == "" {
		return "", fmt.Errorf("root go.mod has no module path")
	}
	return mod.Module.Path, nil
}

func currentModuleVersion(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, "version.go"))
	if err != nil {
		return "", err
	}
	match := regexp.MustCompile(`const Version = "([^"]+)"`).FindSubmatch(content)
	if len(match) != 2 {
		return "", fmt.Errorf("version.go does not declare const Version")
	}
	return string(match[1]), nil
}

func normalizeModuleDir(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "./")
	if path == "" {
		return "."
	}
	return path
}

func checkCoreRuntime(root string) {
	cmd := exec.Command("go", "list", "-deps", ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		fatal(fmt.Errorf("core runtime dependency check: %s: %w", strings.TrimSpace(stderr.String()), err))
	}

	allowed := map[string]struct{}{
		"github.com/google/uuid":          {},
		"github.com/gorhill/cronexpr":     {},
		"github.com/sony/sonyflake":       {},
		"github.com/sony/sonyflake/types": {},
	}
	for _, dependency := range strings.Fields(string(output)) {
		firstComponent := dependency
		if index := strings.IndexByte(firstComponent, '/'); index >= 0 {
			firstComponent = firstComponent[:index]
		}
		if !strings.Contains(firstComponent, ".") ||
			strings.HasPrefix(dependency, "vendor/") ||
			strings.HasPrefix(dependency, "github.com/agscheduler/agscheduler") {
			continue
		}
		if _, ok := allowed[dependency]; !ok {
			fatal(fmt.Errorf("core runtime has unexpected dependency %s", dependency))
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
