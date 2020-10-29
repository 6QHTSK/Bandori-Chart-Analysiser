package controller

import (
	"ayachan/model"
	"fmt"
	"log"
	"math"
	"sort"
)

var mainBPM float32

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
	//手指类，用于存放用这根手指击打的所有音符全集
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

//计算bestdori格式的各个note的时间；
//输入：bestdori格式的谱面；
//返回值：含时间的纯净note列表，最低 bpm， 最高bpm
func calcTime(chart []Note) ([]Note, []Note, float32, float32, float32) {
	var noteList []Note //纯净的，全部是note的列表，返回值
	var bpmChart []Note //记录bpm的chart
	var bpm, offsetTime, offsetBeat, BPMLow, BPMHigh, maxOffsetTime float32
	bpm = 60                 //谱面的实时BPM
	offsetTime = 0.0         //上一个BPM 变化点的时间
	offsetBeat = 0.0         //上一个BPM 变化点的拍数
	BPMLow = math.MaxFloat32 //最小bpm
	BPMHigh = -1.0           //最大bpm，不存在负bpm
	maxOffsetTime = -1.0
	isThereNote := false //排除前面用来改变offset用的bpm变化点

	for index, note := range chart {
		if note.Type == "System" {
			//此时应为BPM 变化点
			lastOffsetTime := offsetTime
			//计算与上一个BPM变化点的拍数，结合上一个BPM算出此时的时间
			offsetTime = (note.Beat-offsetBeat)*(60.0/bpm) + offsetTime
			bpm = abs(note.BPM).(float32) //更新实时BPM，为防止负BPM搞事情加绝对值
			if note.BPM < 0 {
				bpm = -bpm
			}
			offsetBeat = note.Beat //更新上一个BPM的拍数
			bpmChart = append(bpmChart, note)
			if !isThereNote {
				//检测到是用来改变offset用的BPM变化点
				BPMLow = math.MaxFloat32
				BPMHigh = -1.0
			}
			if offsetTime-lastOffsetTime > maxOffsetTime {
				mainBPM = bpm
				maxOffsetTime = offsetTime - lastOffsetTime
			}
			if bpm < BPMLow {
				//更新bpm最小值
				BPMLow = bpm
			}
			if bpm > BPMHigh {
				//更新bpm最大值
				BPMHigh = bpm
			}
		} else if note.Type == "Note" {
			//此时为正常的音符
			isThereNote = true //排除上一个BPM变化点是用来改变offset用的变化点
			//计算与上一个BPM变化点的拍数，结合上一个BPM算出此时的时间
			note.Time = (note.Beat-offsetBeat)*(60.0/bpm) + offsetTime
			note.index = index
			noteList = append(noteList, note) //加入note列表以便后面使用
		}
	}
	sort.Slice(noteList, func(i, j int) bool {
		if noteList[i].Beat == noteList[j].Beat {
			return noteList[i].Lane > noteList[j].Lane
		}
		return noteList[i].Beat < noteList[j].Beat
	})
	if noteList[len(noteList)-1].Time-offsetTime > maxOffsetTime {
		mainBPM = bpm
	}
	//返回值：含时间的纯净note列表（以beat为顺序，相同时间以lane为顺序），最低 bpm， 最高bpm
	return noteList, bpmChart, BPMLow, BPMHigh, mainBPM
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
				//如果都是绿条的一部分
				return float32(pre.Lane) + float32(cur.Lane-pre.Lane)*(time-pre.Time)/(cur.Time-pre.Time)
				//当前轨道计算公式：起始轨道+ 轨道差值 * （当前时间差值/总时间差值)
			} else if time-pre.Time > 0.5 {
				//离上一个note远
				return float32(finger.defaultLane)
			}
			//离上一个note近
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
			//如果有相同beat的音符，判定此手指不可用
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

func eraseTick(note1, note2, note3 Note) bool {
	//输入一段绿条中的三个音符,判断其是否需要滑动来击中并标记
	//输入：三个slide音符，输出，是否需要滑动来击中
	if note1.Lane == note3.Lane && abs(note2.Lane-note1.Lane).(int) < 1 && isSameSlide(note1, note2, note3) {
		//满足起始、终止轨道相同，且中间键与上述键轨道数差距小于1
		note2.ignore = true //可忽略中间键的存在
		return true
	}
	return false
}

//判断一个note是不是需要击打的note;
//输入：一堆note，输出：这些note是不是都可击打
func isHitNote(args ...Note) bool {
	for _, note := range args {
		if note.Note.Note == "Single" || (note.Note.Note == "Slide" && note.Note.Start) {
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
		//判定为双压
		if note1.Lane < note2.Lane {
			//note1在note2左边
			note1.finger = -2 //note1 必须用左手击打
			note2.finger = 2  //note2 必须用右手击打
		} else if note1.Lane > note2.Lane {
			note1.finger = 2  //note1 必须用右手击打
			note2.finger = -2 //note2 必须用左手击打
		} else {
			//note重叠，抛出错误
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
		//先左后右，先右后左，判定为交互的一部分
		if note2.Lane > note1.Lane && note1.Lane <= 5 && note3.Lane <= 5 && note2.Lane >= 3 {
			//左右左结构
			note1.finger = -1
			note2.finger = 1
			note3.finger = -1
			return true
		}
		if note2.Lane < note1.Lane && note1.Lane >= 3 && note3.Lane >= 3 && note2.Lane <= 5 {
			//右左右结构
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
func slidePos(st int, chart *[]Note) {
	tmp := Finger{defaultLane: 4}      //暂借finger类存储绿条
	end := tmp.appendSlide(st, *chart) //计算出绿条的结尾
	for i := st; i < len(*chart); i++ {
		note := (*chart)[i]
		if note.Note.Beat > end.Note.Beat {
			//已经运行出绿条范围
			break
		}
		if !tmp.existNote(note.index) {
			//发现绿条范围内非该绿条元素
			cur := tmp.position(note.Note.Beat, note.Time)
			if float32(note.Lane) < cur {
				(*chart)[st].finger = 3 //有元素出现在左边，用右手打
				break
			} else if float32(note.Note.Lane) > cur {
				(*chart)[st].finger = -3 //有元素出现在右边，用左手打
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
		//标记双压与计算绿条
		if chart[i-1].Note.Note == "Slide" && chart[i-1].Note.Start {
			slidePos(i-1, &chart)
		}
		_, err := isDouble(&chart[i-1], &chart[i])
		if err != nil {
			log.Println(err.Error())
		}
	}
	//标记交互
	for i := 2; i < len(chart); i++ {
		isTrill(&chart[i-2], &chart[i-1], &chart[i])
	}
	//正式游玩
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
			//左手可用，右手不可用
			if float32(note.Lane) >= rPos {
				return lHand, rHand, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			lHand.appendNote(i, chart)
		} else if !lAva && rAva {
			//右手可用，左手不可用
			if float32(note.Lane) <= lPos {
				//右手需要跨手
				return lHand, rHand, fmt.Errorf("Hand crossing!(Left -> Right)")
			}
			rHand.appendNote(i, chart)
		} else if !lAva && !rAva {
			//双手均不可用，报错
			return lHand, rHand, fmt.Errorf("No Hand available")
		} else {
			//双手均可用
			if abs(note.finger).(int) > 0 {
				//如果之前有双压/绿条/交互标记
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
	noteList := &finger.noteList //首先读取一根手指所要击打的所有note
	//对其进行排序
	sort.Slice(*noteList, func(i, j int) bool {
		return (*noteList)[i].Time < (*noteList)[j].Time
	})
	(*noteList)[0].back = []float32{0, -1} //第一个note是没有前面note的，所以置-1
	(*noteList)[len(*noteList)-1].back = []float32{0, -1} //最后一个note是没有后面note的，所以置-1
	for i := 0; i < len(*noteList)-1; i++ {
		//对每一个note进行计算
		speed := abs(float32((*noteList)[i+1].Lane) - float32((*noteList)[i].Lane)/(*noteList)[i+1].Time).(float32)
		deltaTime := (*noteList)[i+1].Time - (*noteList)[i].Time
		if (*noteList)[i].Note.Note == "Slide" && (*noteList)[i].End {
			deltaTime *= 1.5 //对于绿条尾，因为不涉及按下再松手，惩罚性加时（这里是避免a2z note 接粉难度过高的问题）
		}
		(*noteList)[i+1].back = []float32{speed, deltaTime}
		(*noteList)[i].front = []float32{speed, deltaTime}
		//这个note的front_delta值按定义等于下一个note的back_delta值。注意到返回值有两个部分
		//前半部分是速度。即这个note和上个note之间手指在X轴上的移动速度
		//后半部分是间隔时间，即这个note和上个note之间的时间差值（Y轴）
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

//计算谱面的需要分离左右手的数据
func calcDetails(lHand, rHand Finger) (float32, float32, float32, float32, float32, error) {
	var lPercent, fingerMaxHPS, noteFlickInterval, flickNoteInterval, maxSPD float32
	var rankStart, rankEnd int
	var subHPSList, flickInterval1, flickInterval2, SPD, tmp []float32
	var hitNoteList, flickNoteList []Note

	l := lHand.calcDelta() //计算左手击打音符的数量和标记差值时间
	r := rHand.calcDelta() //计算右手击打音符的数量和标记差值时间
	lPercent = float32(l) / float32(l+r) //计算左右手分配压力
	ptrList := []*[]Note{&lHand.noteList, &rHand.noteList} //打包note_list便于循环
	for _, handList := range ptrList {
		//先进行爆发单手检测
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
		//再进行爆发粉键检测，包括前向检测和后向检测
		sort.Slice(*handList, func(i, j int) bool {
			return (*handList)[i].front[1] < (*handList)[j].front[1]
		})
		flickNoteList = []Note{} //为单个粉键的时候，可以计算该粉键和前面那个note的间隔
		for _, note := range *handList {
			if note.Flick && note.Note.Note != "Slide" {
				flickNoteList = append(flickNoteList, note)
			}
		}
		if len(flickNoteList) >= 10 {
			//排除掉最前面的可能是手分配错误造成的较小note间隔
			rankStart = int(math.Ceil(float64(len(flickNoteList)) * 0.9))
			rankEnd = rankStart - 3 //计算3个note的间隔
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
			//排除掉最前面的可能是手分配错误造成的较小note间隔
			rankStart = int(math.Ceil(float64(len(flickNoteList)) * 0.95))
			rankEnd = rankStart - 3 //计算6个note的间1隔
			flickInterval2 = append(flickInterval2, 1/average(flickNoteList[rankEnd:rankStart], 1, 0))
		} else {
			flickInterval2 = append(flickInterval2, 0)
		}
		//再进行位移速度的检测
		sort.Slice(*handList, func(i, j int) bool {
			return (*handList)[i].front[0] < (*handList)[j].front[0]
		})
		rankStart = len(*handList)
		rankEnd = int(math.Floor(float64(len(*handList)) * 0.96)) //取前 4% 的音符进行最小位移计算
		SPD = append(SPD, average((*handList)[rankEnd:rankStart], 0, 0))
	}
	//返回这些检测值的最大值
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
			return chartData[i] < chartData[j]
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
