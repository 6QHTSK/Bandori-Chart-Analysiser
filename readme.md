# 工具站后端

## /DiffAnalysis

### 输入数据

GET 方法参数：

"id": 谱面id（官谱/自制谱，优先官谱）

"diff": 指定官谱的难度（0~4）

"speed": 正浮点数，指定游玩时的倍速，默认1.0（即1.0倍速100%，若开启dt此处为1.5即150%倍速）

例： <https://api.ayachan.fun/DiffAnalysis?id=1&diff=1&speed=2.0>

请求id为1的谱面，normal难度下，开200%倍速的谱面难度（以normal难度下标级）

返回数据示例：

```json
{
    "author": {
        "AuthorID": 55,                 //内部ID
        "UserName": "psk2019",          //谱师用户名
        "NickName": "稳音绫与6QHTSK"    //谱师昵称
    },
    "basic": {
        "ID": 9220,                     //谱面ID
        "Diff": 3,                      //等级
        "Level": 29,                    //难度
        "AuthorID": 55,                 //谱师内部ID
        "Artist": "ナナヲアカリ",       //作曲家
        "Title": "[FULL] One room sugar life",//曲名
        "Notes": null
    },
    "detail": {
        "ActiveHPS": 8.121212,          //主要部分HPS
        "TotalTime": 162.76668,         //总时间
        "FingerMaxHPS": 6,              //单手最大HPS
        "TotalNPS": 8.109768,           //总NPS
        "LeftPercent": 0.49242425,      //左手占比
        "MaxSpeed": 5.618928,           //最大移动速度
        "ID": 9220,                     //谱面ID
        "Error": "",                    //分析过程中的错误信息
        "Diff": 3,
        "ActiveNPS": 9.752294,          //主要部分NPS
        "TotalHitNote": 1008,           //总需要击打的音符数
        "ActivePercent": 0.6898734,     //主要部分占比
        "FlickNoteInterval": 5.9999084, //粉键-普通键间隔的倒数
        "MainBPM": 180,                 //主要BPM
        "TotalNote": 1320,              //物量
        "BPMLow": 180,                  //最低BPM
        "NoteFlickInterval": 6.0000114, //普通键-粉键间隔倒数
        "BPMHigh": 180,                 //最大BPM
        "MaxScreenNPS": 12,             //最大瞬时NPS
        "TotalHPS": 6.1929135           //总HPS
    },
    "diff": { // 根据机器学习的结果，选择最有代表性的几项作难度计算，符号见上
        "FingerMaxHPS": 27.652174,
        "TotalNPS": 27.608696,
        "FlickNoteInterval": 26.653913,
        "NoteFlickInterval": 28.9,
        "MaxScreenNPS": 28.3,
        "TotalHPS": 26.225563,
        "BlueWhiteFunc": 27.062134,
        "MaxSpeed": 28.130001
    },
    "result": true // 一般是True，否则服务器错误
}
```

**注意**，原先文档的null在这里体现为0值（换了框架）

## /calcAuthor 计算某位谱师的发谱情况，需要发5张谱才计数

输入参数： author: 谱师的用户名

例如：<https://api.ayachan.fun/calcAuthor?author=psk2019>

输出参数：

```json
{
    "chartcount": [
        8, // 发谱数量的位次
        81 // 发谱的数量
    ],
    "highdiffcount": [
        245, // 平均难度的位次，由高到低
        25.545454545454547 // 平均难度
    ],
    "lastupdate": "2020-11-01 13:43:10", // 最后一次上传谱面时间（咕咕咕）
    "lastupdatechart": [
        [
            141,    // 0 排名
            30304,  // 1 ID
            "psk2019",
            "Heart+Heart",  // 3 曲名
            "一柳梨璃&白井夢結（CV：赤尾ひかる&夏吉ゆうこ）", // 4 艺术家
            "2020-11-01 13:43:10" // 5 更新时间/长度等特定项
        ],
        // ...
    ],
    "lencount": [
        // 最长谱面
    ],
    "likechartcount": [
        // 最受喜爱谱面
    ],
    "likecount": [
        11, // 获赞排名
        262 // 获赞数
    ],
    "lowdiffcount": [
        17, // 平均难度的位次，由低到高
        25.545454545454547 // 平均难度
    ],
    "nickname": "稳音绫与6QHTSK", // 谱师昵称
    "notecount": [
        // 最多音符谱面
    ],
    "npscount": [
        // 最高nps谱面
    ],
    "result": true, // 返回结果的正确性
    "username": "psk2019" // 谱师名
}

```
