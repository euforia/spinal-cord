package main

import (
    "fmt"
    "os"
    "encoding/json"
    "flag"
    "github.com/streadway/amqp"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/revent"
    "github.com/euforia/spinal-cord/aggregator/inputs"
    "strings"
    "time"
    zmq "github.com/pebbe/zmq3"
)

const VERSION string = "0.0.2"

var (
    SHOW_VERSION = flag.Bool("version", false, "Show version")
    NURV_TYPE    = flag.String("type", "amqp", "Type of input. Options - amqp|reqp")

    REQP_CONNECT_URI = flag.String("task-server-uri", "tcp://localhost:55055", "REQP URI to task server")
    EVENT_TYPE   = flag.String("event-type", "", "REQP event type to fire.")
    EVENT_DATA   = flag.String("data", "", "REQP data/payload for event.")

    LOGLEVEL     = flag.String("log-level", "trace", "Log level")
    NAMESPACE    = flag.String("namespace", "misc", "Namespace for this event input i.e. 'nurv'.")

    FEED_CONNECT_URI  = flag.String("feed-server", "tcp://localhost:45454", "URI to spinal cord server")

    AMQP_URI     = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
    exchangeType = flag.String("exchange-type", "direct", "AMQP exchange type. Options - direct|fanout|topic|x-custom")
    QUEUE_NAME   = flag.String("queue", "test-queue", "AMQP ephemeral queue name")
    ROUTING_KEY  = flag.String("routing-key", "", "AMQP routing key")
    consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
    LIFETIME     = flag.Duration("lifetime", 0, "AMQP consumer lifetime of process before shutdown (0s=infinite)")
    bindTo       = flag.String("bind-to", "", "AMQP exchanges to bind to. (comma separated)")
    bindToExch []string
)

func checkAMQP(logger *logging.Logger) {

    if *bindTo == "" {
        flag.PrintDefaults()
        logger.Error.Fatal("Must provide '-bind-to'!")
    }
    for _, b := range strings.Split(*bindTo, ",") {
        if b != "" {
            bindToExch = append(bindToExch, b)
        }
    }
    if len(bindToExch) <= 0 {
        flag.PrintDefaults()
        logger.Error.Fatal("Must provide '-bind-to'!")
    }
    logger.Warning.Printf("Namespace => %s\n", *NAMESPACE)
    if *NAMESPACE == "misc" {
        logger.Warning.Println("Using '-namespace' recommended!")
    }
}

func checkReqRep(logger *logging.Logger) {
    if *EVENT_TYPE == "" {
        fmt.Println("\n Must specify event type '-event-type'!\n")
        flag.PrintDefaults()
        os.Exit(1)
    }
    if *EVENT_DATA == "" {
        fmt.Println("\n Must specify data '-data'!\n")
        flag.PrintDefaults()
        os.Exit(2)
    }
}

func startAmqpInput(logger *logging.Logger) {
    c, err := inputs.NewAMQPInput(*AMQP_URI, bindToExch, *exchangeType,
                        *QUEUE_NAME, *ROUTING_KEY, *consumerTag, logger)
    if err != nil {
        logger.Error.Fatalf("%s\n", err)
    } else {
        logger.Warning.Printf("Connected: %s\n", *AMQP_URI)
    }

    if err = c.Start(handle); err != nil {
        logger.Error.Fatalf("%s\n", err)
    }

    if *LIFETIME > 0 {
        logger.Info.Printf("Running for %s...", *LIFETIME)
        time.Sleep(*LIFETIME)
    } else {
        //log.Printf("running forever")
        select {}
    }

    logger.Info.Printf("Shutting down...\n")

    if err := c.Shutdown(); err != nil {
        logger.Error.Fatalf("Error during shutdown: %s\n", err)
    }
}

func encodeEvent(evt revent.Event) (string, error) {

    bytes, err := json.Marshal(&evt)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

func fireEvent(logger *logging.Logger) {
    zsock, _ := zmq.NewSocket(zmq.REQ)
    zsock.Connect(*REQP_CONNECT_URI)
    defer zsock.Close()

    var payload map[string]interface{}
    err := json.Unmarshal([]byte(*EVENT_DATA), &payload)
    if err != nil {
        logger.Error.Fatalf("Could not serialize data: %s\n", EVENT_DATA)
    }
    evt := revent.Event{
                Namespace: *NAMESPACE,
                Type: *EVENT_TYPE,
                Payload: payload,
                Timestamp: float64(time.Now().UnixNano())/1000000000,
                }
    msg, err := encodeEvent(evt)
    if err != nil {
        logger.Error.Fatal(err)
    }
    zsock.Send(msg, 0)
    resp, err := zsock.Recv(0)
    if err != nil {
        logger.Error.Fatal(err)
    }
    fmt.Printf("%v\n", resp)
}

func Init(logger *logging.Logger) {
    flag.Parse()

    if *SHOW_VERSION {
        fmt.Println(VERSION)
        os.Exit(0)
    }

    logger.SetLogLevel(*LOGLEVEL)

    switch(*NURV_TYPE) {
        case "amqp":
            checkAMQP(logger)
            break
        case "reqp":
            checkReqRep(logger)
            break
        default:
            flag.PrintDefaults()
            logger.Error.Fatal("Nurv type not supported: '%s'!", *NURV_TYPE)
    }
}

func decodeMessage(d amqp.Delivery) (revent.Event, error) {
    var event revent.Event
    err := json.Unmarshal(d.Body, &event)
    if err != nil {
        return event, err
    }
    if event.Namespace == "" {
        event.Namespace = *NAMESPACE
    }
    return event, nil
}

func handle(deliveries <-chan amqp.Delivery, done chan error, logger *logging.Logger) {

    zsock, _ := zmq.NewSocket(zmq.PUSH)
    err := zsock.Connect(*FEED_CONNECT_URI)
    defer zsock.Close()
    if err != nil {
        logger.Error.Println(err)
    }
    spinalCord := inputs.BasicSock{zsock, logger}

    for d := range deliveries {
        //log.Printf("got %dB delivery: [%v] %q", len(d.Body),d.DeliveryTag,d.Body)
        event, err := decodeMessage(d)
        if err != nil {
            logger.Error.Printf("%v\n", err)
            continue
        }
        logger.Info.Printf("amqp => namespace: %s; event: %s\n", event.Namespace, event.Type)

        msg, err := encodeEvent(event)
        if err != nil {
            logger.Error.Println(err)
            continue
        }

        spinalCord.Sock.Send(msg, 0)
        logger.Debug.Printf("Event sent: %s\n", msg)
        d.Ack(false)
    }
    logger.Warning.Printf("'deliveries' channel closed\n")
    done <- nil
}

func main() {

    var logger = logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
    Init(logger)

    switch(*NURV_TYPE) {
        case "reqp":
            fireEvent(logger)
            break
        case "amqp":
            logger.Warning.Printf("Nurv Type: %s\n", *NURV_TYPE)
            startAmqpInput(logger)
            break
        default:
            break
    }
}