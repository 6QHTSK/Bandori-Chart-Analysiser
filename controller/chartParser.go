package controller

import (
	"ayachan/model"
	"fmt"
	"log"
	"math"
	"sort"
)

var mainBPM float32

//var BPMLow, BPMHigh float32

type Note struct {
	model.Note
	index  int
	finger int
	ignore bool
	back   []float32
	front  []float32
}
type Finger struct {
	defaultLane int
	noteList    []Note
	hold        bool
	pos         string
}

func getChartDetail(chartID int, diff int, chart []Note) (res model.Detail) {
	res.ID = chartID
	res.Diff = diff
	chart, _, res.BPMLow, res.BPMHigh, res.MainBPM = calcTime(chart)
	res.TotalTime, res.TotalNote, res.TotalNPS, res.TotalHitNote, res.TotalHPS = CalcChartDetails(chart)
	if res.TotalTime <= 20 {
		res.Error = "20+ Seconds Required"
		return res
	}
	if res.TotalNote <= 20 {
		res.Error = "20+ Notes Required"
		return res
	}
	res.ActiveNPS, res.ActiveHPS, res.ActivePercent, res.MaxScreenNPS = activateDetails(chart)
	finger1, finger2, err := play(chart)
	if err != nil {
		res.Error += err.Error() + "\n"
	}
	res.LeftPercent, res.FingerMaxHPS, res.FlickNoteInterval, res.NoteFlickInterval, res.MaxSpeed, err = calcDetails(finger1, finger2)
	if err != nil {
		res.Error += err.Error() + "\n"
	}
	return res
}

func calcTime(chart []Note) ([]Note, []Note, float32, float32, float32) {
	var noteList, bpmChart []Note
	var bpm, offsetTime, offsetBeat, BPMLow, BPMHigh, maxOffsetTime float32
	bpm = 60
	offsetTime = 0.0
	offsetBeat = 0.0
	BPMLow = math.MaxFloat32
	BPMHigh = -1.0
	maxOffsetTime = -1.0
	isThereNote := false

	for index, note := range chart {
		if note.Type == "System" {
			lastOffsetTime := offsetTime
			offsetTime = (note.Beat-offsetBeat)*(60.0/bpm) + offsetTime
			bpm = abs(note.BPM).(float32)
			if note.BPM < 0 {
				bpm = -bpm
			}
			offsetBeat = note.Beat
			bpmChart = append(bpmChart, note)
			if !isThereNote {
				BPMLow = math.MaxFloat32
				BPMHigh = -1.0
			}
			if offsetTime-lastOffsetTime > maxOffsetTime {
				mainBPM = bpm
				maxOffsetTime = offsetTime - lastOffsetTime
			}
			if bpm < BPMLow {
				BPMLow = bpm
			}
			if bpm > BPMHigh {
				BPMHigh = bpm
			}
		} else if note.Type == "Note" {
			isThereNote = true
			note.Time = (note.Beat-offsetBeat)*(60.0/bpm) + offsetTime
			note.index = index
			noteList = append(noteList, note)
		}
	}
	sort.Slice(noteList, func(i, j int) bool {
		if noteList[i].Beat == noteList[j].Beat {
			return noteList[i].Lane > noteList[j].Lane
		}
		return noteList[i].Beat > noteList[j].Beat
	})
	if noteList[len(noteList)-1].Time-offsetTime > maxOffsetTime {
		mainBPM = bpm
	}
	return noteList, bpmChart, BPMLow, BPMHigh, mainBPM
}

func (finger *Finger) position(beat float32, time float32) float32 {
	if len(finger.noteList) == 0 {
		return float32(finger.defaultLane)
	}
	for i := 0; i < len(finger.noteList); i++ {
		cur := &finger.noteList[i]
		if cur.Beat == beat {
			return float32(cur.Lane)
		}
		if cur.Beat > beat {
			if i == 0 {
				return float32(finger.defaultLane)
			}
			pre := &finger.noteList[i-1]
			if cur.Note.Note == "Slide" && pre.Note.Note == "Slide" {
				return float32(pre.Lane) + float32(cur.Lane-pre.Lane)*(time-pre.Time)/(cur.Time-pre.Time)
			} else if time-pre.Time > 0.5 {
				return float32(finger.defaultLane)
			}
			return float32(pre.Lane)
		}
	}
	return float32(finger.defaultLane)
}

func (finger *Finger) available(note Note) bool {
	hold := false
	pos := ""
	for _, n := range finger.noteList {
		if n.Beat == note.Beat {
			return false
		}
		if n.Beat > note.Beat {
			break
		}
		if n.Note.Note == "Slide" && n.Start {
			hold = true
			pos = n.Pos
		} else if n.Note.Note == "Slide" && n.End {
			hold = false
			pos = ""
		}
	}
	if hold {
		return pos == note.Pos
	}
	return true
}

func (finger *Finger) append(note Note) (status int, err error) {
	if finger.available(note) {
		finger.noteList = append(finger.noteList, note)
		if note.Note.Note == "Slide" {
			if note.End {
				if !finger.hold {
					return 0, fmt.Errorf("Unexpected End Note!")
				}
				finger.hold = false
				finger.pos = ""
				return 2, nil
			} else if note.Start {
				if finger.hold {
					return 0, fmt.Errorf("Unexpected Start Note!")
				}
				finger.hold = true
				finger.pos = note.Pos
				return 3, nil
			} else {
				if !finger.hold {
					return 4, fmt.Errorf("Unexpected Tick Note!")
				}
			}
		}
		return 1, nil
	}
	return 0, fmt.Errorf("unknown error")
}

func (finger Finger) appendNote(index int, chart []Note) {
	if chart[index].Note.Note == "Slide" && chart[index].Start {
		finger.appendSlide(index, chart)
	} else {
		_, err := finger.append(chart[index])
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (finger Finger) appendSlide(index int, chart []Note) Note {
	for i := index; i < len(chart); i++ {
		status, err := finger.append(chart[i])
		if err != nil {
			log.Println(err.Error())
		}
		if status == 2 {
			return chart[i]
		}
	}
	return Note{}
}

func (finger *Finger) simplified() {
	for i := 2; i < len(finger.noteList); i++ {
		eraseTick(finger.noteList[i-2], finger.noteList[i-1], finger.noteList[i])
	}
}

func (finger Finger) lastNoteTime(time float32) float32 {
	if len(finger.noteList) != 0 {
		for i := len(finger.noteList) - 1; i >= 0; i-- {
			if finger.noteList[i].Time < time {
				return finger.noteList[i].Time
			}
		}
		return 0.0
	}
	return 0.0
}

func (finger Finger) existNote(index int) bool {
	for i := 0; i < len(finger.noteList); i++ {
		if finger.noteList[i].index == index {
			return true
		}
	}
	return false
}

func isSameSlide(args ...Note) bool {
	startFlag := false
	endFlag := false
	for _, note := range args {
		if note.Note.Note != "Slide" {
			return false
		}
		if note.Start {
			if startFlag {
				return false
			}
		} else if note.End {
			endFlag = true
		} else {
			startFlag = true
			if endFlag {
				return false
			}
		}
	}
	return true
}

func eraseTick(note1, note2, note3 Note) bool {
	if note1.Lane == note3.Lane && abs(note2.Lane-note1.Lane).(int) < 1 && isSameSlide(note1, note2, note3) {
		note2.ignore = true
		return true
	}
	return false
}

func isHitNote(args ...Note) bool {
	for _, note := range args {
		if (note.Note.Note == "Single" || note.Note.Note == "Slide") && note.Start {
			continue
		} else {
			return false
		}
	}
	return true
}

func isDouble(note1, note2 *Note) (bool, error) {
	if note1.Beat == note2.Beat && isHitNote(*note1, *note2) {
		if abs(note1.finger) == 2 || abs(note2.finger) == 2 {
			return false, fmt.Errorf("Triple or more notes!")
		}
		if note1.Lane < note2.Lane {
			note1.finger = -2
			note2.finger = 2
		} else if note1.Lane > note2.Lane {
			note1.finger = 2
			note2.finger = -2
		} else {
			return false, fmt.Errorf("Notes in the same lane!")
		}
		return true, nil
	}
	return false, nil
}

func isTrill(note1, note2, note3 *Note) bool {
	var minTime float32
	minTime = 0.15
	if (note1.Lane-note2.Lane)*(note2.Lane-note3.Lane) < 0 && isHitNote(*note1, *note2, *note3) && note1.Time != note2.Time && note2.Time != note3.Time && abs(note1.Time-note2.Time).(float32) < minTime && abs(note2.Time-note3.Time).(float32) < minTime && abs(abs(note1.Beat-note2.Beat).(float32)-abs(note2.Beat-note3.Beat).(float32)).(float32) < 0.01 && note1.finger == 0 && note2.finger == 0 && note3.finger == 0 {
		if note2.Lane > note1.Lane && note1.Lane <= 5 && note3.Lane <= 5 && note2.Lane >= 3 {
			note1.finger = -1
			note2.finger = 1
			note3.finger = -1
			return true
		}
		if note2.Lane < note1.Lane && note1.Lane >= 3 && note3.Lane >= 3 && note2.Lane <= 5 {
			note1.finger = 1
			note2.finger = -1
			note3.finger = 1
			return true
		}
	}
	return false
}

func slidePos(st int, chart *[]Note) {
	tmp := Finger{defaultLane: 4}
	end := tmp.appendSlide(st, *chart)
	for i := st; i < len(*chart); i++ {
		note := (*chart)[i]
		if note.Beat > end.Beat {
			break
		}
		if !tmp.existNote(note.index) {
			cur := tmp.position(note.Beat, note.Time)
			if float32(note.Lane) < cur {
				(*chart)[st].finger = 3
				break
			} else if float32(note.Lane) > cur {
				(*chart)[st].finger = -3
				break
			} else {
				log.Println("Slide and Note in same lane!")
			}
		}
	}
}

func play(chart []Note) (lHand, rHand Finger, err error) {
	lHand = Finger{defaultLane: 2}
	rHand = Finger{defaultLane: 6}
	for i := 1; i < len(chart); i++ {
		if chart[i-1].Note.Note == "Slide" && chart[i-1].Start {
			slidePos(i-1, &chart)
		}
		_, err := isDouble(&chart[i-1], &chart[i])
		if err != nil {
			log.Println(err.Error())
		}
	}
	for i := 2; i < len(chart); i++ {
		isTrill(&chart[i-2], &chart[i-1], &chart[i])
	}
	for i := 0; i < len(chart); i++ {
		note := chart[i]
		lAva := lHand.available(note)
		rAva := rHand.available(note)
		lPos := lHand.position(note.Beat, note.Time)
		rPos := rHand.position(note.Beat, note.Time)
		if lHand.existNote(i) || rHand.existNote(i) {
			continue
		}
		if lAva && !rAva {
			if float32(note.Lane) >= rPos {
				return lHand, rHand, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			lHand.appendNote(i, chart)
		} else if !lAva && rAva {
			if float32(note.Lane) <= lPos {
				return lHand, rHand, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			rHand.appendNote(i, chart)
		} else if !lAva && !rAva {
			return lHand, rHand, fmt.Errorf("No Hand available")
		} else {
			if abs(note.finger).(int) > 0 {
				if note.finger < 0 {
					lHand.appendNote(i, chart)
				}
				if note.finger > 0 {
					rHand.appendNote(i, chart)
				}
			} else if abs(float32(note.Lane)-lPos).(float32) < abs(float32(note.Lane)-rPos).(float32) || abs(float32(note.Lane)-lPos) == abs(float32(note.Lane)-rPos) && (note.Lane < 4 || (note.Lane == 4 && lHand.lastNoteTime(note.Time) < rHand.lastNoteTime(note.Time) && note.Time-lHand.lastNoteTime(note.Time) < 0.15)) {
				lHand.appendNote(i, chart)
			} else {
				rHand.appendNote(i, chart)
			}
		}
	}
	return lHand, rHand, nil
}

func (finger *Finger) calcDelta() int {
	noteList := &finger.noteList
	sort.Slice(*noteList, func(i, j int) bool {
		return (*noteList)[i].Time < (*noteList)[j].Time
	})
	(*noteList)[0].back = []float32{0, -1}
	(*noteList)[len(*noteList)-1].back = []float32{0, -1}
	for i := 0; i < len(*noteList)-1; i++ {
		speed := abs(float32((*noteList)[i+1].Lane) - float32((*noteList)[i].Lane)/(*noteList)[i+1].Time).(float32)
		deltaTime := (*noteList)[i+1].Time - (*noteList)[i].Time
		if (*noteList)[i].Note.Note == "Slide" && (*noteList)[i].End {
			deltaTime *= 1.5
		}
		(*noteList)[i+1].back = []float32{speed, deltaTime}
		(*noteList)[i].front = []float32{speed, deltaTime}
	}
	return len(*noteList)
}

func SliceNoteList(noteList []Note) (res []*[]Note, sliceTime float32) {
	var subNoteTime float32
	var subNoteList []Note
	var tmp *[]Note
	subNoteTime = 0.0
	sliceTime = 1.0
	tmp = new([]Note)
	for _, note := range noteList {
		if note.Time-subNoteTime > sliceTime {
			copy(*tmp, subNoteList)
			res = append(res, tmp)
			subNoteList = []Note{}
			tmp = new([]Note)
			subNoteTime += sliceTime
		}
		subNoteList = append(subNoteList, note)
	}
	return res, sliceTime
}

func calcDetails(lHand, rHand Finger) (float32, float32, float32, float32, float32, error) {
	var lPercent, fingerMaxHPS, noteFlickInterval, flickNoteInterval, maxSPD float32
	var rankStart, rankEnd int
	var subHPSList, flickInterval1, flickInterval2, SPD, tmp []float32
	var hitNoteList, flickNoteList []Note

	l := lHand.calcDelta()
	r := rHand.calcDelta()
	lPercent = float32(l) / float32(l+r)
	ptrList := []*[]Note{&lHand.noteList, &rHand.noteList}
	for _, handList := range ptrList {
		hitNoteList = []Note{}
		for _, note := range *handList {
			if isHitNote(note) {
				hitNoteList = append(hitNoteList, note)
			}
		}
		slicedList, subTime := SliceNoteList(hitNoteList)
		subHPSList = []float32{}
		for _, tmpList := range slicedList {
			subHPSList = append(subHPSList, float32(len(*tmpList))/subTime)
		}
		sort.Slice(*handList, func(i, j int) bool {
			return (*handList)[i].front[1] > (*handList)[j].front[1]
		})
		flickNoteList = []Note{}
		for _, note := range *handList {
			if note.Flick && note.Note.Note != "Slide" {
				flickNoteList = append(flickNoteList, note)
			}
		}
		if len(flickNoteList) >= 10 {
			rankStart = int(math.Ceil(float64(len(flickNoteList)) * 0.9))
			rankEnd = rankStart - 3
			flickInterval1 = append(flickInterval1, 1/average(flickNoteList[rankEnd:rankStart], 1, 1))
		} else {
			flickInterval1 = append(flickInterval1, 0)
		}
		flickNoteList = []Note{}
		for _, note := range *handList {
			if note.Flick {
				flickNoteList = append(flickNoteList, note)
			}
		}
		if len(flickNoteList) >= 10 {
			rankStart = int(math.Ceil(float64(len(flickNoteList)) * 0.95))
			rankEnd = rankStart - 3
			flickInterval2 = append(flickInterval2, 1/average(flickNoteList[rankEnd:rankStart], 1, 0))
		} else {
			flickInterval2 = append(flickInterval2, 0)
		}
		sort.Slice(*handList, func(i, j int) bool {
			return (*handList)[i].front[0] < (*handList)[j].front[0]
		})
		rankStart = len(*handList)
		rankEnd = int(math.Floor(float64(len(*handList)) * 0.96))
		SPD = append(SPD, average((*handList)[rankEnd:rankStart], 0, 0))
	}
	copy(tmp, subHPSList)
	sort.Slice(tmp, func(i, j int) bool {
		return subHPSList[i] < subHPSList[j]
	})
	rankStart = len(tmp)
	rankEnd = int(math.Floor(float64(len(tmp)) * 0.95))
	var sum float32
	sum = 0.0
	if rankStart >= len(tmp) || rankEnd <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("list out of range")
	}
	for i := rankEnd; i < rankStart; i++ {
		sum += tmp[i]
	}
	// !!!!!!!!!!!!wei!!!!!!!!!!!!!!!
	fingerMaxHPS = sum / float32(rankStart-rankEnd)
	noteFlickInterval = 0.0
	flickNoteInterval = 0.0
	maxSPD = 0.0
	// get max value of flickInterval1
	for _, x := range flickInterval1 {
		if x > noteFlickInterval {
			noteFlickInterval = x
		}
	}
	for _, x := range flickInterval2 {
		if x > flickNoteInterval {
			flickNoteInterval = x
		}
	}
	for _, x := range SPD {
		if x > maxSPD {
			maxSPD = x
		}
	}
	return lPercent, fingerMaxHPS, noteFlickInterval, flickNoteInterval, maxSPD, nil
}

func activateDetails(chart []Note) (float32, float32, float32, float32) {
	var sliceTime, maxScreenNPS, maxActivePercent float32
	var activeData, activePercent []float32
	var chartData, activeChartData []float32
	var slicedChart []*[]Note

	activeData = []float32{0.0, 0.0}
	activePercent = []float32{0.0, 0.0}
	for i := 0; i < 2; i++ {
		if i == 0 {
			slicedChart, sliceTime = SliceNoteList(chart)
		} else {
			// filter: isHitNote
			var hitNoteChart []Note
			for _, note := range chart {
				if isHitNote(note) {
					hitNoteChart = append(hitNoteChart, note)
				}
			}
			slicedChart, sliceTime = SliceNoteList(hitNoteChart)
		}
		chartData = []float32{}
		for _, subChart := range slicedChart {
			chartData = append(chartData, float32(len(*subChart))/sliceTime)
		}
		sort.Slice(chartData, func(i, j int) bool {
			return chartData[i] > chartData[j]
		})
		for _, data := range chartData {
			activeChartData = []float32{}
			for _, x := range chartData {
				if x > data*0.7 {
					activeChartData = append(activeChartData, x)
				}
			}
			if float32(len(activeChartData))/float32(len(chartData)) >= 0.60 {
				var sum float32
				sum = 0.0
				for _, x := range activeChartData {
					sum += x
				}
				activeData[i] = sum / float32(len(activeChartData))
				activePercent[i] = float32(len(activeChartData)) / float32(len(chartData))
				break
			}
		}
		if i == 1 {
			var sum float32
			sum = 0.0
			for _, x := range chartData[0:5] {
				sum += x
			}
			maxScreenNPS = sum / 5
		}
	}
	if activePercent[0] > activePercent[1] {
		maxActivePercent = activePercent[0]
	} else {
		maxActivePercent = activePercent[1]
	}
	return activeData[0], activeData[1], maxActivePercent, maxScreenNPS
}

func CalcChartDetails(chart []Note) (float32, int, float32, int, float32) {
	totalTime := chart[len(chart)-1].Time - chart[0].Time + 0.1
	totalNote := len(chart)
	totalNPS := float32(totalNote) / totalTime
	totalHitNote := 0
	for _, note := range chart {
		if isHitNote(note) {
			totalHitNote++
		}
	}
	totalHPS := float32(totalHitNote) / totalTime
	return totalTime, totalNote, totalNPS, totalHitNote, totalHPS
}

func average(notes []Note, key int, position int) float32 {
	var total float32
	if (key != 0 && key != 1) || len(notes) == 0 {
		return 0
	}
	// position: 0-front 1-back
	for _, note := range notes {
		if position == 1 {
			total += note.front[key]
		} else {
			total += note.back[key]
		}
	}
	return total / float32(len(notes))
}

func abs(x interface{}) interface{} {
	switch x.(type) {
	case int:
		{
			if x.(int) < 0 {
				return -x.(int)
			}
			return x.(int)
		}
	case float32:
		{
			if x.(float32) < 0 {
				return -x.(float32)
			}
			return x.(float32)
		}
	}
	return 0
}
