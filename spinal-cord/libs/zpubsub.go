package libs

import(
    "github.com/euforia/spinal-cord/logging"
    zmq "github.com/pebbe/zmq3"
)

type PubSubServer struct {
    Sock *zmq.Socket
    Logger *logging.Logger
}

func NewPubSubServer(listenAddr string, logger *logging.Logger) *PubSubServer {

    pubServer, _ := zmq.NewSocket(zmq.PUB)
    err := pubServer.Bind(listenAddr)
    if err != nil {
        logger.Error.Fatalf("%s %v\n", listenAddr, err)
    }
    logger.Warning.Printf("Publishing Service started: %s\n", listenAddr)
    return &PubSubServer{pubServer, logger}
}

func (p *PubSubServer) Start(ch chan string) {
    for {
        p.Logger.Trace.Println("Waiting for events...")
        msg := <- ch
        p.Logger.Debug.Printf("Publishing: %s\n", msg)
        p.Sock.Send(msg, 0)
    }
}