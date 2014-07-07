package main 


import (
    "./logger"
    "./utils"
    "fmt"
    "bufio"
    "mime/multipart"
    "net"
    "net/http"
    "net/textproto"
    "strings"
    "strconv"
    "encoding/base64"
    "code.google.com/p/go-uuid/uuid"
)


const MESSAGE_BOUNDARY  string = "---jsonrpcprotocolboundary---"
const MJPEG_BOUNDARY string = "mjpegboundary"

type Client struct {
    ID string
    secret string
    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer
    inbound chan utils.Message
    outbound chan utils.Message
    image_stream chan []byte
    image_write_locked bool
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
        }
    }
}


func (client *Client) Write() {
    for data := range client.outbound {
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
                logger.Info.Printf("Client %v authenticated\n", client.ID)
                m := utils.Message{"authenticated", client.ID}
                client.outbound <- m
                go client.AttachHandler()
            }
            case "heartbeat": {
                m := utils.Message{"heartbeat", ""}
                client.outbound <- m
            }
            case "image": {
                if !client.image_write_locked {
                    bin, _ := base64.StdEncoding.DecodeString(data.Payload)
                    client.image_stream <- bin
                }
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


func (client *Client) AttachHandler() {
    http.HandleFunc("/" + string(client.secret) + "/", func(w http.ResponseWriter, r *http.Request){
            logger.Info.Println("Starting the image stream")
            client.outbound <- utils.Message{"start_stream", ""}
            w.Header().Set("Content-type", "multipart/x-mixed-replace;boundary=" + MJPEG_BOUNDARY)
            multipartWriter := multipart.NewWriter(w)
            multipartWriter.SetBoundary(MJPEG_BOUNDARY)
            for image := range client.image_stream {
                client.image_write_locked = true
                iw, _ := multipartWriter.CreatePart(textproto.MIMEHeader{
                        "Content-type": []string{"image/jpeg"},
                        "Content-length": []string{strconv.Itoa(len(image))},
                    })
                iw.Write(image)
                //TODO Handle Error here - Stop the streaming
                client.image_write_locked = false
            }
            client.outbound <- utils.Message{"stop_stream", ""}
        })
}


func NewClient(conn net.Conn) *Client{
    writer := bufio.NewWriter(conn)
    reader := bufio.NewReader(conn)

    client := &Client {
        ID : uuid.New(),
        secret : "",
        inbound: make(chan utils.Message),
        outbound: make(chan utils.Message),
        image_stream: make(chan []byte),
        conn: conn,
        reader: reader,
        writer: writer,
        image_write_locked: false,
    }

    logger.Info.Printf("Client %v is accepting messages\n", conn.RemoteAddr())
    
    client.Listen()
    return client
}


func ConnectionMade(conn net.Conn) {
    logger.Info.Printf("Client %v connected\n", conn.RemoteAddr())
    NewClient(conn)
}


func StartHttpServer() {
    logger.Info.Println("Serving HTTP at 8080")
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
            fmt.Fprintf(w, "Check your client for the webcam url.")
        })
    http.ListenAndServe(":8080", nil)
}


func main() {
    logger.Init()

    go StartHttpServer()

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

        go ConnectionMade(conn)
    }

}