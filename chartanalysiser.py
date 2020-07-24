import json
import math
import requests as rq
import sqlite3
from time import process_time
from functools import cmp_to_key
'''bestdori 音符类解释
所有音符类共有元素: beat : <此音符的拍数>
    1. BPM 变化音符：必定含有 type : System , cmd: BPM, bpm : <此BPM变化音符变化的BPM> 
    2. 一般音符类：均含有 type : Note，lane : <该音符所在轨道号> , 
        (1). 单点类音符：note: Single
            i. 粉键： flick : True
            ii. 蓝键：无额外内容
        (2). 绿条类音符: note: Slide pos: <A,B等用于区分绿条的编号>
            i. 起始键：start: True
            ii. 中间节点： 无额外内容
            iii. 尾节点： end: True
            iv. 尾粉：end: True, flick : True
'''
mainbpm = 120
totalnote = 0
def calctime(chartjson):
    '''计算bestdori格式的各个note的时间；
    输入：bestdori格式的谱面；
    返回值：含时间的纯净note列表，最低 bpm， 最高bpm
    '''
    global mainbpm
    global totalnote
    def cmp(note1,note2):
        '''排序notelist使用的cmp函数'''
        if note1["beat"] == note2["beat"]:
            #拍数相同，判断轨道
            return note1["lane"] - note2["lane"]
        else:
            return note1["beat"] - note2["beat"]
    bpm = 60 # 谱面的实时BPM
    offsettime = 0.0 # 上一个BPM 变化点的时间
    offsetbeat = 0 # 上一个BPM 变化点的拍数
    notelist = [] # 纯净的，全部是note的列表，返回值
    bpmchart = [] # 记录bpm的chart
    bpmlow = math.inf # 最小bpm
    bpmhigh = -1.0 #最大bpm，不存在负bpm
    istherenote = False # 排除前面用来改变offset用的bpm变化点
    maxoffsettime = -1.0
    totalnote = 0
    for note in chartjson :
        if note["type"] == "System" :
            # 此时应为BPM 变化点
            lastoffsettime = offsettime
            offsettime = (note["beat"] - offsetbeat) * (60.0 / bpm) + offsettime # 计算与上一个BPM变化点的拍数，结合上一个BPM算出此时的时间
            bpm = abs(note["bpm"]) # 更新实时BPM，为防止负BPM搞事情加绝对值
            offsetbeat = note["beat"] # 更新上一个BPM的拍数
            bpmchart.append(note)
            if (istherenote == False) :
                # 检测到是用来改变offset用的BPM变化点
                bpmlow = math.inf
                bpmhigh = -1.0
            if (bpm < bpmlow) :
                #更新bpm最小值
                bpmlow = bpm 
            if (bpm > bpmhigh) :
                #更新bpm最大值
                bpmhigh = bpm
            if offsettime - lastoffsettime > maxoffsettime:
                #计算主要bpm值
                mainbpm = bpm
                maxoffsettime = offsettime - lastoffsettime
        elif (note["type"] == "Note") :
            # 此时为正常的音符
            istherenote = True # 排除上一个BPM变化点是用来改变offset用的变化点
            note["time"] = (note["beat"] - offsetbeat) * (60.0 / bpm) + offsettime # 计算与上一个BPM变化点的拍数，结合上一个BPM算出此时的时间
            note["finger"] = 0 #后面用于音符处理时用的参数
            note["ignore"] = False
            totalnote = totalnote + 1 #计算总共的note数
            notelist.append(note) # 加入note列表以便后面使用
    notelist.sort(key=cmp_to_key(cmp))
    return ( notelist, bpmchart) #返回值：含时间的纯净note列表（以beat为顺序，相同时间以lane为顺序），最低 bpm， 最高bpm

def issameslide(*args):
    '''检查输入的音符组中是否为同一个滑条，音符组按时间顺序'''
    startflag = False
    endflag = False
    for note in args:
        if note["note"] != "Slide": # 如果不是滑条音符混入，那肯定不是
            return False
        if "start" in note and note["start"]: # 如果检测到开始音符（一般是第一个，不是就是两个滑条
            if startflag:
                return False
        elif "end" in note and note["end"]: # 如果检测到结束音符，做个标记
            endflag = True
        else:
            startflag = True # 第一个音符不是开始音符
            if endflag:
                return False # 结束音符后还有音符，是两个滑条
    return True

def erasetick(note1,note2,note3):
    '''输入一段绿条中的三个音符,判断其是否需要滑动来击中并标记
    输入：三个slide音符，输出，是否需要滑动来击中'''
    if note1["lane"] == note3["lane"] and abs(note2["lane"] - note1["lane"]) < 1 and issameslide(note1,note2,note3):
        # 满足起始、终止轨道相同，且中间键与上述键轨道数差距小于1
        note2["ignore"] = True #可忽略中间键的存在
        return True
    else:
        return False

class Finger:
    '''手指类，用于存放用这根手指击打的所有音符全集'''
    return_time = 0.5 # 两个note距离时间超出此即返回
    def __init__(self,default_lane):
        '''初始化手指类 输入：默认轨道'''
        self.default_lane = default_lane
        self.notelist = [] # 存放该手击打的所有音符
        self.hold = False # 该手是否击打绿条
        self.pos = "" # 存储击打绿条类型

    def position(self,beat,time):
        '''计算在某个beat、某个time时，该手指的位置。
        输入：拍数beat,时间time; 返回:浮点数lane'''
        if len(self.notelist) == 0:
            return self.default_lane # 没有击打音符，返回标准位置
        for i in range(0,len(self.notelist)):
            current = self.notelist[i]
            if current["beat"] == beat:
                return current["lane"] # 在该beat上有音符，则返回该音符的轨道值
            if current["beat"] > beat:
                if i == 0:
                    return self.default_lane # 第一个音符的时间已经超出当前beat值了，返回默认轨道
                
                previous = self.notelist[i-1]
                if previous["note"] == "Slide" and current["note"] == "Slide":
                    # 如果都是绿条的一部分
                    return previous["lane"] + (current["lane"] - previous["lane"]) * (time - previous["time"]) / (current["time"]-previous["time"]) 
                    # 当前轨道计算公式：起始轨道+ 轨道差值 * （当前时间差值/总时间差值)
                elif time - previous["time"] > self.return_time:
                    # 离上一个note远
                    return self.default_lane
                else:
                    # 离上一个note近
                    return previous["lane"]
        return self.default_lane # 为了不报WARNINGS就扔一个默认出口

    def available(self,note):
        '''判断击打输入的音符时，手指是否可用
        输入：bestdori音符，输出：布尔值，表示是否可用'''
        hold = False
        pos = ""
        for i in range(0,len(self.notelist)):
            n = self.notelist[i]
            if n["beat"] == note["beat"]:
                #如果有相同beat的音符，判定此手指不可用
                return False
            if n["beat"] > note["beat"]: #超出范围，结束
                break
            if n["note"] == "Slide" and "start" in n and n["start"]: #检测到绿条开始键，标记
                hold = True
                pos = n["pos"]
            elif n["note"] == "Slide" and "end" in n and n["end"]:#检测到绿条结束键，结束标记
                hold = False
                pos = ""
        if hold: #如果在击打长条的过程中
            if "pos" in note and pos == note["pos"]: #相同滑条，返回真
                return True
            else:# 不同滑条，返回假
                return False
        return True

    def append(self,note):
        '''将音符插入notelist中，输入要插入note，
        返回 0:插入失败，1 插入一般音符 ， 2 插入结束音符 ， 3 插入开始音符 ， 4 插入节点'''
        if self.available(note):
            self.notelist.append(note) # 插入音符
            if note["note"] == "Slide": # 如果是绿条
                if "end" in note and note["end"] == True:
                    if self.hold == False: # 如果遇到了结束音符，而且根本没在滑绿条
                        raise Exception("Unexpected End Note!",note["beat"])
                    self.hold = False
                    self.pos = ""
                    return 2
                elif "start" in note and note["start"] == True:
                    if self.hold == True: # 如果遇到了开始音符，而且已经在滑绿条了
                        raise Exception("Unexpected Start Note!",note["beat"])
                    self.hold = True
                    self.pos = note["pos"]
                    return 3
                else:
                    if self.hold == False: #如果遇到了节点音符，而且没有在滑绿条
                        raise Exception("Unexpected Tick Note!",note["beat"])
                    return 4
            return 1
        else:
            return 0

    def simplified(self):
        for i in range(2,self.notelist):
            erasetick(self.notelist[i-2],self.notelist[i-1],self.notelist[i]) #简化音符，用于消除扭曲的绿条，暂时弃用
    
    def lastnotetime(self,time):
        '''返回time时间前最近的那个note的时间'''
        if len(self.notelist) != 0:
            rtr = self.notelist[-1]["time"]
            for i in range(len(self.notelist)-1,0,-1):
                rtr = self.notelist[i]["time"]
                if rtr < time:
                    return rtr
        return 0.0

def isHitnote(*args):
    '''判断一个note是不是需要击打的note;
    输入：一堆note，输出：这些note是不是都可击打'''
    for note in args:
        if note["note"] == "Single" or note["note"] == "Slide" and "start" in note and note["start"] == True:
            pass
        else:
            return False
    return True

def isDouble(note1,note2):
    '''返回并标记双压.
    输入值：两个需要判断的note，输出值：是否为双压，且对双压执行了标级'''
    if note1["beat"] == note2["beat"] and isHitnote(note1,note2):
        if abs(note1["finger"]) == 2 or abs(note2["finger"]) == 2:
            raise Exception("Triple or more notes!",note1["beat"])
        #判定为双压
        if note1["lane"] < note2["lane"]:
            #note1在note2左边
            note1["finger"] = -2 # note1 必须用左手击打
            note2["finger"] = 2 # note2 必须用右手击打
        elif note1["lane"] > note2["lane"]:
            note1["finger"] = 2 # note1 必须用右手击打
            note2["finger"] = -2 # note2 必须用左手击打
        else:
            raise Exception("Notes in same lane!",note1["beat"]) #note重叠，抛出错误
        return True
    else:
        # 不是双压
        return False

def isTrill(note1,note2,note3):
    '''返回并标记双压.
    输入值：两个需要判断的note，输出值：是否为双压，且对双压执行了标级'''
    mintime = 0.15
    if (note1["lane"] - note2["lane"]) * (note2["lane"] - note3["lane"]) < 0 and isHitnote(note1,note2,note3) and note1["time"] != note2["time"] and note2["time"] != note3["time"] and abs(note1["time"] - note2["time"]) < mintime and abs(note2["time"] - note3["time"])<mintime and abs(abs(note1["beat"] - note2["beat"]) - abs(note2["beat"] - note3["beat"]))<0.01 and note1["finger"] == 0 and note2["finger"] == 0 and note3["finger"] == 0:
        # 先左后右，先右后左，判定为交互的一部分
        if note2["lane"] > note1["lane"] and note1["lane"] <= 5 and note3["lane"] <= 5 and note2["lane"] >= 3:
            # 左右左结构
            note1["finger"] = -1
            note2["finger"] = 1
            note3["finger"] = -1        
            return True     
        elif note2["lane"] < note1["lane"] and note1["lane"] >= 3 and note3["lane"] >= 3 and note2["lane"] <= 5:
            # 右左右结构
            note1["finger"] = 1
            note2["finger"] = -1
            note3["finger"] = 1
            return True 
    return False

def appendslide(note,chart,finger):
    '''将绿条中所有的元素append到finger里
    输入：起始绿条note，谱面，起始绿条的index，append入的手指
    返回 绿条尾的note'''
    index = chart.index(note)
    for i in range(index,len(chart)):
        note = chart[i]
        if finger.append(note) == 2:
            return note # 返回绿条尾的note
    #return finger.notelist[-1]

def slidepos(startnote,chart):
    '''计算这个绿条要用哪只手接，并标记
    输入一个绿条开始键，无输出'''
    tmp = Finger(4) # 暂借finger类存储绿条
    endnote = appendslide(startnote,chart,tmp) # 计算出绿条的结尾
    index = chart.index(startnote) # 计算出绿条的开头
    for i in range(index,len(chart)):
        note = chart[i] 
        if note["beat"] > endnote["beat"]:
            # 已经运行出绿条范围
            break
        if note not in tmp.notelist:
            # 发现绿条范围内非该绿条元素
            currentpos = tmp.position(note["beat"],note["time"]) #计算当前绿条位置
            if note["lane"] < currentpos:
                startnote["finger"] = 3 # 有元素出现在左边，用右手打
                break
            elif note["lane"] > currentpos:
                startnote["finger"] = -3 # 有元素出现在右边，用左手打
                break
            else:
                raise Exception("Slide and Note in same lane!",note["beat"])

def appendnote(note,chart,finger):
    '''插入音符
    输入一个键，谱面和要插的手指，无输出'''
    if note["note"] == "Slide" and "start" in note and note["start"]: #如果是绿条，那把整个绿条都append进去
        appendslide(note,chart,finger)
    else: # 如果是普通键，把普通键append进去
        finger.append(note)

def checkallslide(finger1,finger2,bpmchart): # 测试谱面用的allslide生成器
    checkchart = bpmchart.copy()
    checkchart.append({"type":"Note","note":"Slide","beat":finger1.notelist[0]["beat"],"pos":"A","start":True,"lane":finger1.notelist[0]["lane"]})
    checkchart.append({"type":"Note","note":"Slide","beat":finger2.notelist[0]["beat"],"pos":"B","start":True,"lane":finger2.notelist[0]["lane"]})
    for i in range(1,len(finger1.notelist)-1):
        note = finger1.notelist[i]
        checkchart.append({"type":"Note","note":"Slide","beat":note["beat"],"pos":"A","lane":note["lane"]})
    for i in range(1,len(finger2.notelist)-1):
        note = finger2.notelist[i]
        checkchart.append({"type":"Note","note":"Slide","beat":note["beat"],"pos":"B","lane":note["lane"]})
    checkchart.append({"type":"Note","note":"Slide","beat":finger1.notelist[-1]["beat"],"pos":"A","lane":finger1.notelist[-1]["lane"],"end":True})
    checkchart.append({"type":"Note","note":"Slide","beat":finger2.notelist[-1]["beat"],"pos":"B","lane":finger2.notelist[-1]["lane"],"end":True})
    return checkchart

def play(chartjson):
    chart,bpmchart = calctime(chartjson)
    lefthand = Finger(2)
    righthand = Finger(6)
    for i in range(1,len(chart)):
        #标记双压与计算绿条 
        note = chart[i-1]
        if note["note"] == "Slide" and "start" in note and note["start"] == True:
            slidepos(note,chart)
        isDouble(chart[i-1],chart[i])
    for i in range(2,len(chart)):
        #标记交互
        isTrill(chart[i-2],chart[i-1],chart[i])
    for i in range(0,len(chart)):
        #正式游玩
        note = chart[i]
        leftavailable = lefthand.available(note) # 看看左手能不能打这个音符
        rightavailable = righthand.available(note) # 看看右手能不能打这个音符
        leftpos = lefthand.position(note["beat"],note["time"]) # 计算左手此时的位置
        rightpos = righthand.position(note["beat"],note["time"]) #计算右手此时的位置
        if note in lefthand.notelist or note in righthand.notelist:
            continue
        if leftavailable and not rightavailable:
            #左手可用，右手不可用
            if note["lane"] >= rightpos:
                # 左手需要跨手
                raise Exception("Hand crossing!(Left -> Right)",note["beat"])
            appendnote(note,chart,lefthand)
        elif not leftavailable and rightavailable:
            #右手可用，左手不可用
            if note["lane"] <= leftpos:
                # 右手需要跨手
                raise Exception("Hand crossing!(Right -> Left)",note["beat"])
            appendnote(note,chart,righthand)
        elif not leftavailable and not rightavailable:
            #双手均不可用，报错
            raise Exception("No Hand available",note["beat"])
        else:
            #双手均可用
            if abs(note["finger"]) > 0:
                #如果之前有双压/绿条/交互标记
                if note["finger"] < 0:
                    appendnote(note,chart,lefthand)
                elif note["finger"] > 0:
                    appendnote(note,chart,righthand)
            elif abs(note["lane"] - leftpos) < abs(note["lane"] - rightpos) or abs(note["lane"] - leftpos) == abs(note["lane"] - rightpos) and ( note["lane"] < 4 or (note["lane"] == 4 and lefthand.lastnotetime(note["time"]) < righthand.lastnotetime(note["time"]) and note["time"] - lefthand.lastnotetime(note["time"]) < 0.15)):
                # 贪心思想插note
                appendnote(note,chart,lefthand)
            else:
                appendnote(note,chart,righthand)
    return lefthand,righthand,bpmchart,chart

def calcdelta(finger):
    '''计算每个音符之间的时间差'''
    notelist = finger.notelist
    notelist.sort(key=lambda note: note["time"])
    notelist[0]["backdelta"] = (0,-1) # 第一个、最后一个音符的两个时间差不计
    notelist[-1]["frontdelta"] = (0,-1)
    for i in range(0,len(notelist)-1):# 下面就是计算公式
        notelist[i+1]["backdelta"] = notelist[i]["frontdelta"] = (abs(notelist[i+1]["lane"]-notelist[i]["lane"]) / (notelist[i+1]["time"]-notelist[i]["time"]),notelist[i+1]["time"]-notelist[i]["time"])
    return len(notelist)
        
def calcdetails(lefthand,righthand,chart):
    def isflick(note):
        return "flick" in note and note["flick"] and note["note"] != "Slide" #检查是不是粉键
    def mean(notes,key,position="frontdelta"): #计算平均值
        total = 0.0
        count = 0
        for note in notes:
            total = total + note[position][key]
            count = count + 1
        if count == 0:
            return 0
        else:
            return total / count
    def notelistslice(totaltime,notelist): # 将谱面分片处理
        res = []
        subnotelist = []
        subtime = 0.0
        offsettime = 1.5
        for note in notelist:
            if note["time"] - subtime > offsettime:
                res.append(subnotelist)
                subnotelist = []
                subtime = subtime + offsettime
            subnotelist.append(note)
        return res,offsettime

    leftcount = calcdelta(lefthand) # 左手音符数
    rightcount = calcdelta(righthand) #右手音符数
    leftpercent = leftcount * 1.0/ (leftcount+rightcount) # 左手的压力
    notelist = [lefthand.notelist,righthand.notelist] 
    totaltime = chart[-1]["time"] - chart[0]["time"] # 总时间
    totalnps = len(chart) / totaltime # nps值
    totalhps = len(list(filter(isHitnote,chart))) / totaltime #hps值
    results = {"fingermaxhps":[],"fingermaxfps":[[],[]],"fingermaxspd":[]} # 初始化
    if totaltime < 15.0: #避免有人拿短谱炸定级器
        raise Exception("Chart is too short!")
    for ntlist in notelist:
        #先进行爆发单手检测，检测1.5s内的note数
        hitnotelist = list(filter(isHitnote,ntlist))
        hitnotelists,subtime = notelistslice(totaltime,hitnotelist)
        for tmplist in hitnotelists:
            results["fingermaxhps"].append(len(tmplist) / subtime)
            #print(len(tmplist) , subtime)
        #再进行爆发粉键检测，包括前向检测和后向检测，不包括尾粉,仅检测10个粉键
        ntlist.sort(key=lambda note: note["frontdelta"][1],reverse=True)
        flicknotelist = list(filter(isflick,ntlist))
        if len(flicknotelist) >= 20:
            count = len(flicknotelist)
            calcup = math.ceil(count*0.9)
            calcdown = calcup - 10
            results["fingermaxfps"][0].append(1.0 / mean(flicknotelist[calcdown:calcup],1,position="frontdelta"))
            results["fingermaxfps"][1].append(1.0 / mean(flicknotelist[calcdown:calcup],1,position="backdelta"))
        else:
            results["fingermaxfps"][0].append(0)
            results["fingermaxfps"][1].append(0)
        #再进行位移速度的检测
        ntlist.sort(key=lambda note: note["frontdelta"][0])
        count = len(ntlist)
        calcup = count
        calcdown = count - 10
        results["fingermaxspd"].append(mean(ntlist[calcdown:calcup],0))
    #返回这些检测值的最大值
    tmp = sorted(results["fingermaxhps"],reverse=True)
    fingermaxhps = (tmp[0] + tmp[1] + tmp[2] + tmp[3] + tmp[4]) / 5.0
    return (totalnps,totalhps,leftpercent,fingermaxhps,max(results["fingermaxfps"][0]),max(results["fingermaxfps"][1]),max(results["fingermaxspd"]))

def update(): #更新数据库函数
    def updatejson(i,diff,diffname):
        try:
            chartjson = rq.get("http://106.55.249.77/bdofftobdfan?id={}&diff={}".format(i,diffname)).json()
            if(chartjson["result"]):
                chartdata = chartjson["data"]
                level = rq.get("https://player.banground.fun/api/bestdori/official/info/"+str(i)+"/zh").json()["data"]["difficulty"][diff]["level"]
                finger1,finger2,__,chart = play(chartdata)
                cur.execute("INSERT INTO {} values (?,?,?,?,?,?,?,?,?)".format(diffname+"songlist"),(i,level) + calcdetails(finger1,finger2,chart))
                con.commit()
        except Exception as e:
            print(i,str(e))
    con = sqlite3.connect("standard.db")
    cur = con.cursor()
    cur.execute("CREATE TABLE IF NOT EXISTS songlist(id INTEGER PRIMARY KEY, level INTEGER, nps REAL, hps REAL, leftpercent REAL, maxfinnps REAL, maxfinfpsfront REAL, maxfinfpsback REAL, maxspd REAL)")
    con.commit()
    cur.execute("CREATE TABLE IF NOT EXISTS hardsonglist(id INTEGER PRIMARY KEY, level INTEGER, nps REAL, hps REAL, leftpercent REAL, maxfinnps REAL, maxfinfpsfront REAL, maxfinfpsback REAL, maxspd REAL)")
    con.commit()
    cur.execute("CREATE TABLE IF NOT EXISTS normalsonglist(id INTEGER PRIMARY KEY, level INTEGER, nps REAL, hps REAL, leftpercent REAL, maxfinnps REAL, maxfinfpsfront REAL, maxfinfpsback REAL, maxspd REAL)")
    con.commit()
    cur.execute("CREATE TABLE IF NOT EXISTS easysonglist(id INTEGER PRIMARY KEY, level INTEGER, nps REAL, hps REAL, leftpercent REAL, maxfinnps REAL, maxfinfpsfront REAL, maxfinfpsback REAL, maxspd REAL)")
    con.commit()
    for i in range(1,290):
        try:
            chartjson = rq.get("http://106.55.249.77/bdofftobdfan?id={}&diff=expert".format(i)).json()
            if(chartjson["result"]):
                chartdata = chartjson["data"]
                level = rq.get("https://player.banground.fun/api/bestdori/official/info/"+str(i)+"/zh").json()["data"]["difficulty"]["3"]["level"]
                finger1,finger2,__,chart = play(chartdata)
                cur.execute("INSERT INTO songlist values (?,?,?,?,?,?,?,?,?)",(i,level) + calcdetails(finger1,finger2,chart))
                con.commit()
            chartjson = rq.get("http://106.55.249.77/bdofftobdfan?id={}&diff=special".format(i)).json()
            if(chartjson["result"]):
                chartdata = chartjson["data"]
                level = rq.get("https://player.banground.fun/api/bestdori/official/info/"+str(i)+"/zh").json()["data"]["difficulty"]["4"]["level"]
                finger1,finger2,__,chart = play(chartdata)
                cur.execute("INSERT INTO songlist values (?,?,?,?,?,?,?,?,?)",(i+1000,level) + calcdetails(finger1,finger2,chart))
                con.commit()
            print(i)
                #print(calcdetails(finger1,finger2,chart))
            #error = 10/0
        except Exception as e:
            print(i,str(e))
            #cur.execute("INSERT INTO songlist values")
        diffnames = ["easy","normal","hard"]
        for j in range(0,3):
            updatejson(i,str(j),diffnames[j])
    con.close()

def newdiffcalc(levelnum): # 金克拉jin大佬的函数
    global mainbpm,totalnote
    x = [3.13188,4.58318,5.27976,6.63676,6.63676]
    return x[levelnum] * math.log10(totalnote*totalnote*mainbpm / (2545.37*math.sqrt(2)*math.pi))

def drawchartmap(): # ALL SLIDE谱面检错使用
    chartjson = rq.get("http://106.55.249.77/bdofftobdfan?id={}&diff=expert".format(88)).json()
    chartdata = chartjson["data"]
    finger1,finger2,bpmchart,__ = play(chartdata)
    print(json.dumps(checkallslide(finger1,finger2,bpmchart)))

def calcchartdetails(chart): # 出错时计算totalnps和totalhps用
    totaltime = chart[-1]["time"] - chart[0]["time"]
    totalnps = len(chart) / totaltime
    totalhps = len(list(filter(isHitnote,chart))) / totaltime
    return (totalnps,totalhps)

def calcdiff(id,levelnum=3,chart=""): #基于对比法的难度计算法
    levelname = list(["easy","normal","hard","expert","special"])[levelnum]
    con = sqlite3.connect("standard.db") #链接缓存数据库和对比数据库
    cur = con.cursor()
    rtr = {}
    if levelname == "expert" or levelname == "special": #指定对比数据库
        databasename = "songlist"
    else:
        databasename = levelname + "songlist"
    cur.execute("SELECT level, count(*) COUNT from {} where level in (select level from {} group by level) group by level order by level desc".format(databasename,databasename))
    levelcounter = cur.fetchall()
    levelcount = {} # 记录位置 ， 以便于计算难度
    s = 0
    maxlevel = -1
    minlevel = 99
    for level in levelcounter: # 生成对比表
        levelcount[level[0]] = s + level[1]
        s = s + level[1]
        maxlevel = max(maxlevel,level[0])
        minlevel = min(minlevel,level[0])

    cur.execute("CREATE TABLE IF NOT EXISTS levellist(id INTEGER PRIMARY KEY, totalnps REAL, totalhps REAL, leftpercent REAL, fingermaxhps REAL, fingermaxfpsfront REAL, fingermaxfpsback REAL, fingermaxspd REAL, fingermaxhpsdiff REAL, fingermaxfpsfrontdiff REAL, fingermaxfpsbackdiff REAL, fingermaxspddiff REAL, totalnpsdiff REAL, totalhpsdiff REAL, error TEXT)")
    con.commit()
    def calcposition(rank): # 计算难度里函数
        if rank <= levelcount[maxlevel]:
            return maxlevel
        if rank >= levelcount[minlevel]:
            return minlevel
        for i in range(minlevel,maxlevel):
            if levelcount[i] >= rank and levelcount[i+1] < rank:
                return i + (rank - levelcount[i]) * 1.0/ (levelcount[i+1] - levelcount[i]) 
    def calcitemdiff(name,value): # 计算难度外函数
        cur.execute("SELECT count(*) from {} where {} >= {} ".format(databasename,name,value))
        level = round(calcposition(cur.fetchone()[0] + 1 ),1)
        if level == maxlevel:
            cur.execute("SELECT {} from {} order by {} desc".format(name,databasename,name))
            tmplist = cur.fetchall()
            level28 = tmplist[levelcount[maxlevel]][0]
            level29 = tmplist[0][0]
            newlevel = round(maxlevel + (math.log(value) - math.log(level29)) / (math.log(level29) - math.log(level28))*0.2,1)
            level = max(level,newlevel)
        return level
    starttime = process_time()
    if(id==-1): #POST指令专用
        chartdata = chart
    elif(id<=500): # 这些都是拉取谱面的代码
        chartjson = rq.get("http://106.55.249.77/bdofftobdfan?id={}&diff={}".format(id,levelname)).json()
        if(chartjson["result"]):
            chartdata = chartjson["data"]
        else:
            raise Exception("Chart not found!")
    else:
        cur.execute("SELECT * from levellist where id=?",(id,)) #看看谱面是否缓存
        res = cur.fetchone()
        totaltime = process_time() - starttime
        chartjson = rq.get("https://player.banground.fun/api/bestdori/community/{}".format(id)).json()
        if(chartjson["result"]):
            chartdata = chartjson["data"]["notes"]
        else:
            raise Exception("Chart not found!")
        if(res != None):
            calctime(chartdata)
            return {"totalnps":res[1],
                    "totalhps":res[2],
                    "leftpercent":res[3],
                    "fingermaxhps":res[4],
                    "fingermaxfpsfront":res[5],
                    "fingermaxfpsback":res[6],
                    "fingermaxspd":res[7],
                    "fingermaxhpsdiff":res[8],
                    "fingermaxfpsfrontdiff":res[9],
                    "fingermaxfpsbackdiff":res[10],
                    "fingermaxspddiff":res[11],
                    "totalnpsdiff":res[12],
                    "totalhpsdiff":res[13],
                    "error":res[14],
                    "totaltime":totaltime,
                    "newdiffcalc":newdiffcalc(levelnum)}
    try:  
        finger1,finger2,__,chart = play(chartdata) # 正常运算
        rtr["totalnps"],rtr["totalhps"],rtr["leftpercent"],rtr["fingermaxhps"],rtr["fingermaxfpsfront"],rtr["fingermaxfpsback"],rtr["fingermaxspd"] = calcdetails(finger1,finger2,chart)
    except Exception as e: # 出现分析错误
        rtr["error"] = str(e)
        chart = calctime(chartdata)[0]
        rtr["totalnps"],rtr["totalhps"] = calcchartdetails(chart)
        rtr["leftpercent"] = rtr["fingermaxhps"] = rtr["fingermaxfpsfront"] = rtr["fingermaxfpsback"] = rtr["fingermaxspd"] = rtr["fingermaxhpsdiff"] = rtr["fingermaxfpsfrontdiff"] = rtr["fingermaxfpsbackdiff"] = rtr["fingermaxspddiff"] = None
    else: # 正常计算
        rtr["fingermaxhpsdiff"] = calcitemdiff("maxfinnps",rtr["fingermaxhps"])
        if(rtr["fingermaxfpsfront"] != 0):
            rtr["fingermaxfpsfrontdiff"] = calcitemdiff("maxfinfpsfront",rtr["fingermaxfpsfront"])
            rtr["fingermaxfpsbackdiff"] = calcitemdiff("maxfinfpsback",rtr["fingermaxfpsback"])
        else:
            rtr["fingermaxfpsfront"] = None
            rtr["fingermaxfpsback"] = None
            rtr["fingermaxfpsfrontdiff"] = None
            rtr["fingermaxfpsbackdiff"] = None
        rtr["fingermaxspddiff"] = calcitemdiff("maxspd",rtr["fingermaxspd"])  
        rtr["error"] = None
    finally: # 补充npsdiff和hpsdiff
        rtr["totalnpsdiff"] = calcitemdiff("nps",rtr["totalnps"])
        rtr["totalhpsdiff"] = calcitemdiff("hps",rtr["totalhps"])
    rtr["totaltime"] = process_time()-starttime
    if(id>=500): # 大于500的谱面缓存
        cur.execute("INSERT INTO levellist values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",(id,rtr["totalnps"],rtr["totalhps"],rtr["leftpercent"],rtr["fingermaxhps"],rtr["fingermaxfpsfront"],rtr["fingermaxfpsback"],rtr["fingermaxspd"],rtr["fingermaxhpsdiff"],rtr["fingermaxfpsfrontdiff"],rtr["fingermaxfpsbackdiff"],rtr["fingermaxspddiff"],rtr["totalnpsdiff"],rtr["totalhpsdiff"],rtr["error"]))
        con.commit()
    con.close()
    rtr["newdiffcalc"] = newdiffcalc(levelnum) #加入大佬的运算公式
    return rtr

#drawchartmap()
#update()
#print(json.dumps(calcdiff(22847)))