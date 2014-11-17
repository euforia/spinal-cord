package main

import (
    "fmt"
    "os"
    "encoding/json"
    "flag"
    "github.com/streadway/amqp"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/revent"
    "github.com/euforia/spinal-cord/nurv/libs"
    "strings"
    "time"
    zmq "github.com/pebbe/zmq3"
)

const NURV_VERSION string = "0.0.2"
const LIFETIME time.Duration = 0

var (

    REQP_CONNECT_URI = flag.String("task-server-uri", "tcp://localhost:55055", "REQP URI to task server")
    EVENT_TYPE       = flag.String("event-type", "", "REQP event type to fire.")
    EVENT_DATA       = flag.String("data", "", "REQP data/payload for event.")

    AMQP_URI      = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
    EXCHANGE_TYPE = flag.String("exchange-type", "direct", "AMQP exchange type. Options - direct|fanout|topic|x-custom")
    QUEUE_NAME    = flag.String("queue-name", "", "AMQP ephemeral queue name")
    ROUTING_KEY   = flag.String("routing-key", "", "AMQP routing key")
    CONSUMER_TAG  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
    EXCH_STR      = flag.String("exchanges", "", "AMQP exchanges to bind to. (comma separated)")

    CONFIGFILE    = flag.String("c", "", "Configuration file")
    CONFIG *libs.Config
)

func getAMQPOptions(logger *logging.Logger, config *libs.Config) {

    amqpConfig := config.TypeConfig.(*libs.AMQPConfig)

    // Assign options to config
    if amqpConfig.RoutingKey == "" {
        amqpConfig.RoutingKey = *ROUTING_KEY
    }
    if amqpConfig.URI == "" {
        amqpConfig.URI = *AMQP_URI
    }
    if amqpConfig.ExchangeType == "" {
        amqpConfig.ExchangeType = *EXCHANGE_TYPE
    }
    if amqpConfig.QueueName == "" {
        amqpConfig.QueueName = *QUEUE_NAME
    }
    if amqpConfig.ConsumerTag == "" {
        amqpConfig.ConsumerTag = *CONSUMER_TAG
    }
    // Check options
    if amqpConfig.RoutingKey == "" {
        flag.PrintDefaults()
        logger.Error.Fatal("Must provide '-routing-key'!")
    }
    if amqpConfig.QueueName == "" {
        flag.PrintDefaults()
        logger.Error.Fatal("Must provide '-queue-name'!")
    }
    if len(amqpConfig.Exchanges) <= 0 {
        if *EXCH_STR == "" {
            flag.PrintDefaults()
            logger.Error.Fatal("Must provide '-exchanges'!")
        }
        bindExch := make([]string, 0)
        for _, b := range strings.Split(*EXCH_STR, ",") {
            if b != "" {
                bindExch = append(bindExch, b)
            }
        }
        if len(bindExch) <= 0 {
            flag.PrintDefaults()
            logger.Error.Fatal("Must provide '-exchanges'!")
        }
        amqpConfig.Exchanges = bindExch
    }
    logger.Debug.Printf("%s\n", config)
    //os.Exit(1)
}

func getReqRepOptions(logger *logging.Logger, config *libs.Config) {

    typeConfig := config.TypeConfig.(map[string]string)

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
    typeConfig["event_type"] = *EVENT_TYPE
    typeConfig["event_data"] = *EVENT_DATA
}

func startAmqpInput(logger *logging.Logger, config *libs.Config) {

    amqpConfig := config.TypeConfig.(*libs.AMQPConfig)
    c, err := libs.NewAMQPInput(amqpConfig, logger)
    if err != nil  {
        logger.Error.Fatalf("%s\n", err)
    } else  {
        logger.Warning.Printf("Connected: %s\n", amqpConfig.URI)
    }

    zsock, _ := zmq.NewSocket(zmq.PUSH)
    err = zsock.Connect(config.SpinalCord)
    defer zsock.Close()
    if err != nil {
        logger.Error.Fatal(err)
    }

    if err = c.Start(handle, zsock, config.Namespace); err != nil {
        logger.Error.Fatalf("%s\n", err)
    }

    if LIFETIME > 0 {
        logger.Info.Printf("Running for %s...", LIFETIME)
        time.Sleep(LIFETIME)
    } else {
        select {}
    }

    logger.Info.Printf("Shutting down...\n")
    if err := c.Shutdown(); err != nil {
        logger.Error.Fatalf("Error during shutdown: %s\n", err)
    }
}

func fireEvent(logger *logging.Logger, config *libs.Config) {

    resp, err := libs.FireSynapse(logger, config)
    if err != nil {
        logger.Error.Fatalf("%s\n", err)
    }
    fmt.Println(resp)
}

func decodeMessage(d amqp.Delivery, defaultNamespace string) (revent.Event, error) {
    var event revent.Event
    err := json.Unmarshal(d.Body, &event)
    if err != nil {
        return event, err
    }
    if event.Namespace == "" {
        event.Namespace = defaultNamespace
    }
    return event, nil
}

func handle(deliveries <-chan amqp.Delivery, done chan error, logger *logging.Logger,
                                            zsock *zmq.Socket, defaultNamespace string) {

    for d := range deliveries {
        //log.Printf("got %dB delivery: [%v] %q", len(d.Body),d.DeliveryTag,d.Body)
        event, err := decodeMessage(d, defaultNamespace)
        if err != nil {
            logger.Error.Printf("%v\n", err)
            continue
        }
        logger.Info.Printf("amqp => namespace: %s; event: %s\n", event.Namespace, event.Type)

        msg, err := event.JsonString()
        if err != nil {
            logger.Error.Println(err)
            continue
        }

        zsock.Send(msg, 0)
        logger.Debug.Printf("Event sent: %s\n", msg)
        d.Ack(false)
    }
    logger.Warning.Printf("'deliveries' channel closed\n")
    done <- nil
}

func Init(logger *logging.Logger) *libs.Config {

    config := libs.NewConfig("", "", "", "", NURV_VERSION)

    version := flag.Bool("version", false, "Show version")

    flag.StringVar(&config.LogLevel,   "l", "trace", "Log level (shorthand)")
    flag.StringVar(&config.LogLevel,   "log-level", "trace", "Log level")
    flag.StringVar(&config.Namespace,  "n", "misc", "Namespace for this event. (shorthand)")
    flag.StringVar(&config.Namespace,  "namespace", "misc", "Namespace for this event. i.e. 'nurv'.")
    flag.StringVar(&config.SpinalCord, "spinal-cord", "tcp://localhost:45454", "URI to spinal cord server")
    flag.StringVar(&config.NurvType,   "type", "amqp", "Type of input. Options - amqp|reqp")
    flag.Parse()

    if *version {
        fmt.Println(NURV_VERSION)
        os.Exit(0)
    }

    logger.SetLogLevel(config.LogLevel)

    logger.Debug.Printf("Using namespace => %s\n", config.Namespace)
    if config.Namespace == "misc" {
        logger.Warning.Println("Using '-namespace' recommended!")
    }

    logger.Warning.Printf("Nurv Type: %s\n", config.NurvType)
    switch(config.NurvType) {
        case "amqp":
            config.TypeConfig = libs.NewAMPQConfig()
            if *CONFIGFILE != "" {
                if err := libs.LoadConfigFromFile(*CONFIGFILE, config); err != nil {
                    logger.Error.Fatal(err)
                }
            }
            getAMQPOptions(logger, config)
            break
        case "reqp":
            config.TypeConfig = make(map[string]string)
            // currently no config file for single event
            config.SpinalCord = *REQP_CONNECT_URI
            getReqRepOptions(logger, config)
            break
        default:
            flag.PrintDefaults()
            logger.Error.Fatal("Nurv type not supported: '%s'!", config.NurvType)
    }
    return config
}

func main() {

    var logger = logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
    config := Init(logger)

    switch(config.NurvType) {
        case "reqp":
            fireEvent(logger, config)
            break
        case "amqp":
            startAmqpInput(logger, config)
            break
        default:
            break
    }
}