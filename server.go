package main 


import (
    "./logger"
    "./utils"
    "net"
    "bufio"
    "strings"
    "code.google.com/p/go-uuid/uuid"
)


const MESSAGE_BOUNDARY  string = "---jsonrpcprotocolboundary---"

type Client struct {
    ID string
    secret string
    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer
    inbound chan utils.Message
    outbound chan utils.Message


}


func (client *Client) Read() {

    buff := ""
    for {
        tempBuff, err := client.reader.ReadString('-')
        if err != nil {
            logger.Error.Println(err)
        }

        //TODO Handle EOF


        buff += string(tempBuff)
        if strings.Contains(buff, MESSAGE_BOUNDARY) {
            s := strings.Split(buff, MESSAGE_BOUNDARY)[0]
            o := buff[len(s)+len(MESSAGE_BOUNDARY):]
            client.inbound <- utils.Unpack([]byte(s))
            buff = o
            logger.Info.Println(buff)
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

func (client *Client) Message() {
    for data := range client.inbound {
        switch data.Command {
            case "authenticate": {
                client.secret = data.Payload
                m := utils.Message{"authenticated", client.ID}
                client.outbound <- m
            }
            case "heartbeat": {
                m := utils.Message{"heartbeat", ""}
                client.outbound <- m
            }
        }
    }
}


func (client *Client) Listen() {
    go client.Read()
    go client.Write()
    go client.Message()

    m := utils.Message{"authenticate", client.ID}
    client.outbound <- m
}


func NewClient(conn net.Conn) *Client{
    writer := bufio.NewWriter(conn)
    reader := bufio.NewReader(conn)

    client := &Client {
        ID : uuid.New(),
        secret : "",
        inbound: make(chan utils.Message),
        outbound: make(chan utils.Message),
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