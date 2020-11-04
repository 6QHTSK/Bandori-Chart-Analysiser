package model

import (
	"fmt"
)

type Chart struct {
	ID       int    `bson:"id"`
	Diff     int    `bson:"diff"`
	Level    int    `bson:"level"`
	AuthorID int    `bson:"authorID"`
	Artist   string `bson:"artist"`
	Title    string `bson:"title"`
	Notes    []Note `bson:"chart"`
}

type Detail struct {
	ActiveHPS         float32 `bson:"activeHPS"`
	TotalTime         float32 `bson:"totalTime"`
	FingerMaxHPS      float32 `bson:"fingerMaxHPS"`
	TotalNPS          float32 `bson:"totalNPS"`
	LeftPercent       float32 `bson:"leftPercent"`
	MaxSpeed          float32 `bson:"maxSpeed"`
	ID                int     `bson:"id"`
	Error             string  `bson:"error"`
	Diff              int     `bson:"diff"`
	ActiveNPS         float32 `bson:"activeNPS"`
	TotalHitNote      int     `bson:"totalHitNote"`
	ActivePercent     float32 `bson:"activePercent"`
	FlickNoteInterval float32 `bson:"flickNoteInterval"`
	MainBPM           float32 `bson:"mainBPM"`
	TotalNote         int     `bson:"totalNote"`
	BPMLow            float32 `bson:"BPMLow"`
	NoteFlickInterval float32 `bson:"noteFlickInterval"`
	BPMHigh           float32 `bson:"BPMHigh"`
	MaxScreenNPS      float32 `bson:"maxScreenNPS"`
	TotalHPS          float32 `bson:"totalHPS"`
}

type OfficialBasic struct {
	Result bool `json:"result"`
	Data   struct {
		Name       string `json:"name"`
		Band       string `json:"band"`
		Difficulty map[string]struct {
			Level int `json:"level"`
			Notes int `json:"notes"`
		} `json:"difficulty"`
	} `json:"data"`
}

type FanBasic struct {
	Result bool `json:"result"`
	Post   struct {
		Title  string `json:"title"`
		Artist string `json:"artists"`
		Diff   int    `json:"diff"`
		Level  int    `json:"level"`
		Chart  []Note `json:"notes"`
		Author struct {
			Username string `json:"username"`
			Nickname string `json:"nickname"`
		} `json:"author"`
	} `json:"post"`
}

type Author struct {
	AuthorID int    `bson:"authorID"`
	UserName string `bson:"username"`
	NickName string `bson:"nickname"`
}

type Note struct {
	Type   string  `json:"type" bson:"type"`
	CMD    string  `json:"cmd" bson:"cmd"`
	BPM    float32 `json:"bpm" bson:"bpm"`
	Beat   float32 `json:"beat" bson:"beat"`
	Effect string  `json:"effect" bson:"effect"`
	Time   float32 `json:"time" bson:"time"`
	Note   string  `json:"note" bson:"note"`
	Lane   int     `json:"lane" bson:"lane"`
	Skill  bool    `json:"skill" bson:"skill"`
	Start  bool    `json:"start" bson:"start"`
	End    bool    `json:"end" bson:"end"`
	Flick  bool    `json:"flick" bson:"flick"`
	Charge bool    `json:"charge" bson:"charge"`
	Pos    string  `json:"pos" bson:"pos"`
}

func BD2BDFan(officialChart []Note) (fanChart []Note, err error) {
	var leftOccupied, rightOccupied int
	var endBeatLeft, endBeatRight float32
	leftOccupied = 0
	rightOccupied = 0
	endBeatLeft = -1.0
	endBeatRight = -1.0
	for i := 0; i < len(officialChart); i++ {
		note := &officialChart[i]
		if note.Type == "System" && note.CMD == "BPM" {
			note.Time = 0.0
			fanChart = append(fanChart, *note)
		} else if note.Type == "Note" {
			note.Time = 0.0
			if note.Note == "Long" {
				note.Note = "Slide"
				if note.Start {
					if leftOccupied == 0 && note.Beat > endBeatLeft {
						leftOccupied = note.Lane
						note.Pos = "A"
					} else if rightOccupied == 0 && note.Beat > endBeatRight {
						rightOccupied = note.Lane
						note.Pos = "B"
					} else {
						return fanChart, fmt.Errorf("Too many Slides")
					}
				} else if note.End {
					if leftOccupied == note.Lane {
						leftOccupied = 0
						endBeatLeft = note.Beat
						note.Pos = "A"
					} else if rightOccupied == note.Lane {
						rightOccupied = 0
						endBeatRight = note.Beat
						note.Pos = "B"
					} else {
						return fanChart, fmt.Errorf("Unexpected End Note")
					}
				} else {
					return fanChart, fmt.Errorf("Not Such Tyoe Note")
				}
				fanChart = append(fanChart, *note)
			} else if note.Note == "Slide" {
				var notePos int
				if note.Pos == "A" {
					notePos = 10
				} else {
					notePos = 11
				}
				if note.Start {
					if leftOccupied == 0 && note.Beat > endBeatLeft {
						leftOccupied = notePos
						note.Pos = "A"
					} else if rightOccupied == 0 && note.Beat > endBeatRight {
						rightOccupied = notePos
						note.Pos = "B"
					} else {
						return fanChart, fmt.Errorf("Too many Slides")
					}
				} else if note.End {
					if leftOccupied == notePos {
						leftOccupied = 0
						endBeatLeft = note.Beat
						note.Pos = "A"
					} else if rightOccupied == notePos {
						rightOccupied = 0
						endBeatRight = note.Beat
						note.Pos = "B"
					} else {
						return fanChart, fmt.Errorf("Unexpect End Note")
					}
				} else {
					if leftOccupied == notePos {
						note.Pos = "A"
					} else if rightOccupied == notePos {
						note.Pos = "B"
					} else {
						return fanChart, fmt.Errorf("Unexpected Tick Note")
					}
				}
				fanChart = append(fanChart, *note)
			} else {
				fanChart = append(fanChart, *note)
			}
		} else {
			continue
		}
	}
	return fanChart, nil
}
