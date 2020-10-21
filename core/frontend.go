// +build js,wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for WebAssembly interface to the irradiance simulator
package main

import (
	"syscall/js"
)

func floats1DToJS(array []float32) interface{} {
	result := make([]interface{}, 0)
	for _, value := range array {
		result = append(result, value)
	}
	return result
}

func floats2DToJS(array [][]float32) interface{} {
	result := make([]interface{}, 0)
	for _, value := range array {
		result = append(result, floats1DToJS(value))
	}
	return result
}

func floats3DToJS(array [][][]float32) interface{} {
	result := make([]interface{}, 0)
	for _, value := range array {
		result = append(result, floats2DToJS(value))
	}
	return result
}

func ints1DToJS(array []int) interface{} {
	result := make([]interface{}, 0)
	for _, value := range array {
		result = append(result, value)
	}
	return result
}

func ints2DToJS(array [][]int) interface{} {
	result := make([]interface{}, 0)
	for _, value := range array {
		result = append(result, ints1DToJS(value))
	}
	return result
}

func irradianceSimulation(this js.Value, args []js.Value) interface{} {
	levels := make([]float32, 0)
	for i := 0; i < 6; i++ {
		levels = append(levels, float32(args[0].Index(i).Float()))
	}
	orientationIndex := rrdncOrientation(args[1].Int())
	elevationIndex := rrdncElevation(args[2].Int())
	countAlongX := uint8(args[3].Int())
	countAlongY := uint8(args[4].Int())
	spacingInches := uint8(args[5].Int())
	simulation := rrdncSimulation(levels, orientationIndex, elevationIndex, countAlongX, countAlongY, spacingInches)
	result := make(map[string]interface{})
	result["irradianceMaps"] = floats3DToJS(simulation.IrradianceMaps)
	result["luminairesMap"] = ints2DToJS(simulation.LuminairesMap)
	result["minima"] = floats1DToJS(simulation.Minima)
	result["maxima"] = floats1DToJS(simulation.Maxima)
	result["means"] = floats1DToJS(simulation.Means)
	return result
}

func irradianceSpectrum(this js.Value, args []js.Value) interface{} {
	irradiances := make([]float32, 0)
	for i := 0; i < 6; i++ {
		irradiances = append(irradiances, float32(args[0].Index(i).Float()))
	}
	spectrum := rrdncSpectrum(irradiances)
	result := make([]interface{}, 0)
	for _, value := range spectrum {
		result = append(result, value)
	}
	return result
}

func main() {
	window := js.Global().Get("window")
	window.Set("irradianceSimulation", js.FuncOf(irradianceSimulation))
	window.Set("irradianceSpectrum", js.FuncOf(irradianceSpectrum))
	select {}
}
