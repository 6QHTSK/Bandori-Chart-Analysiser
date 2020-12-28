package controller

import (
	"ayachan/model"
	"math"
	"sort"
)

func getDiff(detail model.Detail) (diff model.Diffs) {
	blueWhite := BlueWhiteFunc(detail)
	if blueWhite < 0 || blueWhite >= 60 {
		blueWhite = 0
	}
	diff = model.Diffs{
		FingerMaxHPS:      CalcItemDiff("fingerMaxHPS", detail.FingerMaxHPS, detail.Diff),
		TotalNPS:          CalcItemDiff("totalNPS", detail.TotalNPS, detail.Diff),
		FlickNoteInterval: CalcItemDiff("flickNoteInterval", detail.FlickNoteInterval, detail.Diff),
		NoteFlickInterval: CalcItemDiff("noteFlickInterval", detail.NoteFlickInterval, detail.Diff),
		MaxScreenNPS:      CalcItemDiff("maxScreenNPS", detail.MaxScreenNPS, detail.Diff),
		TotalHPS:          CalcItemDiff("totalHPS", detail.TotalHPS, detail.Diff),
		MaxSpeed:          CalcItemDiff("maxSpeed", detail.TotalHPS, detail.Diff),
		BlueWhiteFunc:     blueWhite,
	}
	return diff
}

func CalcItemDiff(key string, value float32, diff int) (itemDiff float32) {
	if value == 0 {
		return 0
	}
	diffDistribution, base, ceil := model.QueryDiffDistribution(diff)
	rank := model.QueryRank(key, value, diff)
	if rank != 0 {
		var j int
		for j = ceil; j >= base && diffDistribution[j] <= rank; j-- {
			rank -= diffDistribution[j]
		}
		if j == ceil {
			itemDiff = float32(ceil) + 0.2 - float32(rank-1)/float32(diffDistribution[ceil])
		} else if j >= base {
			itemDiff = float32(j) + float32(diffDistribution[j]-rank)/float32(diffDistribution[j])
		} else {
			itemDiff = float32(base)
		}
		if key == "flickNoteInterval" && rank != 1 {
			itemDiff *= 0.97
		}
	} else {
		if diff <= 2 {
			return CalcItemDiff(key, value, diff+1)
		}
		k, b := model.CalcDiffLiner(key, diff, diffDistribution[ceil-1], ceil)
		itemDiff = k*value + b
	}
	return itemDiff
}

func BlueWhiteFunc(detail model.Detail) (diff float32) {
	if detail.Diff >= 3 {
		type pair struct {
			Level int
			Value float64
		}
		xs := []pair{{20, 20.94}, {21, 21.92}, {22, 22.90}, {23, 23.88},
			{24, 24.86}, {25, 24.97}, {26, 25.95}, {27, 26.94},
			{28, 27.80}, {29, 28.78}, {30, 29.76}, {31, 30.75},
			{32, 31.73}, {33, 32.71}, {34, 33.69}, {35, 34.67},
			{36, 35.65}, {37, 36.63}, {38, 37.61}, {39, 38.60},
			{40, 39.58}, {41, 40.56}, {42, 41.54}, {43, 42.52},
			{44, 43.50}, {45, 44.48}, {46, 45.46}, {47, 46.44},
			{48, 47.42}, {49, 48.41}, {50, 49.39}, {51, 50.37},
			{52, 51.35}, {53, 52.33}, {54, 53.31}, {55, 54.29},
			{56, 55.27}, {57, 56.25}, {58, 57.23}, {59, 58.21},
			{60, 59.20}}
		var rtr []pair
		for _, item := range xs {
			rtr = append(rtr, pair{item.Level, item.Value + 0.0625*math.Log10(float64(float32(detail.TotalNote)*detail.MainBPM*detail.TotalTime)/11002803938.0*(math.Pow(float64(detail.TotalNote), 3.0)/90000.0+math.Pow(float64(detail.TotalNote), 2)/40000.0+float64(detail.TotalNote)/10000.0))})
		}
		sort.Slice(rtr, func(i, j int) bool {
			return math.Abs(float64(rtr[i].Level)-rtr[i].Value) < math.Abs(float64(rtr[j].Level)-rtr[j].Value)
		})
		return float32(rtr[0].Value)
	} else {
		x := []float64{3.13188, 4.58318, 5.27976}
		return float32(x[detail.Diff] * math.Log10(math.Pow(float64(detail.TotalNote), 2.0)*float64(detail.MainBPM)/(2545.37*math.Sqrt(2.0)*math.Pi)))
	}
}
