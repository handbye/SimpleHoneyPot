### 工具说明：

低版本goby扫描特定IP（vps），会出现如下图右边红框中的payload，点击图中的位置即可CS上线，上线成功后会在本地生成以当时日期命名的txt文件，里面包含着用户访问的IP+端口+时间

<img src="https://cdn.jsdelivr.net/gh/dummersoul/Picture@main/img/image-20210924155107660.png" alt="image-20210924155107660" style="zoom: 67%;" />



<img src="https://cdn.jsdelivr.net/gh/dummersoul/Picture@main/img/image-20210924160939591.png" alt="image-20210924160939591" style="zoom:80%;" />



<img src="https://cdn.jsdelivr.net/gh/dummersoul/Picture@main/img/image-20210924160707339.png" alt="image-20210924160707339" style="zoom: 67%;" />



### 准备工作：

- python3运行脚本

- test.js应放置在vps上，可远程访问，名字可任意修改

- CS 先准备 Scripted Web Delivery 64位 powershell shellcode

- goby-win-x64-1.8.293

- test.js代码中需修改IP为vps的IP

  

### 示例：

python3 fz.py 1.1.1.1 80 test.js 888













