import socket

tcp_socket = socket.create_connection(("localhost", 8001))


while True:
    data = input("CMD > ")
    if data == "exit":
        tcp_socket.close()
        break

    data = data.strip().encode("utf-8")
    tcp_socket.sendall(data)
    response = tcp_socket.recv(4096)
    print("Received from server:", response.decode())
