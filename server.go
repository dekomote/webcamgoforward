package main

import (
	"bufio"
	"code.google.com/p/go-uuid/uuid"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

// Defining the message boundaries. The first one is used by the protocol
// and the second one is used by the HTTP server while streaming mjpeg
const MESSAGE_BOUNDARY string = "---jsonrpcprotocolboundary---"
const MJPEG_BOUNDARY string = "mjpegboundary"

// Client struct - keeps state for one connected client as well as channels
// associated with the Client.
type Client struct {
	ID                 string
	secret             string
	conn               net.Conn
	reader             *bufio.Reader
	writer             *bufio.Writer
	inbound            chan Message
	outbound           chan Message
	image_stream       chan []byte
	image_write_locked bool
}

// Read method to the client struct reads string by string from the bufio
// reader tied to the connection, splitting the messages by MESSAGE_BOUNDARY
// and sending each message to the inbound channel
func (client *Client) Read() {
	//
	buff := ""
	for {
		tempBuff, err := client.reader.ReadString('-')
		if err != nil {
			Error.Println(err)
			if err == io.EOF {
				return
			}
		}

		buff += string(tempBuff)
		if strings.Contains(buff, MESSAGE_BOUNDARY) {
			s := strings.Split(buff, MESSAGE_BOUNDARY)[0]
			o := buff[len(s)+len(MESSAGE_BOUNDARY):]
			client.inbound <- Unpack([]byte(s))
			buff = o
		}
	}
}

// Write method to the client struct writes data in the outbound channel
// to the bufio writer tied to the connection
func (client *Client) Write() {
	for data := range client.outbound {
		client.writer.WriteString(data.Pack())
		client.writer.WriteString(MESSAGE_BOUNDARY)
		client.writer.Flush()
	}
}

// Message method gets each message from the inbound channel and switches based
// on the command and payload
func (client *Client) Message() {
	for data := range client.inbound {
		switch data.Command {
		case "authenticate":
			{
				client.secret = data.Payload
				Info.Printf("Client %v authenticated\n", client.ID)
				m := Message{"authenticated", client.ID}
				client.outbound <- m
				go client.AttachHandler()
			}
		case "heartbeat":
			{
				m := Message{"heartbeat", ""}
				client.outbound <- m
			}
		case "image":
			{
				if !client.image_write_locked {
					bin, _ := base64.StdEncoding.DecodeString(data.Payload)
					client.image_stream <- bin
				}
			}
		}
	}
}

// Listen is a helper method that gophers Read, Write and Message.
func (client *Client) Listen() {
	go client.Read()
	go client.Write()
	go client.Message()

	m := Message{"authenticate", client.ID}
	client.outbound <- m
}

// AttachHandler method attaches a http serve handler function to an url
// specified by the Client's ID and secret. The attached function starts a
// mjpeg loop, writing multipart data fed from image_stream channel
func (client *Client) AttachHandler() {
	http.HandleFunc("/"+string(client.secret)+"/"+string(client.ID), func(w http.ResponseWriter, r *http.Request) {
		Info.Println("Starting the image stream")
		client.outbound <- Message{"start_stream", ""}
		w.Header().Set("Content-type", "multipart/x-mixed-replace;boundary="+MJPEG_BOUNDARY)
		multipartWriter := multipart.NewWriter(w)
		multipartWriter.SetBoundary(MJPEG_BOUNDARY)
		for image := range client.image_stream {
			// So we don't write to the channel while rendering this image.
			client.image_write_locked = true
			iw, parterr := multipartWriter.CreatePart(textproto.MIMEHeader{
				"Content-type":   []string{"image/jpeg"},
				"Content-length": []string{strconv.Itoa(len(image))},
			})
			if parterr != nil {
				Error.Println(parterr)
			} else {
				_, err := iw.Write(image)
				if err != nil {
					// The browser closed connection, or crashed...
					Error.Println(err)
					client.image_write_locked = false
					client.outbound <- Message{"stop_stream", ""}
					return
				}
			}
			client.image_write_locked = false
		}
		//client.outbound <- Message{"stop_stream", ""}
	})
}

// NewClient handles setting up a new client instance
func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		ID:                 uuid.New(),
		secret:             "",
		inbound:            make(chan Message),
		outbound:           make(chan Message),
		image_stream:       make(chan []byte),
		conn:               conn,
		reader:             reader,
		writer:             writer,
		image_write_locked: false,
	}

	Info.Printf("Client %v is accepting messages\n", conn.RemoteAddr())

	client.Listen()
	return client
}

// ConnectionMade is called after a client connects to the server. New instance
// of client is created per connection
func ConnectionMade(conn net.Conn) {
	Info.Printf("Client %v connected\n", conn.RemoteAddr())
	NewClient(conn)
}

// StartHttpServer starts the HTTP server with a dummy home page
func StartHttpServer() {
	Info.Println("Serving HTTP at 8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Check your client for the webcam url.")
	})
	http.ListenAndServe(":8080", nil)
}

func main() {
	Init()

	go StartHttpServer()

	var ln, err = net.Listen("tcp", ":9000")
	Info.Printf("Started listening on %v\n", ln.Addr())

	if err != nil {
		Error.Println(err)
		return
	}

	for {
		Info.Println("Accepting connections...")
		var conn, err = ln.Accept()

		if err != nil {
			Error.Println(err)
			continue
		}

		go ConnectionMade(conn)
	}

}
