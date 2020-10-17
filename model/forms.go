package model

type Chart struct {
	Notes    []Note `bson:"chart"`
	Level    int    `bson:"level"`
	AuthorID int    `bson:"authorID"`
	Artist   string `bson:"artist"`
	Title    string `bson:"title"`
	ID       int    `bson:"id"`
	Diff     int    `bson:"diff"`
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
		Name       string `json:"string"`
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

type Note struct {
	Type   string  `json:"type"`
	CMD    string  `json:"cmd"`
	BPM    float32     `json:"bpm"`
	Beat   float32 `json:"beat"`
	Effect string  `json:"effect"`
	Time   float32 `json:"time"`
	Note   string  `json:"note"`
	Lane   int     `json:"lane"`
	Skill  bool    `json:"skill"`
	Start  bool    `json:"start"`
	End    bool    `json:"end"`
	Flick  bool    `json:"flick"`
	Charge bool    `json:"charge"`
	Pos    string  `json:"pos"`
}
