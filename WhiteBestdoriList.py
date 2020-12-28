import time
import random
import flask
import requests
import zlib
import bestdoriany

server = flask.Flask(__name__)
diffstr = ["EZ", "NM", "HD", "EX", "SP"]


@server.route('/level.json', methods=["get", "post"])
def func():
    def combineauthor(username, nickname):
        if nickname is None:
            return username
        else:
            return nickname + "@" + username

    page = flask.request.values.get("page")
    search = flask.request.values.get("search")
    customcover = True
    nocover = False

    if search != None:
        search = search.lower()
        searchString = search.split()
        search = ""
        for string in searchString:
            if string == "nocover":
                customcover = False
                nocover = True
            else:
                search = "{} {}".format(search, string)
    else:
        search = ""
    passed = True
    rqjson = {}
    while(passed):
        try:
            rqjson = bestdoriany.return_search(search)["result"]
            passed = False
        except:
            passed = True
            time.sleep(0.01 * random.randint(75,125))
            print("Retry searching")
    pageCount = int((len(rqjson) - 1) / 20) + 1
    res = {"pageCount": pageCount, "list": []}
    if page is None:
        page = 0
    else:
        page = int(page)
    start = page * 20
    end = min(len(rqjson), start + 21)
    for i in range(start, end):
        item = rqjson[i]
        item["bgm"] = item["bgm"].replace("https://bestdori.com/", "http://jiashule.sonolus.reikohaku.fun/bestdori/")
        if customcover:
            item["cover"] = item["cover"].replace("https://bestdori.com/",
                                                  "http://jiashule.sonolus.reikohaku.fun/bestdori/")
        elif nocover:
            item["cover"] = "https://assets.ayachan.fun/pic/black.png"
        else:
            if item["cover"][0:21] == "https://bestdori.com/":
                item["cover"] = item["cover"].replace("https://bestdori.com/",
                                                      "http://jiashule.sonolus.reikohaku.fun/bestdori/")
            else:
                item["cover"] = "https://assets.ayachan.fun/pic/black.png"
        title = "[{}] {} [{}.{}]".format(str(item["id"]), item["title"], diffstr[item["diff"]], item["level"])
        res["list"].append({
            "artists": item["artists"],
            "author": combineauthor(item["username"], item["nickname"]),
            "bgm": item["bgm"],
            "cover": item["cover"],
            "title": title,
            "difficulties": [
                {
                    "name": str(item["id"]),
                    "rating": str(item["level"])
                }
            ]
        })

    return res, 200


"""
item["cover"] = "https://assets.sonolus.ayachan.fun/pic/" + \
                        str(zlib.compress(bytes(item["cover"], encoding="utf-8")).hex())
        item["bgm"] = "https://assets.sonolus.ayachan.fun/song/" + \
                      str(zlib.compress(bytes(item["bgm"], encoding="utf-8")).hex())
"""


@server.route('/pic/<compress_url>', methods=["get", "post"])
@server.route('/song/<compress_url>', methods=["get", "post"])
def uncompress(compress_url):
    try:
        url = zlib.decompress(bytes.fromhex(compress_url))
        return flask.redirect(url)
    except:
        return "", 404


@server.route('/404', methods=["get", "post"])
def request404():
    return flask.redirect("https://assets.ayachan.fun/pic/black.png")


server.run(port=21000, host='0.0.0.0')
