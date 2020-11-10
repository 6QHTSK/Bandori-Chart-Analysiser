package controller

import (
	"ayachan/model"
	"fmt"
	"log"
	"math"
	"sort"
)

/*
bestdori 音符类解释
所有音符类共有元素: beat : <此音符的拍数>
	1. BPM 变化音符：必定含有 type : System , cmd: BPM, bpm : <此BPM变化音符变化的BPM>
	2. 一般音符类：均含有 type : Note，lane : <该音符所在轨道号> ,
		(1). 单点类音符：note: Single
			i. 粉键： flick : True
			ii. 蓝键：无额外内容
		(2). 绿条类音符: note: Slide pos: <A,B等用于区分绿条的编号>
			i. 起始键：start: True
			ii. 中间节点： 无额外内容
			iii. 尾节点： end: True
			iv. 尾粉：end: True, flick : True
*/

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
	chart, res.BPMLow, res.BPMHigh, res.MainBPM, res.TotalTime, res.TotalNote = calcTime(chart)
	if res.TotalTime <= 20 {
		res.Error = "20+ Seconds Required"
		return res
	}
	if res.TotalNote <= 20 {
		res.Error = "20+ Notes Required"
		return res
	}
	res.TotalNPS, res.TotalHitNote, res.TotalHPS = CalcChartDetails(chart)
	res.ActiveNPS, res.ActiveHPS, res.ActivePercent, res.MaxScreenNPS = activateDetails(chart)
	finger1, finger2, err := play(chart)
	if err != nil {
		res.Error += err.Error() + "\n"
		return res
	}
	res.LeftPercent, res.FingerMaxHPS, res.FlickNoteInterval, res.NoteFlickInterval, res.MaxSpeed, err = calcDetails(&finger1, &finger2)
	if err != nil {
		res.Error += err.Error() + "\n"
	}
	return res
}

//计算bestdori格式的各个note的时间；
//输入：bestdori格式的谱面；
//返回值：含时间的纯净note列表，最低 bpm， 最高bpm
func calcTime(chart []Note) ([]Note, float32, float32, float32, float32, int) {
	var noteList []Note
	var bpm, offsetTime, offsetBeat, BPMLow, BPMHigh, maxBPMTime, mainBPM float32
	var bpmMap map[float32]float32
	var BPMCounter int
	bpmMap = make(map[float32]float32)
	bpm = 60
	offsetTime = 0.0
	offsetBeat = 0.0
	BPMLow = math.MaxFloat32
	BPMHigh = -1.0
	mainBPM = 60
	maxBPMTime = -1.0
	BPMCounter = 0
	isThereNote := false

	sort.Slice(noteList, func(i, j int) bool {
		if noteList[i].Beat == noteList[j].Beat {
			return noteList[i].Lane < noteList[j].Lane
		}
		return noteList[i].Beat < noteList[j].Beat
	})

	for index, note := range chart {
		if note.Type == "System" {
			lastOffsetTime := offsetTime
			offsetTime = (note.Beat-offsetBeat)*(60.0/bpm) + offsetTime
			bpmMap[bpm] = bpmMap[bpm] + offsetTime - lastOffsetTime
			bpm = abs(note.BPM).(float32)
			BPMCounter++
			if note.BPM < 0 {
				bpm = -bpm
			}
			offsetBeat = note.Beat
			if !isThereNote {
				BPMLow = math.MaxFloat32
				BPMHigh = -1.0
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
			note.index = index - BPMCounter
			noteList = append(noteList, note)
		}
	}
	bpmMap[bpm] = bpmMap[bpm] + noteList[len(noteList)-1].Time - offsetTime
	for bpmItem := range bpmMap {
		if bpmMap[bpmItem] > maxBPMTime {
			maxBPMTime = bpmMap[bpmItem]
			mainBPM = bpmItem
		}

	}
	//返回值：含时间的纯净note列表（以beat为顺序，相同时间以lane为顺序），最低 bpm， 最高bpm
	return noteList, BPMLow, BPMHigh, mainBPM, noteList[len(noteList)-1].Time - noteList[0].Time + 0.1, len(noteList)
}

//计算在某个beat、某个time时，该手指的位置。
//输入：拍数beat,时间time; 返回:浮点数lane
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

//判断击打输入的音符时，手指是否可用
//输入：bestdori音符，输出：布尔值，表示是否可用
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

//将音符插入noteList中，输入要插入note，
//返回 0:插入失败，1 插入一般音符 ， 2 插入结束音符 ， 3 插入开始音符 ， 4 插入节点
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

func (finger *Finger) appendNote(index int, chart []Note) {
	if chart[index].Note.Note == "Slide" && chart[index].Start {
		finger.appendSlide(index, chart)
	} else {
		_, err := finger.append(chart[index])
		if err != nil {
			log.Println(err.Error())
		}
	}
}

//将绿条中所有的元素append到finger里
//输入：起始绿条note，谱面，起始绿条的index，append入的手指
//返回 绿条尾的note
func (finger *Finger) appendSlide(index int, chart []Note) Note {
	for i := index; i < len(chart); i++ {
		status, _ := finger.append(chart[i])
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

//返回time时间前最近的那个note的时间
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

//输入一段绿条中的三个音符,判断其是否需要滑动来击中并标记
//输入：三个slide音符，输出，是否需要滑动来击中
func eraseTick(note1, note2, note3 Note) bool {
	if note1.Lane == note3.Lane && abs(note2.Lane-note1.Lane).(int) < 1 && isSameSlide(note1, note2, note3) {
		note2.ignore = true
		return true
	}
	return false
}

//判断一个note是不是需要击打的note;
//输入：一堆note，输出：这些note是不是都可击打
func isHitNote(args ...Note) bool {
	for _, note := range args {
		if note.Note.Note == "Single" || note.Note.Note == "Slide" && note.Note.Start {
			continue
		} else {
			return false
		}
	}
	return true
}

//返回并标记双压.
//输入值：两个需要判断的note，输出值：是否为双压，且对双压执行了标级
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

//返回并标记双压.
//输入值：两个需要判断的note，输出值：是否为双压，且对双压执行了标级
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

//计算这个绿条要用哪只手接，并标记
//输入一个绿条开始键，无输出
func slidePos(st int, chart *[]Note) (err error) {
	tmp := Finger{defaultLane: 4}
	end := tmp.appendSlide(st, *chart)
	for i := st; i < len(*chart); i++ {
		note := (*chart)[i]
		if note.Note.Beat > end.Note.Beat {
			break
		}
		if !tmp.existNote(note.index) {
			cur := tmp.position(note.Note.Beat, note.Note.Time)
			if float32(note.Note.Lane) < cur {
				(*chart)[st].finger = 3
				break
			} else if float32(note.Note.Lane) > cur {
				(*chart)[st].finger = -3
				break
			} else {
				return fmt.Errorf("Slide and Note in same lane!")
			}
		}
	}
	return nil
}

func play(chart []Note) (lHand, rHand Finger, err error) {
	lHand = Finger{defaultLane: 2}
	rHand = Finger{defaultLane: 6}
	for i := 1; i < len(chart); i++ {
		if chart[i-1].Note.Note == "Slide" && chart[i-1].Note.Start {
			err := slidePos(i-1, &chart)
			if err != nil {
				return Finger{}, Finger{}, err
			}
		}
		_, err := isDouble(&chart[i-1], &chart[i])
		if err != nil {
			log.Println(err.Error())
			return Finger{}, Finger{}, err
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
				return Finger{}, Finger{}, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			lHand.appendNote(i, chart)
		} else if !lAva && rAva {
			if float32(note.Lane) <= lPos {
				return Finger{}, Finger{}, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			rHand.appendNote(i, chart)
		} else if !lAva && !rAva {
			return Finger{}, Finger{}, fmt.Errorf("No Hand available")
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

// 计算一只手/一根手指的note的前后间距
// 假定note的形式是如图所示的：
//		---> time
// note <- back_delta -> current_note <- front_delta -> note
// back_delta: 当前note和前一个note的时间间距
// front_delta： 当前note和后一个note的时间间距
func (finger *Finger) calcDelta() int {
	noteList := &finger.noteList
	sort.Slice(*noteList, func(i, j int) bool {
		return (*noteList)[i].Time < (*noteList)[j].Time
	})
	(*noteList)[0].back = []float32{0, -1}
	(*noteList)[len(*noteList)-1].front = []float32{0, -1}
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

//对note_list进行划分
func SliceNoteList(noteList []Note) (res []*[]Note, sliceTime float32) {
	var subNoteTime float32
	var subNoteList []Note
	var tmp *[]Note
	subNoteTime = 0.0
	sliceTime = 1.0
	tmp = new([]Note)
	for _, note := range noteList {
		if note.Time-subNoteTime > sliceTime {
			for _, note := range subNoteList {
				*tmp = append(*tmp, note)
			}
			res = append(res, tmp)
			subNoteList = []Note{}
			tmp = new([]Note)
			subNoteTime += sliceTime
		}
		subNoteList = append(subNoteList, note)
	}
	return res, sliceTime
}

//计算谱面的需要分离左右手的数据
func calcDetails(lHand, rHand *Finger) (float32, float32, float32, float32, float32, error) {
	var lPercent, fingerMaxHPS, noteFlickInterval, flickNoteInterval, maxSPD float32
	var rankStart, rankEnd int
	var subHPSList, flickInterval1, flickInterval2, SPD []float32
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
	sort.Slice(subHPSList, func(i, j int) bool {
		return subHPSList[i] < subHPSList[j]
	})
	rankStart = len(subHPSList)
	rankEnd = int(math.Floor(float64(len(subHPSList)) * 0.95))
	var sum float32
	sum = 0.0
	for i := rankEnd; i < rankStart; i++ {
		sum += subHPSList[i]
	}
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

//计算chart中达到最大（或标杆较大）nps/hps的70%部分的平均nps/hps
//一般认为，活跃部分谱面要求达到整个谱面的50%
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
			if len(chartData) >= 5 {
				for _, x := range chartData[0:5] {
					sum += x
				}
				maxScreenNPS = sum / 5
			}
		}
	}
	if activePercent[0] > activePercent[1] {
		maxActivePercent = activePercent[0]
	} else {
		maxActivePercent = activePercent[1]
	}
	return activeData[0], activeData[1], maxActivePercent, maxScreenNPS
}

func CalcChartDetails(chart []Note) (float32, int, float32) {
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
	return totalNPS, totalHitNote, totalHPS
}

//计算这些note中某些差值时间的平均值
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

func multipleSpeed(detail *model.Detail, Speed float32) {
	detail.MainBPM = detail.MainBPM * Speed
	detail.FlickNoteInterval = detail.FlickNoteInterval * Speed
	detail.NoteFlickInterval = detail.NoteFlickInterval * Speed
	detail.TotalTime = detail.TotalTime / Speed
	detail.MaxScreenNPS = detail.MaxScreenNPS * Speed
	detail.TotalHPS = detail.TotalHPS * Speed
	detail.TotalNPS = detail.TotalNPS * Speed
	detail.MaxSpeed = detail.MaxSpeed * Speed
	detail.BPMHigh = detail.BPMHigh * Speed
	detail.BPMLow = detail.BPMLow * Speed
	detail.FingerMaxHPS = detail.FingerMaxHPS * Speed
	detail.ActiveHPS = detail.ActiveHPS * Speed
	detail.ActiveNPS = detail.ActiveNPS * Speed
}
