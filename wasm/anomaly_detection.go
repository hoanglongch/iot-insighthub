// +build wasm

package main

import (
	"syscall/js"
	"time"
)

// DetectAnomaly processes a sensor reading and returns true if it exceeds the threshold.
func DetectAnomaly(this js.Value, args []js.Value) interface{} {
	// Retrieve the sensor value from the first argument.
	value := args[0].Float()
	threshold := 75.0 // Example threshold value.
	isAnomaly := value > threshold
	return js.ValueOf(isAnomaly)
}

// BenchmarkAnomaly runs the anomaly detection repeatedly and returns the average time per call (in ms).
// Usage from JS: BenchmarkAnomaly(iterations, value)
func BenchmarkAnomaly(this js.Value, args []js.Value) interface{} {
	iterations := int(args[0].Int())
	if iterations <= 0 {
		iterations = 1000
	}
	value := args[1].Float()
	threshold := 75.0
	var total time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_ = value > threshold // The detection logic.
		total += time.Since(start)
	}
	avg := total.Milliseconds() / int64(iterations)
	return js.ValueOf(avg)
}

func main() {
	c := make(chan struct{}, 0)
	// Expose the anomaly detection functions to JavaScript.
	js.Global().Set("DetectAnomaly", js.FuncOf(DetectAnomaly))
	js.Global().Set("BenchmarkAnomaly", js.FuncOf(BenchmarkAnomaly))
	<-c
}
