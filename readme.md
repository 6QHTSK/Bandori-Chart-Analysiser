# 定级器使用说明

下载python文件chartanalyser.py 和 standard.db 数据库

执行命令即可计算

    ```python
    import chartanalysiser
    chartanalysiser.calcdiff(id,diffnum,chart) # id bestdori谱面id，diffnum 0-4分别为easy-special，使用chart项可计算自定义谱面，此时应把id置为-1
    ```
## 思想

SHITCODE警告

采用贪心算法和经验优化来分开双手，引入了@金坷垃jin大佬的“白蓝物量节拍公式”作为公式计算的补充。

有大量的坑没填

###致谢###

感谢灵喵提供机器人查难度支持

感谢banground游戏群群友的测试

感谢@金坷垃jin大佬的“白蓝物量节拍公式”

### 其它

定级器算出的难度**仅供参考**，如果有更好的算法，欢迎各位pull request！