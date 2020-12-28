import requests as rq
import json
import sqlite3
import os


def getfile(chart_id, diff="expert"):
    if chart_id <= 500 or chart_id == 1000 or chart_id == 1001:
        url = "https://player.banground.fun/api/bestdori/official/map/" + \
              str(chart_id) + "/" + diff
        res = rq.get(url)
        try:
            chartres = res.json()
            if chartres["result"]:
                return True, json.dumps(chartres["data"]), None, None
        except Exception as exception:
            print(exception)
            return False, None, None, None
    else:
        url = "https://player.banground.fun/api/bestdori/community/" + str(chart_id)
        res = rq.get(url)
        try:
            chartres = res.json()
            if chartres["result"]:
                return (True, json.dumps(chartres["data"]["notes"]), chartres["data"]["difficulty"],
                        chartres["data"]["level"])
        except Exception as exception:
            print(exception)
            return False, None, None, None
    return False, None, None, None


def rtr_file_url(chart_id, diff="expert"):
    # print(id,diff)
    if chart_id <= 500 or chart_id == 1000 or chart_id == 1001:
        return os.path.join("chart", str(chart_id) + "." + diff + ".json")
    else:
        return os.path.join("chart", str(chart_id) + ".json")


def buffer_file(chart_id, diff="expert"):
    status, chart, chart_diff, chart_level = getfile(chart_id, diff)
    if chart_id >= 500 and chart_id != 1000 and chart_id != 1001 and chart_diff is not None:
        con = sqlite3.connect('chartLevel.db')
        cur = con.cursor()
        cur.execute("CREATE TABLE IF NOT EXISTS songList(id INTEGER PRIMARY KEY, diff INTEGER,level INTEGER)")
        con.commit()
        cur.execute("INSERT INTO songList VALUES (?,?,?)", (chart_id, chart_diff, chart_level))
        con.commit()
        con.close()
    if status:
        file = open(rtr_file_url(chart_id, diff), "w")
        file.write(chart)
        file.close()
    return chart


def get_chart(chart_id, diff="expert"):
    try:
        file = open(rtr_file_url(chart_id, diff), "r")
    except Exception as exception:
        print(exception)
        return buffer_file(chart_id, diff)
    chart = file.read()
    file.close()
    return chart


def get_basic(chart_id, flag=True):
    con = sqlite3.connect('chartLevel.db')
    cur = con.cursor()
    cur.execute("CREATE TABLE IF NOT EXISTS songList(id INTEGER PRIMARY KEY, diff INTEGER,level INTEGER)")
    con.commit()
    cur.execute("SELECT * FROM songList where id = ?", (chart_id,))
    info = cur.fetchall()
    con.close()
    if len(info) == 0 and flag:
        buffer_file(chart_id)
        # time.sleep(2)
        return get_basic(chart_id, False)
    elif len(info) == 0 and not flag:
        return None
    return info[0][1]

# print(get_basic(9219))
