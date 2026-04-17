package numbat

import (
	"sync"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}
	defer ctx.Free()
	if ctx.wrapper == nil {
		t.Error("Context wrapper is nil after initialization")
	}
}
func TestInterpret_SimpleCalculation(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	res, err := ctx.Interpret("1 + 1")
	if err != nil {
		t.Fatalf("Interpret() returned error: %v", err)
	}
	if res.StringOutput == "" {
		t.Error("Expected non-empty string output")
	}
}
func TestInterpret_UnitConversion(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	res, err := ctx.Interpret("120 km/h -> m/s")
	if err != nil {
		t.Fatalf("Interpret() returned error: %v", err)
	}
	if res.StringOutput == "" {
		t.Error("Expected non-empty string output")
	}
	if !res.IsQuantity {
		t.Error("Expected IsQuantity to be true for unit conversion")
	}
	// 120 km/h = 33.333... m/s
	expectedValue := 33.333333333333336
	if res.Value != expectedValue {
		t.Errorf("Expected value %f, got %f", expectedValue, res.Value)
	}
	if res.Unit != "m/s" {
		t.Errorf("Expected unit 'm/s', got '%s'", res.Unit)
	}
}
func TestInterpret_ArithmeticOperations(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	tests := []struct {
		name     string
		code     string
		expected float64
	}{
		{"addition", "10 + 5", 15},
		{"subtraction", "10 - 3", 7},
		{"multiplication", "6 * 7", 42},
		{"division", "100 / 4", 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ctx.Interpret(tt.code)
			if err != nil {
				t.Fatalf("Interpret(%q) returned error: %v", tt.code, err)
			}
			if res.Value != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, res.Value)
			}
		})
	}
}
func TestInterpret_ComplexExpression(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	res, err := ctx.Interpret("(10 + 5) * 2 - 3")
	if err != nil {
		t.Fatalf("Interpret() returned error: %v", err)
	}
	expected := 27.0
	if res.Value != expected {
		t.Errorf("Expected %f, got %f", expected, res.Value)
	}
}
func TestInterpret_InvalidExpression(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	_, err := ctx.Interpret("invalid syntax here!!!")
	if err == nil {
		t.Error("Expected error for invalid syntax, got nil")
	}
}
func TestInterpret_EmptyString(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	// Empty string may be valid or error depending on Numbat behavior
	// Just ensure it doesn't crash
	_, _ = ctx.Interpret("")
}
func TestSetVariable(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	err := ctx.SetVariable("test_var", 42.0, "")
	if err != nil {
		t.Fatalf("SetVariable() returned error: %v", err)
	}
	// Verify the variable was set by using it
	res, err := ctx.Interpret("test_var * 2")
	if err != nil {
		t.Fatalf("Interpret() returned error: %v", err)
	}
	if res.Value != 84.0 {
		t.Errorf("Expected 84.0, got %f", res.Value)
	}
}
func TestSetVariable_WithUnit(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	err := ctx.SetVariable("distance", 100.0, "meters")
	if err != nil {
		t.Fatalf("SetVariable() returned error: %v", err)
	}
	res, err := ctx.Interpret("distance -> km")
	if err != nil {
		t.Fatalf("Interpret() returned error: %v", err)
	}
	if !res.IsQuantity {
		t.Error("Expected IsQuantity to be true")
	}
	if res.Unit != "km" {
		t.Errorf("Expected unit 'km', got '%s'", res.Unit)
	}
}
func TestSetVariable_InvalidName(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	// Test with an invalid variable name
	err := ctx.SetVariable("123invalid", 42.0, "")
	if err == nil {
		t.Error("Expected error for invalid variable name, got nil")
	}
}
func TestInterpret_ConcurrentAccess(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			expr := "2 + 2"
			_, err := ctx.Interpret(expr)
			if err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("Concurrent interpretation error: %v", err)
		}
	}
	if errorCount > 0 {
		t.Errorf("Got %d errors during concurrent access", errorCount)
	}
}
func TestInterpret_ConcurrentWithDifferentExpressions(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	expressions := []string{
		"1 + 1",
		"10 * 10",
		"100 / 4",
		"50 - 25",
		"2^8",
	}
	var wg sync.WaitGroup
	results := make(map[int]Result)
	var mu sync.Mutex
	for i, expr := range expressions {
		wg.Add(1)
		go func(id int, code string) {
			defer wg.Done()
			res, err := ctx.Interpret(code)
			if err != nil {
				t.Logf("Error in goroutine %d: %v", id, err)
				return
			}
			mu.Lock()
			results[id] = res
			mu.Unlock()
		}(i, expr)
	}
	wg.Wait()
	if len(results) != len(expressions) {
		t.Errorf("Expected %d results, got %d", len(expressions), len(results))
	}
}
func TestInterpret_UnitsAndDimensions(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	tests := []struct {
		name        string
		code        string
		expectValue float64
		expectUnit  string
		expectIsQty bool
	}{
		{"length conversion", "1000 meters -> km", 1.0, "km", true},
		{"mass conversion", "1000 g -> kg", 1.0, "kg", true},
		{"time conversion", "3600 seconds -> hours", 1.0, "h", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ctx.Interpret(tt.code)
			if err != nil {
				t.Fatalf("Interpret(%q) returned error: %v", tt.code, err)
			}
			if res.Value != tt.expectValue {
				t.Errorf("Expected value %f, got %f", tt.expectValue, res.Value)
			}
			if res.Unit != tt.expectUnit {
				t.Errorf("Expected unit '%s', got '%s'", tt.expectUnit, res.Unit)
			}
			if res.IsQuantity != tt.expectIsQty {
				t.Errorf("Expected IsQuantity=%v, got %v", tt.expectIsQty, res.IsQuantity)
			}
		})
	}
}
func TestResult_StructFields(t *testing.T) {
	// Test that Result struct has all expected fields
	res := Result{
		StringOutput: "test output",
		Value:        42.0,
		IsQuantity:   true,
		Unit:         "meters",
	}
	if res.StringOutput != "test output" {
		t.Error("StringOutput field not set correctly")
	}
	if res.Value != 42.0 {
		t.Error("Value field not set correctly")
	}
	if !res.IsQuantity {
		t.Error("IsQuantity field not set correctly")
	}
	if res.Unit != "meters" {
		t.Error("Unit field not set correctly")
	}
}
func TestFree_MultipleCalls(t *testing.T) {
	ctx := NewContext()
	ctx.Free()
	// Should not panic when called multiple times
	ctx.Free()
}
func BenchmarkInterpret_SimpleCalculation(b *testing.B) {
	ctx := NewContext()
	defer ctx.Free()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ctx.Interpret("2 + 2")
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkInterpret_UnitConversion(b *testing.B) {
	ctx := NewContext()
	defer ctx.Free()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ctx.Interpret("120 km/h -> m/s")
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkInterpret_ComplexExpression(b *testing.B) {
	ctx := NewContext()
	defer ctx.Free()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ctx.Interpret("(10 + 5) * 2 - 3 / 1.5")
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkSetVariable(b *testing.B) {
	ctx := NewContext()
	defer ctx.Free()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ctx.SetVariable("bench_var", float64(i), "meters")
		if err != nil {
			b.Fatal(err)
		}
	}
}
