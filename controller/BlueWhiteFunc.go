package controller

import (
	"ayachan/model"
	"math"
)

func calculator(k float64, h float64, t float64, b float64, m float64) (F float64) {
	sum := 0.0
	for i := 0; i <= 120; i++ {
		sum += math.Pow(math.Log(t*t*t*h*k/3-k*t/math.Pi*math.Cos(math.Pi*h)+2*h/math.Pi*math.Sin(math.Pi*k/2)), float64(i)) * b * m
	}
	return sum / 121.0
}

func BlueWhiteFunc(song model.Detail) (diff float32) {
	var C1 float64
	var C2 int
	C1 = math.Log10(calculator(float64(song.TotalNote), float64(song.TotalHitNote), float64(song.TotalTime), float64(song.MainBPM), float64(song.MaxScreenNPS))) / 16.0
	minDiff := 0
	minValue := math.Inf(1)
	if song.TotalTime >= 180 {
		C2 = 1
	} else {
		C2 = 0
	}
	for a := 0; a < 121; a++ {
		C3 := math.Abs(C1 - model.MinusTable[a][C2] - float64(a)) //"-0.1" is correctable.
		if C3 < minValue {
			minDiff = a
			minValue = C3
		}
	}
	diff = float32(C1 - model.MinusTable[minDiff][C2])
	if diff < 0 || diff > 120 {
		diff = 0.0
	}
	return diff
}
