import pandas as pd  # 导入另一个包“pandas” 命名为 pd，理解成pandas是在 numpy 基础上的升级包
import numpy as np  # 导入一个数据分析用的包“numpy” 命名为 np
import matplotlib.pyplot as plt  # 导入 matplotlib 命名为 plt，类似 matlab，集成了许多可视化命令

def normfun(x,mu,sigma):
    pdf = np.exp(-((x - mu)**2)/(2*sigma**2)) / (sigma * np.sqrt(2*np.pi))
    return pdf


time = [140,146,150,150,150,150,150,150,150,155]

mean= 149.22101123595513
std = 1.6278164717748154
x = np.arange(142, 157, 0.1)
# 设定 y 轴，载入刚才的正态分布函数
y = normfun(x, mean, std)
plt.plot(x, y)
# 画出直方图，最后的“normed”参数，是赋范的意思，数学概念
plt.hist(time, bins=10, rwidth=0.9, normed=True)
plt.title('Time distribution')
plt.xlabel('Time')
plt.ylabel('Probability')
# 输出
plt.show()
