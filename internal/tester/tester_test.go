package tester

import (
	"strings"
	"testing"
)

func TestParseTestOutput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		showPassed bool
		lang       string
		wantIn     []string
		mustNotIn  []string
	}{
		{
			name:       "All passed",
			input:      "=== RUN TestA\n--- PASS: TestA (0.00s)\nok  pkg 0.001s",
			showPassed: true,
			lang:       "en",
			wantIn:     []string{"Successful tests:", "--- PASS", "All tests passed successfully!"},
			mustNotIn:  []string{"Errors in tests:"},
		},
		{
			name:       "Failure with details",
			input:      "=== RUN TestB\n    b_test.go:10: error!\n--- FAIL: TestB (0.00s)\nFAIL",
			showPassed: false,
			lang:       "ru",
			wantIn:     []string{"Ошибки в тестах:", "--- FAIL", "error!"},
			mustNotIn:  []string{"Успешные тесты:"},
		},
		{
			name:       "Compilation error",
			input:      "# pkg\n./main.go:5:2: undefined: x",
			showPassed: true,
			lang:       "en",
			wantIn:     []string{"Errors in tests:", "undefined: x"},
			mustNotIn:  []string{"All tests passed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTestOutput(tt.input, tt.showPassed, tt.lang)
			for _, want := range tt.wantIn {
				if !strings.Contains(got, want) {
					t.Errorf("[%s] Result should contain %q, but got:\n%s", tt.name, want, got)
				}
			}
			for _, notWant := range tt.mustNotIn {
				if strings.Contains(got, notWant) {
					t.Errorf("[%s] Result should NOT contain %q, but it does", tt.name, notWant)
				}
			}
		})
	}
}
