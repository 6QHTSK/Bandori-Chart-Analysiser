import flask
import flask_cors
import bestdoriany
import fetch

server = flask.Flask(__name__)
flask_cors.CORS(server)


@server.route('/calcData', methods=["get", "post"])
def calc_data():
    return bestdoriany.return_bd_any()


@server.route('/calcAuthor', methods=["get", "post"])
def calc_author():
    author = flask.request.values.get("author")
    return bestdoriany.return_author(author)


@server.route('/bestdoriSearch')
@server.route('/bestdoriSearch/<searchString>', methods=["get", "post"])
def search(searchString=""):
    page = flask.request.values.get("page")
    fetch.fetch_basically()
    list = bestdoriany.return_search(searchString)
    pageCount = int((len(list) - 1) / 20) + 1
    if page is None:
        page = 0
    else:
        page = int(page)
    start = page * 20
    end = min(len(list), start + 21)
    res = {"pageCount": pageCount, "list": list[start:end]}
    return res


server.run(port=20009, host='0.0.0.0')
