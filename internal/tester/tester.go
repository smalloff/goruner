package tester

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type TestResult struct {
	Output  []string
	Success bool
}

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

func RunTests(ctx context.Context, path string, showPassed bool, lang string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-v", "./...")
	cmd.Dir = path
	// Hide console window on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	// Capture combined output to ensure we don't miss build errors
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	var passBlocks []string
	var failBlocks []string
	var currentBlock []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		currentBlock = append(currentBlock, line)
	
		// Проверяем маркеры успеха: функциональные тесты или итог пакета
		isPass := strings.Contains(line, "--- PASS") || 
			strings.HasPrefix(trimmedLine, "ok\t") || 
			strings.HasPrefix(trimmedLine, "ok  ") ||
			trimmedLine == "PASS"
	
		// Проверяем маркеры провала
		isFail := strings.Contains(line, "--- FAIL") || 
			strings.Contains(line, "FAIL\t") || 
			strings.HasPrefix(trimmedLine, "FAIL")
	
		if isPass {
			passBlocks = append(passBlocks, strings.Join(currentBlock, "\n"))
			currentBlock = nil
		} else if isFail {
			failBlocks = append(failBlocks, strings.Join(currentBlock, "\n"))
			currentBlock = nil
		}
	}

	// Если остался текст, не привязанный к конкретным PASS/FAIL (например, ошибки компиляции)
	if len(currentBlock) > 0 {
		remaining := strings.TrimSpace(strings.Join(currentBlock, "\n"))
		if remaining != "" {
			// Если в остатке есть явные признаки успеха, не кладем в ошибки
			if !strings.Contains(remaining, "ok\t") && !strings.Contains(remaining, "PASS") {
				failBlocks = append(failBlocks, remaining)
			} else {
				passBlocks = append(passBlocks, remaining)
			}
		}
	}

	var finalOutput strings.Builder
	
	labels := map[string]map[string]string{
		"ru": {"pass": "Успешные тесты:", "fail": "Ошибки в тестах:", "all_ok": "✅ Все тесты пройдены успешно!", "none": "ℹ️ Тесты не найдены или вывод пуст."},
		"en": {"pass": "Successful tests:", "fail": "Errors in tests:", "all_ok": "✅ All tests passed successfully!", "none": "ℹ️ No tests found or output is empty."},
	}
	l := labels[lang]
	if l == nil { l = labels["en"] }
	
	if len(passBlocks) > 0 && showPassed {
		finalOutput.WriteString(l["pass"] + "\n")
		for _, b := range passBlocks {
			finalOutput.WriteString(b + "\n")
		}
		finalOutput.WriteString("\n")
	}
	
	if len(failBlocks) == 0 {
		if len(passBlocks) > 0 {
			finalOutput.WriteString(l["all_ok"] + "\n")
		} else {
			finalOutput.WriteString(l["none"] + "\n")
		}
	} else {
		finalOutput.WriteString(l["fail"] + "\n\n")
		for _, b := range failBlocks {
			if strings.TrimSpace(b) != "" {
				finalOutput.WriteString(b + "\n")
			}
		}
	}

	return finalOutput.String(), nil
}
