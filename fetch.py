import requests
import time
import math
import sqlite3
import json

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
    "CREATE TABLE IF NOT EXISTS songlist(id INTEGER PRIMARY KEY, title TEXT,artists TEXT,level INTEGER,diff INTEGER,"
    "time INTEGER,username TEXT,nickname TEXT,likes INTEGER,new INTEGER,songlen REAL,notes INTEGER,nps REAL)")
cur.execute(
    "CREATE TABLE IF NOT EXISTS basicSonglist(id INTEGER PRIMARY KEY, title TEXT,artists TEXT,level INTEGER,diff INTEGER,"
    "username TEXT,nickname TEXT,cover TEXT, bgm TEXT)")
con.close()


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
    cur.execute(
        "SELECT max(id) FROM songlist WHERE title = ? and diff = ? and username = ?",
        (title, diff, username))
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
            cur.execute(
                "UPDATE songlist SET level = ?,diff = ?, likes = ?, new = ? , nickname = ? WHERE id = ?",
                (song["level"], song["diff"], song["likes"], new, song["author"]["nickname"], song["id"]))
            con.commit()
        else:
            flag = True
            while flag:
                try:
                    diffres = \
                        requests.get(url="http://api.ayachan.fun/DiffAnalysis?speed=-1.0&id=" + str(song["id"])).json()[
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


def fetch_all():
    global con, cur
    con = sqlite3.connect("songlist.db")
    cur = con.cursor()
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
            checksonglistBasic(songlist)
            offset = offset + 50
            blacklist()
            print("{}% {}/{}".format(math.floor(offset / totalpost * 100), offset, totalpost))
        except requests.exceptions.RequestException as e:
            print(e)
            print("retrying")
    con.close()


def isintheBasiclist(id):
    cur.execute("SELECT * FROM basicSonglist WHERE id = ?", (id,))
    song = cur.fetchone()  # 找到相同键值
    return song != None


def checksonglistBasic(songlist):
    counter = 0
    for song in songlist:
        if isintheBasiclist(song["id"]):
            cur.execute(
                "UPDATE basicSonglist SET level = ?,diff = ? WHERE id = ?",
                (song["level"], song["diff"], song["id"]))
            con.commit()
        else:
            counter = counter + 1
            print("Add song {} into basic".format(song["id"]))
            while song["song"]["type"] != "custom":
                try:
                    url = 'https://servers.sonolus.com/0.4.8/bestdori/levels/list.json'
                    res = requests.get(url, params={"search": song["id"]}, timeout=10)
                    resjson = res.json()
                    song["song"]["type"] = "custom"
                    song["song"]["cover"] = resjson["list"][0]["cover"]
                    song["song"]["audio"] = resjson["list"][0]["bgm"]
                except:
                    print("retrying")
            try:
                cur.execute("INSERT INTO basicSonglist VALUES (?,?,?,?,?,?,?,?,?)", (
                    song["id"], song["title"], song["artists"], song["level"], song["diff"],
                    song["author"]["username"], song["author"]["nickname"], song["song"]["cover"],
                    song["song"]["audio"]))
            except:
                pass
            con.commit()
    return counter == len(songlist)


def fetch_basically():
    global con, cur
    con = sqlite3.connect("songlist.db")
    cur = con.cursor()
    headers = {
        'Content-Type': 'application/json'
    }
    data = {
        "following": False,
        "categoryName": "SELF_POST",
        "categoryId": "chart",
        "order": "TIME_DESC",
        "limit": 10,
        "offset": 0
    }
    url = 'https://bestdori.com/api/post/list'
    res = requests.post(url=url, json=data, headers=headers)
    resjson = res.json()
    songlist = resjson["posts"]
    totalpost = resjson["count"]
    offset = 10
    flag = checksonglistBasic(songlist)
    while flag and offset < totalpost:
        data["offset"] = offset
        try:
            res = requests.post(url=url, json=data, headers=headers, timeout=15)
            resjson = res.json()
            songlist = resjson["posts"]
            flag = checksonglistBasic(songlist)
            offset = offset + 10
            print(offset)
        except requests.exceptions.RequestException as e:
            print(e)
            print("retrying")
    con.close()


if __name__ == "__main__":
    fetch_all()
# output = open("Z:\\res3.json","w",encoding='utf8')
# for song in songlist:
#    print("{} - {} LV. {}".format(song["id"],song["title"],song["level"]))
# outputjsons = []
