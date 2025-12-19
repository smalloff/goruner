package tester

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Labels contains localized strings for output headers and summaries.
var Labels = map[string]map[string]string{
	"ru": {
		"pass":   "Успешные тесты:",
		"fail":   "Ошибки в тестах:",
		"all_ok": "✅ Все тесты пройдены успешно!",
		"none":   "ℹ️ Тесты не найдены или вывод пуст.",
	},
	"en": {
		"pass":   "Successful tests:",
		"fail":   "Errors in tests:",
		"all_ok": "✅ All tests passed successfully!",
		"none":   "ℹ️ No tests found or output is empty.",
	},
}

// DiscoverTests scans the root directory for packages containing test files.
func DiscoverTests(root string) ([]string, error) {
	var packages []string
	set := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			dir := filepath.Dir(path)
			rel, _ := filepath.Rel(root, dir)
			if rel == "." {
				set["./"] = true
			} else {
				set["./"+filepath.ToSlash(rel)] = true
			}
		}
		return nil
	})

	for pkg := range set {
		packages = append(packages, pkg)
	}
	return packages, err
}

// RunTests executes 'go test' and returns a formatted result string.
func RunTests(ctx context.Context, path string, showPassed bool, lang string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-v", "./...")
	cmd.Dir = path
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, _ := cmd.CombinedOutput()
	return ParseTestOutput(string(output), showPassed, lang), nil
}

// ParseTestOutput parses raw 'go test' output and groups results by status.
func ParseTestOutput(rawOutput string, showPassed bool, lang string) string {
	lines := strings.Split(rawOutput, "\n")
	l, ok := Labels[lang]
	if !ok {
		l = Labels["en"]
	}

	var passBlocks, failBlocks []string
	var currentBlock []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		currentBlock = append(currentBlock, line)

		// Check for successful test or package markers
		isPass := strings.Contains(line, "--- PASS") || 
		          strings.HasPrefix(trimmed, "ok\t") || 
		          strings.HasPrefix(trimmed, "ok  ") ||
		          trimmed == "PASS"
		
		// Check for failure markers or panics
		isFail := strings.Contains(line, "--- FAIL") || 
		          strings.Contains(line, "FAIL\t") || 
		          strings.HasPrefix(trimmed, "FAIL") ||
		          strings.HasPrefix(trimmed, "panic:")

		if isPass {
			passBlocks = append(passBlocks, strings.Join(currentBlock, "\n"))
			currentBlock = nil
		} else if isFail {
			failBlocks = append(failBlocks, strings.Join(currentBlock, "\n"))
			currentBlock = nil
		}
	}

	// Handle remaining output (e.g., build errors)
	if len(currentBlock) > 0 {
		content := strings.TrimSpace(strings.Join(currentBlock, "\n"))
		if content != "" {
			if strings.Contains(content, "PASS") || strings.Contains(content, "ok\t") || strings.Contains(content, "ok  ") {
				passBlocks = append(passBlocks, content)
			} else {
				failBlocks = append(failBlocks, content)
			}
		}
	}

	var res strings.Builder
	if showPassed && len(passBlocks) > 0 {
		res.WriteString(fmt.Sprintf("%s\n%s\n\n", l["pass"], strings.Join(passBlocks, "\n")))
	}

	if len(failBlocks) > 0 {
		res.WriteString(fmt.Sprintf("%s\n\n%s\n", l["fail"], strings.Join(failBlocks, "\n")))
	} else if len(passBlocks) > 0 {
		res.WriteString(l["all_ok"] + "\n")
	} else {
		res.WriteString(l["none"] + "\n")
	}

	return res.String()
}
