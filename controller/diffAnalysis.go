package controller

import (
	"ayachan/model"
)

func getDiff(detail model.Detail) (diff model.Diffs) {
	var keys = [7]string{"fingerMaxHPS", "totalNPS", "maxSpeed", "flickNoteInterval", "noteFlickInterval", "maxScreenNPS", "totalHPS"}
	var values = [7]float32{
		detail.FingerMaxHPS,
		detail.TotalNPS,
		detail.MaxSpeed,
		detail.FlickNoteInterval,
		detail.NoteFlickInterval,
		detail.MaxScreenNPS,
		detail.TotalHPS,
	}

	diffDistribution, base, ceil := model.QueryDiffDistribution(detail.Diff)
	var diffMap map[string]float32
	diffMap = make(map[string]float32)
	for i := 0; i < len(keys); i++ {
		rank := model.QueryRank(keys[i], values[i], detail.Diff)
		if rank != 0 {
			var j int
			for j = base + 1; j <= ceil && diffDistribution[j] >= rank; j++ {
			}
			if j > ceil {
				diffMap[keys[i]] = float32(ceil) + 0.2 - float32(diffDistribution[ceil]-rank)/float32(diffDistribution[ceil])
			} else {
				diffMap[keys[i]] = float32(j) - float32(diffDistribution[j-1]-rank)/float32(diffDistribution[j-1]-diffDistribution[j])
			}
		} else {
			k, b := model.CalcDiffLiner(keys[i], detail.Diff, diffDistribution[ceil-1], ceil)
			diffMap[keys[i]] = k*values[i] + b
		}
	}
	diff = model.Diffs{
		FingerMaxHPS:      diffMap["fingerMaxHPS"],
		TotalNPS:          diffMap["totalNPS"],
		MaxSpeed:          diffMap["maxSpeed"],
		FlickNoteInterval: diffMap["flickNoteInterval"],
		NoteFlickInterval: diffMap["noteFlickInterval"],
		MaxScreenNPS:      diffMap["maxScreenNPS"],
		TotalHPS:          diffMap["totalHPS"],
		BlueWhiteFunc:     0,
		TotalDiff:         0,
	}
	return diff
}
