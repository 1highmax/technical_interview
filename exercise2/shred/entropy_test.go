package shred

import (
	"crypto/rand"
	"math"
	"math/big"
	"testing"

	"gonum.org/v1/gonum/stat/distuv"
)

// GenerateRandomData generates a slice of random numbers
func GenerateRandomData(n int, max int64) ([]int64, error) {
	data := make([]int64, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(max))
		if err != nil {
			return nil, err
		}
		data[i] = num.Int64()
	}
	return data, nil
}

// FrequencyTest performs a frequency test on the data
func FrequencyTest(data []int64, max int64) map[int64]int {
	frequency := make(map[int64]int)
	for _, num := range data {
		frequency[num]++
	}
	return frequency
}

// ChiSquareTest performs a Chi-square test on the data
func ChiSquareTest(data []int64, max int64) float64 {
	n := len(data)
	expected := float64(n) / float64(max)
	frequency := FrequencyTest(data, max)

	chiSquare := 0.0
	for i := int64(0); i < max; i++ {
		observed := float64(frequency[i])
		chiSquare += math.Pow(observed-expected, 2) / expected
	}
	return chiSquare
}

func TestEntropy(t *testing.T) {
	n := 10000 // number of samples
	max := int64(100)

	// Generate random data
	data, err := GenerateRandomData(n, max)
	if err != nil {
		t.Fatalf("Error generating random data: %v\n", err)
	}

	// Perform frequency test
	frequency := FrequencyTest(data, max)
	t.Logf("Frequency Test Results:")
	for num, count := range frequency {
		t.Logf("Number %d: %d times\n", num, count)
	}

	// Perform Chi-square test
	chiSquare := ChiSquareTest(data, max)
	t.Logf("Chi-square value: %f\n", chiSquare)

	// Calculate critical value
	df := float64(max - 1)
	alpha := 0.99
	chiSq := distuv.ChiSquared{K: df}
	criticalValue := chiSq.Quantile(alpha)

	// Validate Chi-square value
	if chiSquare > criticalValue {
		t.Errorf("Chi-square value is too high: %f, critical value: %f", chiSquare, criticalValue)
	}
}
