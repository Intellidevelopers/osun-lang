// server.os

let server = http.createServer(8080)

let helloHandler = func() {
    print("Hello from Osun server!")
}

let rootHandler = func() {
    print("Welcome to your Osun server!")
}

server.Handle("GET", "/", rootHandler)
server.Handle("GET", "/hello", helloHandler)

server.Listen()
