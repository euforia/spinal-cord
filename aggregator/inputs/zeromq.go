package inputs

import(
    "encoding/json"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/revent"
    "time"
    zmq "github.com/pebbe/zmq3"
)

type BasicSock struct {
    Sock *zmq.Socket
    Logger *logging.Logger
}

func NewBasicSock(sock *zmq.Socket, logger *logging.Logger) BasicSock {
    return BasicSock{sock, logger}
}

type InputService struct {
    sockType string
    BasicSock
}

func NewInputService(sockType string, listenAddr string, logger *logging.Logger) *InputService {

    var zsock *zmq.Socket
    switch(sockType) {
        case "PULL":
            zsock, _ = zmq.NewSocket(zmq.PULL)
            break
        case "REP":
            zsock, _ = zmq.NewSocket(zmq.REP)
            break
        default:
            logger.Error.Fatal("Invalid sock type!")
    }

    err := zsock.Bind(listenAddr)
    if err != nil {
        logger.Error.Fatalf("%s %v\n", listenAddr, err)
    }
    logger.Warning.Printf("%s Service started: %s\n", sockType, listenAddr)

    return &InputService{sockType, BasicSock{zsock, logger}}
}

func (b *InputService) CheckMessage(message string) (string, error) {

    var event revent.Event
    err := json.Unmarshal([]byte(message), &event)
    if err != nil {
        return "", err
    }
    if event.Timestamp == nil {
        b.Logger.Trace.Printf("%s - Adding timestamp to: %s\n", b.sockType, message)
        event.Timestamp = float64(time.Now().UnixNano())/1000000000
    }
    bytes, err := json.Marshal(&event)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

func (b *InputService) Start(ch chan string) {

    for {
        msg, err := b.Sock.Recv(0)
        if err != nil {
            b.Logger.Error.Println(err)
            continue
        }
        checkedMsg, err := b.CheckMessage(msg)
        //b.logger.Debug.Printf("Req/Rep - %s\n", msg)
        b.Logger.Debug.Printf("%s - %s\n", b.sockType, msg)
        ch <- checkedMsg
        if b.sockType == "REP" {
            b.Sock.Send(checkedMsg, 0)
        }
    }
}