package worldgen

import (
	"github.com/ojrac/opensimplex-go"
)

// NoiseGenerator wraps OpenSimplex noise with seed support
type NoiseGenerator struct {
	noise opensimplex.Noise
	seed  int64
}

// NewNoiseGenerator creates a new noise generator with the given seed
func NewNoiseGenerator(seed int64) *NoiseGenerator {
	return &NoiseGenerator{
		noise: opensimplex.New(seed),
		seed:  seed,
	}
}

// Eval2D returns the noise value at (x, y), normalized to [0, 1]
func (n *NoiseGenerator) Eval2D(x, y float64) float64 {
	// OpenSimplex returns values in [-1, 1], normalize to [0, 1]
	return (n.noise.Eval2(x, y) + 1) / 2
}

// Octave2D generates fractal noise using multiple octaves
// octaves: number of noise layers to combine
// frequency: base frequency (lower = larger features)
// persistence: amplitude decrease per octave (0.5 is typical)
func (n *NoiseGenerator) Octave2D(x, y float64, octaves int, frequency, persistence float64) float64 {
	var total float64
	var maxValue float64
	amplitude := 1.0
	freq := frequency

	for i := 0; i < octaves; i++ {
		total += n.Eval2D(x*freq, y*freq) * amplitude
		maxValue += amplitude
		amplitude *= persistence
		freq *= 2
	}

	// Normalize to [0, 1]
	return total / maxValue
}

// Seed returns the generator's seed
func (n *NoiseGenerator) Seed() int64 {
	return n.seed
}
