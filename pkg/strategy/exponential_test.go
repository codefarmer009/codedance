package strategy

import (
	"testing"
)

func TestExponentialStrategy_GenerateSteps(t *testing.T) {
	strategy := NewExponentialStrategy()
	steps := strategy.GenerateSteps()

	if len(steps) == 0 {
		t.Fatal("GenerateSteps() returned empty steps")
	}

	expectedWeights := []int{1, 5, 10, 25, 50, 100}
	if len(steps) != len(expectedWeights) {
		t.Errorf("GenerateSteps() returned %d steps, want %d", len(steps), len(expectedWeights))
	}

	for i, step := range steps {
		if step.Weight != expectedWeights[i] {
			t.Errorf("Step %d weight = %d, want %d", i, step.Weight, expectedWeights[i])
		}
		if step.Pause == "" {
			t.Errorf("Step %d pause is empty", i)
		}
	}

	// Verify weights are increasing
	for i := 1; i < len(steps); i++ {
		if steps[i].Weight <= steps[i-1].Weight {
			t.Errorf("Weights not increasing: step %d weight %d <= step %d weight %d",
				i, steps[i].Weight, i-1, steps[i-1].Weight)
		}
	}

	// Last step should be 100%
	if steps[len(steps)-1].Weight != 100 {
		t.Errorf("Last step weight = %d, want 100", steps[len(steps)-1].Weight)
	}

	// First step should be smaller in exponential than linear
	if steps[0].Weight >= 5 {
		t.Errorf("First exponential step weight = %d, expected < 5", steps[0].Weight)
	}
}
