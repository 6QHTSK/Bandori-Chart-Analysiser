import flask
import flask_cors
import bestdoriany

server = flask.Flask(__name__)
flask_cors.CORS(server)

"""
@server.route('/DiffAnalysis', methods=['get', 'post'])
def diff_analysis():
    db = database.Database()
    if flask.request.method == "POST":
        post_data = flask.request.get_json()
        level_num = post_data["diff"]
        chart = post_data["data"]
        details = db.get_chart_details_from_chart(None, level_num, chart)
        diffs = db.diff_calculate(details)
        return json.dumps({'result': True, 'basic': None, 'detail': details, 'diff': diffs})
    else:
        try:
            chart_id = flask.request.values.get('id')
            diff = flask.request.values.get("diff")
            if diff is None:
                diff = 3
            print(chart_id, diff)
            details = db.chart_query(int(chart_id), int(diff))
            # del details["basic"]["chart"]
            print(details)
            return json.dumps(details)
        except Exception as exception:
            print(exception)
            return json.dumps({"result": False})


@server.route('/bdOffToBdFan', methods=["get", "post"])
def bd_off_to_bd_fan():
    chart_id = flask.request.values.get("id")
    diff = flask.request.values.get("diff")
    details = db.chart_query(chart_id, diff)
    if details["result"] is None:
        return {"result": False}
    else:
        return {"result": True, "data": details["basic"]["chart"]}


@server.route("/", methods=["get", "post"])
@server.route("/index.htm", methods=["get", "post"])
@server.route("/index.html", methods=["get", "post"])
@server.route("/index", methods=["get", "post"])
def show_index():
    page = open('index2.html', encoding='utf-8')  # ——---->打开当前文件下的my_index.html(这个html是你自己写的)
    res = page.read()  # ------>读取页面内容，并转义后返回
    return res


@server.route("/favicon.ico", methods=["get", "post"])
def show_icon():
    return flask.send_from_directory('C:\\python', 'favicon.ico', mimetype='image/vnd.microsoft.icon')

"""


@server.route('/calcData', methods=["get", "post"])
def calc_data():
    return bestdoriany.return_bd_any()


@server.route('/calcAuthor', methods=["get", "post"])
def calc_author():
    author = flask.request.values.get("author")
    return bestdoriany.return_author(author)


"""
@server.route('/songList', methods=["get", "post"])
def return_song_list():
    song_list = open('songList.json', encoding='utf-8')
    return json.dumps(json.loads(song_list.read()))


@server.route('/bdChart', methods=["get", "post"])
def return_bestdori_chart():
    chart_id = flask.request.values.get("id")
    diff = flask.request.values.get("diff")
    if diff is None:
        chart = bestdorichartdata.get_chart(int(chart_id))
    else:
        chart = bestdorichartdata.get_chart(int(chart_id), diff)
    if chart is None:
        return json.dumps({"result": False})
    else:
        return json.dumps({"result": True, "data": json.loads(chart)})


@server.route('/bdBasic', methods=["get", "post"])
def return_bestdori_basic():
    chart_id = int(flask.request.values.get('id'))
    if chart_id is not None and chart_id > 500:
        res = bestdorichartdata.get_basic(int(chart_id))
        if res is not None:
            return json.dumps({"result": True, "level": res})
        else:
            return json.dumps({"result": False})
    else:
        return json.dumps({"result": False})

"""
server.run(port=20009, host='0.0.0.0')
