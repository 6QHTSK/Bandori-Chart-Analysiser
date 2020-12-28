import sqlite3
import time
import json
import datetime
import fetch

last_fetch_time = 0


def return_bd_any():
    con = sqlite3.connect("songlist.db")
    cur = con.cursor()
    result = {}
    heatsong = []
    newheatsong = []
    heatartist = []
    newheatartist = []
    # sys.stdout = io.TextIOWrapper(sys.stdout.buffer,encoding='utf8')
    cur.execute("SELECT max(time) from songlist")
    timearray = time.localtime(cur.fetchall()[0][0] / 1000)
    result["time"] = time.strftime("UTC+8 %Y-%m-%d %H:%M:%S", timearray)
    cur.execute(
        "SELECT diff, count(*) COUNT from songlist where diff in (select diff from songlist group by diff) and new = 1 group by diff order by diff asc")
    result["diffcounter"] = cur.fetchall()
    cur.execute(
        "SELECT level, count(*) COUNT from songlist where level in (select level from songlist group by level) and new = 1 group by level order by level asc")
    result["levelcounter"] = cur.fetchall()
    cur.execute(
        "SELECT username, nickname, count(*) COUNT from songlist where username in (select username from songlist group by username) and new = 1 group by username order by COUNT desc limit 30")
    result["charterdetails"] = cur.fetchall()
    cur.execute("SELECT username, nickname, id from songlist where new = 1 group by username having COUNT(*)>5")
    result["usernames"] = cur.fetchall()
    result["chartercount"] = len(result["usernames"])
    cur.execute("SELECT count(*) FROM songlist")
    result["chartcount"] = cur.fetchall()[0][0]
    cur.execute("SELECT count(*) FROM songlist WHERE new = 1")
    result["validchartcount"] = cur.fetchall()[0][0]
    cur.execute("SELECT id, username, nickname, title, artists, likes from songlist order by likes desc limit 30")
    result["likechartdetail"] = cur.fetchall()
    cur.execute(
        "SELECT username, nickname, sum(likes) SUM from songlist where username in (select username from songlist group by username) group by username order by SUM desc limit 30")
    result["likecharterdetail"] = cur.fetchall()
    cur.execute(
        "SELECT title, count(*) COUNT from songlist where title in (select title from songlist where new = 1 group by title) group by title order by COUNT desc limit 40")
    # result["heatedsongdetail"] = cur.fetchall()
    heatsong = cur.fetchall()
    cur.execute(
        "SELECT artists, count(*) COUNT from songlist where artists in (select artists from songlist where new = 1 group by artists) group by artists order by COUNT desc limit 40")
    # result["heatedartistdetail"] = cur.fetchall()
    heatartist = cur.fetchall()
    cur.execute(
        "SELECT username, nickname, avg(level) AVG from songlist where username in (select username from songlist where new = 1 and level > 20 and diff> 2 group by username having COUNT(*)>5) and new = 1 and level > 20 and diff> 2 group by username order by AVG desc limit 30")
    result["highdiffcharterdetail"] = cur.fetchall()
    cur.execute(
        "SELECT username, nickname, avg(level) AVG from songlist where username in (select username from songlist where new = 1 and level > 20 and diff> 2  group by username having COUNT(*)>5) and new = 1 and level > 20 and diff> 2 group by username order by AVG asc limit 30")
    result["lowdiffcharterdetail"] = cur.fetchall()
    cur.execute(
        "SELECT id, username, nickname, title, artists, nps from songlist where new = 1 order by nps desc limit 30")
    result["highestnpschart"] = cur.fetchall()
    cur.execute(
        "SELECT id, username, nickname, title, artists, songlen from songlist where new = 1 order by songlen desc limit 30")
    result["longestchart"] = cur.fetchall()
    cur.execute(
        "SELECT id, username, nickname, title, artists, notes from songlist where new = 1 order by notes desc limit 30")
    result["maxnotechart"] = cur.fetchall()
    for song in heatsong:
        cur.execute("SELECT count(*) from songlist where title like ?", ("%" + song[0] + "%",))
        newheatsong.append([song[0], cur.fetchone()[0]])
    result["heatedsongdetail"] = sorted(newheatsong, key=lambda x: (x[1]), reverse=True)[0:20]
    for artist in heatartist:
        cur.execute("SELECT count(*) from songlist where artists like ?", ("%" + artist[0] + "%",))
        newheatartist.append([artist[0], cur.fetchone()[0]])
    result["heatedartistdetail"] = sorted(newheatartist, key=lambda x: (x[1]), reverse=True)[0:20]
    timeresult = {2019: {}}
    likesperchartresult = {2019: {}}
    year = 2019
    month = 9
    while (True):
        date1 = datetime.date(year, month, 1)
        if month + 1 == 13:
            date2 = datetime.date(year + 1, 1, 1)
        else:
            date2 = datetime.date(year, month + 1, 1)
        starttime = int(time.mktime(date1.timetuple()) * 1000)
        endtime = int(time.mktime(date2.timetuple()) * 1000)
        cur.execute("SELECT count(*) FROM songlist WHERE time > ? and time < ?", (starttime, endtime))
        ans = cur.fetchall()[0][0]
        cur.execute("SELECT sum(likes) FROM songlist WHERE time > ? and time < ?", (starttime, endtime))
        likes = cur.fetchall()[0][0]
        if ans == 0:
            break
        timeresult[year][month] = ans
        likesperchartresult[year][month] = round(likes / ans, 2)
        if month == 12:
            month = 1
            year = year + 1
            timeresult[year] = {}
            likesperchartresult[year] = {}
        else:
            month = month + 1
    result["timecaculate"] = timeresult
    result["likeperchart"] = likesperchartresult
    con.close()
    return json.dumps(result)


def return_author(author):
    def findauthor1(thelist, username):
        i = 0
        while i < len(thelist):
            if thelist[i][0] == username:
                return (i + 1, thelist[i][1])
            i = i + 1
        return (0, 0)

    def findauthor2(thelist, username):
        i = 0
        res = []
        while i < len(thelist) and len(res) < 5:
            if thelist[i][1] == username:
                res.append((i + 1,) + thelist[i])
            i = i + 1
        return res

    result = {"result": True}
    con = sqlite3.connect("songlist.db")
    cur = con.cursor()
    try:
        cur.execute("SELECT nickname from songlist where username = ? group by username having COUNT(*)>5", (author,))
        result["username"] = author
        result["nickname"] = cur.fetchone()[0]
    except:
        return json.dumps({"result": False})
    cur.execute(
        "SELECT username, count(*) COUNT from songlist where username in (select username from songlist group by username) and new = 1 group by username order by COUNT desc")
    result["chartcount"] = findauthor1(cur.fetchall(), author)
    cur.execute(
        "SELECT username, sum(likes) SUM from songlist where username in (select username from songlist group by username) group by username order by SUM desc")
    result["likecount"] = findauthor1(cur.fetchall(), author)
    try:
        cur.execute(
            "SELECT username, avg(level) AVG from songlist where username in (select username from songlist where new = 1 and level > 20 and diff> 2 group by username having COUNT(*)>5) and new = 1 and level > 20 and diff> 2 group by username order by AVG desc")
        result["highdiffcount"] = findauthor1(cur.fetchall(), author)
        cur.execute(
            "SELECT username, avg(level) AVG from songlist where username in (select username from songlist where new = 1 and level > 20 and diff> 2 group by username having COUNT(*)>5) and new = 1 and level > 20 and diff> 2 group by username order by AVG asc")
        result["lowdiffcount"] = findauthor1(cur.fetchall(), author)
    except:
        result["highdiffcount"] = (0, "-")
        result["lowdiffcount"] = (0, "-")
    cur.execute("SELECT id, username, title, artists, nps from songlist where new = 1 order by nps desc")
    result["npscount"] = findauthor2(cur.fetchall(), author)
    cur.execute("SELECT id, username, title, artists, songlen from songlist where new = 1 order by songlen desc")
    result["lencount"] = findauthor2(cur.fetchall(), author)
    cur.execute("SELECT id, username, title, artists, notes from songlist where new = 1 order by notes desc")
    result["notecount"] = findauthor2(cur.fetchall(), author)
    cur.execute("SELECT id, username, title, artists, likes from songlist order by likes desc")
    result["likechartcount"] = findauthor2(cur.fetchall(), author)
    cur.execute("SELECT time from songlist where username = ? and new = 1 order by time desc", (author,))
    result["lastupdate"] = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(cur.fetchone()[0] / 1000))
    cur.execute(
        "SELECT id, username, title, artists, datetime(time / 1000, 'unixepoch','localtime') from songlist where new = 1 order by time desc")
    result["lastupdatechart"] = findauthor2(cur.fetchall(), author)
    con.close()
    return json.dumps(result)


def return_search(searchString):
    con = sqlite3.connect("songlist.db")
    cur = con.cursor()
    global last_fetch_time
    if time.time() - last_fetch_time >= 10.0:
        fetch.fetch_basically()
        last_fetch_time = time.time()

    def beautify(tur):
        return {
            "id": tur[0],
            "title": tur[1],
            "artists": tur[2],
            "level": tur[3],
            "diff": tur[4],
            "username": tur[5],
            "nickname": tur[6],
            "cover": tur[7],
            "bgm": tur[8]
        }

    def combination(parts):
        res = []
        length = len(parts)
        counter = {}
        tmp = {}
        for part in parts:
            for item in part:
                if item[0] not in counter:
                    tmp[item[0]] = item
                    counter[item[0]] = 1
                else:
                    counter[item[0]] = counter[item[0]] + 1
        for key, values in counter.items():
            if values == length:
                res.append(beautify(tmp[key]))
        res.sort(key=lambda x: x["id"], reverse=True)
        return res

    def searchDiff(diff):
        cur.execute("SELECT * from basicSonglist where diff = ?", (diff,))
        return cur.fetchall()

    def searchLevel(level):
        cur.execute("SELECT * from basicSonglist where level = ?", (level,))
        return cur.fetchall()

    def searchID(ID):
        cur.execute("SELECT * from basicSonglist where id = ?", (ID,))
        return cur.fetchall()

    def searchText(text):
        sqlText = "%" + text + "%"
        cur.execute(
            "SELECT * from basicSonglist where LOWER(title) LIKE ? or LOWER(artists) LIKE ? or LOWER(username) LIKE ? or LOWER(nickname) LIKE ?",
            (sqlText, sqlText, sqlText, sqlText))
        return cur.fetchall()

    def searchAll():
        cur.execute(
            "SELECT * from basicSonglist"
        )
        return cur.fetchall()

    if searchString == "":
        return {"result": combination((searchAll(),))}
    searchStrings = searchString.split()
    parts = []
    diffstr = ["easy", "normal", "hard", "expert", "special", "ez", "nm", "hd", "ex", "sp"]
    for string in searchStrings:
        string = string.lower()
        if string.isdigit():
            num = int(string)
            if num > 900:
                parts.append(searchID(num))
            elif 5 <= num <= 30:
                parts.append(searchLevel(num))
            elif 0 <= num <= 4:
                parts.append(searchDiff(num))
            else:
                parts.append(searchText(string))
        elif string in diffstr:
            for i in range(0, 10):
                if diffstr[i] == string:
                    parts.append(searchDiff(i % 5))
                    break
        else:
            parts.append(searchText(string))
    con.close()
    return {"result": combination(parts)}
