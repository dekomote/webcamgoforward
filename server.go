package main 


import (
    "./logger"
    "net"
    "bufio"
)


type Message struct {
    command string
    payload string
}


type Client struct {

    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer
    inbound chan Message
    outbound chan Message

}


func (client *Client) Read() {

    for {
        line, _ := client.reader.ReadString("---jsonrpcprotocolboundary---")
        //client.inbound <- line
    }
}


func (client *Client) Write() {
    //for data := range client.outgoing {
    //    client.writer.WriteString(data)
    //    client.writer.Flush()
    //}
}


func (client *Client) Listen() {
    go client.Read()
    go client.Write()
}


func NewClient(conn net.Conn) *Client{
    writer := bufio.NewWriter(conn)
    reader := bufio.NewReader(conn)

    client := &Client {
        inbound: make(chan Message),
        outbound: make(chan Message),
        conn: conn,
        reader: reader,
        writer: writer,
    }

    logger.Info.Printf("Client %v is accepting messages\n", conn.RemoteAddr())
    client.Listen()

    return client
}


func connectionMade(conn net.Conn) {
    logger.Info.Printf("Client %v connected\n", conn.RemoteAddr())

    client := NewClient(conn)
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