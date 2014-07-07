package main 


import (
    "./logger"
    "./utils"
    "net"
    "bufio"
    "strings"
)


const MESSAGE_BOUNDARY  string = "---jsonrpcprotocolboundary---"

type Client struct {

    conn net.Conn
    Reader *bufio.Reader
    writer *bufio.Writer
    inbound chan utils.Message
    outbound chan utils.Message

}


func (client *Client) Read() {

    buff := ""
    for {
        tempBuff, err := client.Reader.ReadString('-')
        if err != nil {
            logger.Error.Println(err)
        }
        
        
        buff += string(tempBuff)
        if strings.Contains(buff, MESSAGE_BOUNDARY) {
            s := strings.Split(buff, MESSAGE_BOUNDARY)[0]
            o := buff[len(s):]
            client.inbound <- utils.Unpack([]byte(s))
            buff = o
        }
    }
}


func (client *Client) Write() {
    for data := range client.outbound {

        logger.Info.Printf("Write %v", data.Pack())
        client.writer.WriteString(data.Pack())
        client.writer.WriteString(MESSAGE_BOUNDARY)
        client.writer.Flush()
    }
}


func (client *Client) Listen() {
    go client.Read()
    go client.Write()
}


func NewClient(conn net.Conn) *Client{
    writer := bufio.NewWriter(conn)
    Reader := bufio.NewReader(conn)

    client := &Client {
        inbound: make(chan utils.Message),
        outbound: make(chan utils.Message),
        conn: conn,
        Reader: Reader,
        writer: writer,
    }

    logger.Info.Printf("Client %v is accepting messages\n", conn.RemoteAddr())
    client.Listen()

    m := utils.Message{"authenticate","sdjfsodifjsoij"}

    client.outbound <- m

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