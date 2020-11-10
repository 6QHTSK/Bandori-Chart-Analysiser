import requests
import time
import math
import sqlite3

authorlist = {}
authornickname = {}
totalsonglist = []
titleblacklist = ["%test%", "%テスト%", "%てすと%"]
sameartist = [("Pastel*Palettes", "Pastel＊Palettes"), ("Poppin' Party", "Poppin'Party"), ("roselia", "Roselia"),
              ("Sakuzyo", "削除"), ("ハロー、ハッピーワールド！", "Hello, Happy World!")]
idblacklist = [19030, 11949, 18637]
# 初始化数据库
con = sqlite3.connect("songlist.db")
cur = con.cursor()
cur.execute(
    "CREATE TABLE IF NOT EXISTS songlist(id INTEGER PRIMARY KEY, title TEXT,artists TEXT,level INTEGER,diff INTEGER,time INTEGER,username TEXT,nickname TEXT,likes INTEGER,new INTEGER,songlen REAL,notes INTEGER,nps REAL)")


def issame(dict1, dict2):
    if dict1["artists"] == dict2["artists"] and dict1["author"] == dict2["author"] and dict1["diff"] == dict2[
        "diff"] and dict1["id"] == dict2["id"] and dict1["level"] == dict2["level"] and dict1["likes"] == dict2[
        "likes"] and dict1["nickname"] == dict2["nickname"] and dict1["time"] == dict2["time"] and dict1["title"] == \
            dict2["title"]:
        return True
    else:
        return False


def isinthelist(id):
    cur.execute("SELECT * FROM songlist WHERE id = ?", (id,))
    song = cur.fetchone()  # 找到相同键值
    return song != None

def isnewest(title, diff, username, id):
    cur.execute("SELECT max(id) FROM songlist WHERE title = ? and diff = ? and username = ?", (title, diff, username))
    chartid = cur.fetchall()
    return id == chartid[0][0] or chartid[0][0] == None


def checksonglist(songlist):
    i = 0
    for song in songlist:
        new = isnewest(song["title"], song["diff"], song["author"]["username"], song["id"])
        for artist in sameartist:
            if song["artists"] == artist[0]:
                song["artists"] = artist[1]
        if isinthelist(song["id"]):
            cur.execute("UPDATE songlist SET level = ?,diff = ?, likes = ?, new = ? , nickname = ? WHERE id = ?",
                        (song["level"], song["diff"], song["likes"], new, song["author"]["nickname"], song["id"]))
            con.commit()
        else:
            flag = True
            while flag:
                try:
                    diffres = requests.get(url="http://localhost:20008/DiffAnalysis?speed=-1.0&id=" + str(song["id"])).json()[
                        "detail"]
                    flag = False
                    print("Add song {}".format(song["id"]))
                except:
                    print("Retrying {} with details:".format(song["id"]))
                    time.sleep(1)
                    flag = True
            try:
                cur.execute("INSERT INTO songlist VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)", (
                song["id"], song["title"], song["artists"], song["level"], song["diff"], song["time"],
                song["author"]["username"], song["author"]["nickname"], song["likes"], new, diffres["TotalTime"],
                diffres["TotalNote"], diffres["TotalNPS"]))
                con.commit()
            except:
                print("Caculation failed at {}".format(song["id"]))
        i = i + 1


def blacklist():
    for title in titleblacklist:
        cur.execute("UPDATE songlist SET new = 0 WHERE title like ? or id = 0", (title,))
        con.commit()
        cur.execute("UPDATE songlist SET new = 0 WHERE artists like ? or id = 0", (title,))
        con.commit()
    cur.execute("UPDATE songlist set new = 0 where songlen < 50 or id = 0")
    con.commit()
    cur.execute("UPDATE songlist set new = 0 where notes < 30 or id = 0")
    con.commit()
    for id in idblacklist:
        cur.execute("UPDATE songlist SET new = 0 WHERE id = ? or id = 0", (id,))
        con.commit()


headers = {
    'Content-Type': 'application/json'
}
data = {
    "following": False,
    "categoryName": "SELF_POST",
    "categoryId": "chart",
    "order": "TIME_DESC",
    "limit": 50,
    "offset": 0
}
url = 'https://bestdori.com/api/post/list'
res = requests.post(url=url, json=data, headers=headers)
resjson = res.json()
totalpost = resjson["count"]
# totalpost = 100
songlist = resjson["posts"]
offset = 50
checksonglist(songlist)
print("{}% {}/{}".format(math.floor(offset / totalpost * 100), offset, totalpost))
while offset < totalpost:
    time.sleep(10)
    data["offset"] = offset
    try:
        res = requests.post(url=url, json=data, headers=headers, timeout=15)
        resjson = res.json()
        songlist = resjson["posts"]
        checksonglist(songlist)
        offset = offset + 50
        blacklist()
        print("{}% {}/{}".format(math.floor(offset / totalpost * 100), offset, totalpost))
    except requests.exceptions.RequestException as e:
        print(e)
        print("retrying")
con.close()

# output = open("Z:\\res3.json","w",encoding='utf8')
# for song in songlist:
#    print("{} - {} LV. {}".format(song["id"],song["title"],song["level"]))
# outputjsons = []
