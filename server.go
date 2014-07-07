package main 


import (
    "./logger"
    "net"
    "bufio"
    "mime/multipart"
)


type Message struct {
    command string
    payload string
}


type Client struct {

    conn net.Conn
    reader *multipart.Reader
    writer *bufio.Writer
    inbound chan string
    outbound chan string

}


func (client *Client) Read() {

    for {
        part, _ := client.reader.NextPart()
        var line []byte
        part.Read(line)
        client.inbound <- string(line)
    }
}


func (client *Client) Write() {
    for data := range client.outbound {
        logger.Info.Println(data)
        client.writer.WriteString(data)
        client.writer.Flush()
    }
}


func (client *Client) Listen() {
    go client.Read()
    go client.Write()
}


func NewClient(conn net.Conn) *Client{
    writer := bufio.NewWriter(conn)
    reader := multipart.NewReader(conn, "---jsonrpcprotocolboundary---")

    client := &Client {
        inbound: make(chan string),
        outbound: make(chan string),
        conn: conn,
        reader: reader,
        writer: writer,
    }

    logger.Info.Printf("Client %v is accepting messages\n", conn.RemoteAddr())
    client.Listen()

    client.outbound <- "{\"command\": \"authenticate\", \"payload\": \"dasd2342342asfdf234\"} ---jsonrpcprotocolboundary---"

    return client
}


func connectionMade(conn net.Conn) {
    logger.Info.Printf("Client %v connected\n", conn.RemoteAddr())
    NewClient(conn)
}


func main() {
    logger.Init()

    var ln, err = net.Listen("tcp", ":9000")
    logger.Info.Printf("Started listening on %v\n", ln.Addr())

    if err != nil {
        logger.Error.Println(err)
    }
    
    for {
        logger.Info.Println("Accepting connections...")
        var conn, err = ln.Accept()
        
        if err != nil {
            logger.Error.Println(err)
        }

        go connectionMade(conn)
    }

}