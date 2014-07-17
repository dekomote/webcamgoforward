webcamgoforward
===============

Webcamgoforward - webcamforward forwarder server written in Go

The server proxies the webcam stream received from the clients (http://github.com/dekomote/webcamforward)
to the web. URL where you can access your stream is shown on the client.


Build
+++++

To build it, just get the single required package:

    go get code.google.com/p/go-uuid/uuid

Then get the server:

    go get github.com/dekomote/webcamgoforward

Build it and run it:

    cd $GOPATH
    cd github.com/dekomote/webcamgoforward
    go build
    ./webcamgoforward

At this point, you probably want to try it, so you should also get the client
from http://github.com/dekomote/webcamforward

Use QtCreator to build it. I don't have much time to create binary packages
right now.
