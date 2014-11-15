package main

import(
    "flag"
    "fmt"
    "os"
    "net/http"
    "strings"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/aggregator"
    "github.com/euforia/spinal-cord/aggregator/inputs"
    "github.com/euforia/spinal-cord/reactor"
    "github.com/euforia/spinal-cord/reactor/handler"
    "github.com/euforia/spinal-cord/reactor/task"
    "github.com/euforia/spinal-cord/web"
)

const SPINAL_CORD_VERSION string = "0.0.2"

var (
    I_SPINAL_CORD_VERSION = flag.Bool("version", false, "Show version")
    // worker connection to task server //
    WORKER           = flag.Bool("worker", false, "Start worker")
    TASK_CONNECT_URI = flag.String("task-server-uri", "tcp://127.0.0.1:44444", "Worker connection to task server")

    // subreactor connection to spinal cord //
    //PSUB_CONNECT_URI = flag.String("psub-server-uri", "tcp://127.0.0.1:55000", "Spinal cord server")
    PSUB_CONNECT_URI string

    SP_LOGLEVEL      = flag.String("log-level", "trace", "Log level")
    HANDLERS_DIR     = flag.String("handlers-dir", "", "Directory to store handlers. (required)")

    HTTP_LISTEN_URI = flag.String("http-listen-addr", ":8080", "HTTP server")
    WEBROOT         = flag.String("webroot", "", "HTTP server web root")

    FEED_LISTEN_URI = flag.String("feed-listen-addr", "tcp://*:45454", "Input feed server")
    TASK_LISTEN_URI = flag.String("task-listen-addr", "tcp://*:44444", "Task server")
    PSUB_LISTEN_URI = flag.String("psub-listen-addr", "tcp://*:55000", "Publishing server")
    REQP_LISTEN_URI = flag.String("reqp-listen-addr", "tcp://*:55055", "Request/Response server")
)

func InitFlags(logger *logging.Logger) {
    flag.Parse()

    if *I_SPINAL_CORD_VERSION {
        fmt.Println(SPINAL_CORD_VERSION)
        os.Exit(0)
    }

    err := logger.SetLogLevel(*SP_LOGLEVEL)
    if err != nil {
        logger.Error.Fatal(err)
    }

    if *HANDLERS_DIR == "" {
        flag.PrintDefaults()
        logger.Error.Fatal("Handler directory required! (-handlers-dir)")
    }
    _, err = os.Stat(*HANDLERS_DIR)
    if err != nil {
        logger.Error.Fatalf("Could not open handlers directory: '%s'; Reason: %s\n", *HANDLERS_DIR, err)
    }

    pHostP := strings.Split(*PSUB_LISTEN_URI, ":")
    if pHostP[1] == "//*" {
        PSUB_CONNECT_URI = fmt.Sprintf("%s://127.0.0.1:%s", pHostP[0], pHostP[2])
    } else {
        PSUB_CONNECT_URI = *PSUB_LISTEN_URI
    }
    logger.Debug.Printf("Pub/Sub connect URI set: %s\n", PSUB_CONNECT_URI)

    if *WEBROOT != "" {
        _, err := os.Stat(*WEBROOT)
        if err != nil {
            logger.Error.Fatalf("Could not open webroot: '%s'; Reason: %s\n", *WEBROOT, err)
        }
    }
}

func StartWebService(logger *logging.Logger) {
    mgr := handler.NewHandlersManager(*HANDLERS_DIR, logger)

    h := web.NewRESTRouter("/api/ns", "*", logger) // prefix, default acl, logger
    h.Register("/",                            web.NewNamespaceHandle(mgr))
    h.Register("/namespace",                   web.NewEventTypeHandle(mgr))
    h.Register("/namespace/eventType",         web.NewEventTypeHandlersHandle(mgr))
    h.Register("/namespace/eventType/handler", web.NewEventHandlerHandle(mgr))

    http.Handle("/api/ns/", h)

    http.Handle("/", http.FileServer(http.Dir(*WEBROOT)))

    logger.Warning.Printf("Starting web service on: %s\n", *HTTP_LISTEN_URI)
    logger.Error.Fatal(http.ListenAndServe(*HTTP_LISTEN_URI, nil))
}

func StartSpinalCord(logger *logging.Logger, pubChan chan string) {
    /* load input feed service */
    go func(ch chan string) {
        reqRep := inputs.NewInputService("PULL", *FEED_LISTEN_URI, logger)
        reqRep.Start(ch)
    }(pubChan)

    /* load default request/response service */
    go func(ch chan string) {
        reqRep := inputs.NewInputService("REP", *REQP_LISTEN_URI, logger)
        reqRep.Start(ch)
    }(pubChan)

    /* start pub/sub server */
    go func(ch chan string) {
        pubSubServer := aggregator.NewPubSubServer(*PSUB_LISTEN_URI, logger)
        pubSubServer.Start(ch)
    }(pubChan)
}

func StartSubreactor(logger *logging.Logger) {
    sreactor := reactor.NewSubReactor(PSUB_CONNECT_URI, *TASK_LISTEN_URI, *HANDLERS_DIR, logger)
    sreactor.Start(true) // true = create samples
}

func main() {

    var logger = logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)

    InitFlags(logger)

    if *WORKER {

        worker := task.NewTaskWorker(*TASK_CONNECT_URI, *HANDLERS_DIR, logger)
        worker.Start()

    } else {

        pubSubChan := make(chan string)
        StartSpinalCord(logger, pubSubChan) // async
        if *WEBROOT != "" {
            go StartWebService(logger)
        } else {
            logger.Warning.Println("Web Service not starting. Webroot not provided!")
        }
        StartSubreactor(logger)
    }
}

