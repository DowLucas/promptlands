package worldgen

import (
	"testing"
)

func TestGeneratorDeterminism(t *testing.T) {
	seed := int64(12345)
	size := 20

	// Generate two worlds with the same seed
	gen1 := NewWorldGenerator(seed)
	tiles1 := gen1.Generate(size)

	gen2 := NewWorldGenerator(seed)
	tiles2 := gen2.Generate(size)

	// Verify they are identical
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if tiles1[y][x].Terrain != tiles2[y][x].Terrain {
				t.Errorf("Terrain mismatch at (%d, %d): %v != %v",
					x, y, tiles1[y][x].Terrain, tiles2[y][x].Terrain)
			}
		}
	}
}

func TestGeneratorDifferentSeeds(t *testing.T) {
	size := 20

	gen1 := NewWorldGenerator(12345)
	tiles1 := gen1.Generate(size)

	gen2 := NewWorldGenerator(54321)
	tiles2 := gen2.Generate(size)

	// Count differences
	differences := 0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if tiles1[y][x].Terrain != tiles2[y][x].Terrain {
				differences++
			}
		}
	}

	// With different seeds, there should be significant differences
	if differences == 0 {
		t.Error("Expected different worlds with different seeds, but they are identical")
	}

	minExpectedDifferences := size * size / 10 // At least 10% different
	if differences < minExpectedDifferences {
		t.Errorf("Expected at least %d differences, got %d", minExpectedDifferences, differences)
	}
}

func TestBiomeDistribution(t *testing.T) {
	// Use a larger map to ensure all biome types appear
	// With default noise frequency, small maps may not have enough variation
	seed := int64(42)
	size := 100

	gen := NewWorldGenerator(seed)
	tiles := gen.Generate(size)

	// Count terrain types
	counts := make(map[TerrainType]int)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			counts[tiles[y][x].Terrain]++
		}
	}

	total := size * size

	// Verify all terrain types are present
	terrainTypes := []TerrainType{
		TerrainPlains,
		TerrainForest,
		TerrainMountain,
		TerrainWater,
	}

	for _, terrain := range terrainTypes {
		count := counts[terrain]
		if count == 0 {
			t.Errorf("Expected some %s tiles, got none", terrain)
		}
		t.Logf("%s: %d tiles (%.1f%%)", terrain, count, float64(count)*100/float64(total))
	}
}

func TestNoiseNormalization(t *testing.T) {
	gen := NewNoiseGenerator(12345)

	// Test many points
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			val := gen.Eval2D(float64(x), float64(y))
			if val < 0 || val > 1 {
				t.Errorf("Noise value at (%d, %d) out of range [0, 1]: %f", x, y, val)
			}
		}
	}
}

func TestOctaveNoiseNormalization(t *testing.T) {
	gen := NewNoiseGenerator(12345)

	// Test octave noise normalization
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			val := gen.Octave2D(float64(x), float64(y), 4, 0.02, 0.5)
			if val < 0 || val > 1 {
				t.Errorf("Octave noise value at (%d, %d) out of range [0, 1]: %f", x, y, val)
			}
		}
	}
}

func TestDetermineBiome(t *testing.T) {
	thresholds := DefaultThresholds()

	tests := []struct {
		elevation float64
		moisture  float64
		expected  TerrainType
	}{
		{0.1, 0.5, TerrainWater},     // Low elevation = water
		{0.2, 0.8, TerrainWater},     // Low elevation = water regardless of moisture
		{0.7, 0.5, TerrainMountain},  // High elevation = mountain
		{0.9, 0.9, TerrainMountain},  // High elevation = mountain regardless of moisture
		{0.5, 0.7, TerrainForest},    // Mid elevation, high moisture = forest
		{0.5, 0.4, TerrainPlains},    // Mid elevation, low moisture = plains
		{0.4, 0.5, TerrainPlains},    // Mid elevation, mid moisture = plains
	}

	for _, tc := range tests {
		result := DetermineBiome(tc.elevation, tc.moisture, thresholds)
		if result != tc.expected {
			t.Errorf("DetermineBiome(%.1f, %.1f) = %v, expected %v",
				tc.elevation, tc.moisture, result, tc.expected)
		}
	}
}

func TestIsPassable(t *testing.T) {
	tests := []struct {
		terrain  TerrainType
		expected bool
	}{
		{TerrainPlains, true},
		{TerrainForest, true},
		{TerrainMountain, false},
		{TerrainWater, false},
	}

	for _, tc := range tests {
		result := IsPassable(tc.terrain)
		if result != tc.expected {
			t.Errorf("IsPassable(%v) = %v, expected %v", tc.terrain, result, tc.expected)
		}
	}
}

func TestIsPassableString(t *testing.T) {
	tests := []struct {
		terrain  string
		expected bool
	}{
		{"plains", true},
		{"forest", true},
		{"mountain", false},
		{"water", false},
	}

	for _, tc := range tests {
		result := IsPassableString(tc.terrain)
		if result != tc.expected {
			t.Errorf("IsPassableString(%v) = %v, expected %v", tc.terrain, result, tc.expected)
		}
	}
}

func TestGeneratorSize(t *testing.T) {
	sizes := []int{10, 20, 50, 100}

	for _, size := range sizes {
		gen := NewWorldGenerator(12345)
		tiles := gen.Generate(size)

		if len(tiles) != size {
			t.Errorf("Expected %d rows, got %d", size, len(tiles))
		}

		for y, row := range tiles {
			if len(row) != size {
				t.Errorf("Expected %d columns in row %d, got %d", size, y, len(row))
			}
		}
	}
}

func TestTilePositions(t *testing.T) {
	size := 20
	gen := NewWorldGenerator(12345)
	tiles := gen.Generate(size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			tile := tiles[y][x]
			if tile.X != x || tile.Y != y {
				t.Errorf("Tile at [%d][%d] has position (%d, %d)",
					y, x, tile.X, tile.Y)
			}
		}
	}
}

func TestCustomConfig(t *testing.T) {
	config := GeneratorConfig{
		ElevationOctaves:     2,
		ElevationFrequency:   0.05,
		ElevationPersistence: 0.6,
		MoistureOctaves:      2,
		MoistureFrequency:    0.05,
		MoisturePersistence:  0.6,
		Thresholds: BiomeThresholds{
			WaterMax:     0.2,
			MountainMin:  0.8,
			ForestMinMoi: 0.5,
		},
	}

	gen := NewWorldGeneratorWithConfig(12345, config)
	tiles := gen.Generate(20)

	// Just verify it generates without errors
	if len(tiles) != 20 {
		t.Errorf("Expected 20 rows, got %d", len(tiles))
	}
}

// Benchmark for performance testing
func BenchmarkGenerate(b *testing.B) {
	gen := NewWorldGenerator(12345)
	size := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(size)
	}
}
