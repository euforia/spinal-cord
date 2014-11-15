package libs

import(
    "fmt"
    "github.com/streadway/amqp"
    "github.com/euforia/spinal-cord/logging"
    zmq "github.com/pebbe/zmq3"
)


type AMQPCallback func(<-chan amqp.Delivery, chan error, *logging.Logger, *zmq.Socket)

type AMQPInput struct {
    conn      *amqp.Connection
    channel   *amqp.Channel
    queueName string
    tag       string
    done      chan error
    logger *logging.Logger
}

func (c *AMQPInput) connect(amqpURI string) error {
    var err error
    c.logger.Debug.Printf("Dialing %q...\n", amqpURI)
    c.conn, err = amqp.Dial(amqpURI)
    if err != nil {
        return err
    }
    go func() {
        fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
    }()

    c.logger.Debug.Printf("Connection established. Getting channel...\n")
    c.channel, err = c.conn.Channel()
    return err
}

func NewAMQPInput(amqpURI string, bindExch []string, exchangeType, queueName, key, ctag string, logger *logging.Logger) (*AMQPInput, error) {
    c := &AMQPInput{
        conn:    nil,
        channel: nil,
        queueName: queueName,
        tag:     ctag,
        done:    make(chan error),
        logger: logger,
    }

    var err error

    err = c.connect(amqpURI)
    if err != nil {
        return c, err
    }

    queue, err := c.channel.QueueDeclare(
        queueName, // name of the queue
        true,      // durable
        false,     // delete when usused
        false,     // exclusive
        false,     // noWait
        nil,       // arguments
    )
    if err != nil {
        return nil, fmt.Errorf("Queue Declare: %s", err)
    }

    c.logger.Warning.Printf("Queue: %q; Messages: %d; Consumers: %d; Routing key: %s\n",
        queue.Name, queue.Messages, queue.Consumers, key)

    // bind to multiples
    c.BindToExchanges(bindExch, key, queueName)
    return c, nil
}

func (c *AMQPInput) Start(callback AMQPCallback, sock *zmq.Socket) error {
    deliveries, err := c.channel.Consume(
        c.queueName,
        c.tag,      // consumerTag,
        false,      // noAck
        false,      // exclusive
        false,      // noLocal
        false,      // noWait
        nil,        // arguments
    )
    if err != nil {
        return err
    }

    go callback(deliveries, c.done, c.logger, sock)
    return nil
}

func (c *AMQPInput) BindToExchanges(exchanges []string, key string, queueName string) {
    success := 0
    for _, exch := range exchanges {
        c.logger.Debug.Printf("Binding to exchange: %s...\n", exch)
        err := c.channel.QueueBind(
            queueName, // name of the queue
            key,        // routingKey
            exch,   // exchange sourceExchange
            false,      // noWait
            nil,        // arguments
        )
        if err != nil {
            c.logger.Error.Printf("Could not bind to queue: %s\n", err)
            continue
        }
        c.logger.Warning.Printf("Queue: %s <= bound to Exchange: %s\n", queueName, exch)
        success++
    }
    if success == 0 {
        c.logger.Error.Fatal("Could not bind to any queues!")
    }
}

func (c *AMQPInput) Shutdown() error {
    // will close() the deliveries channel
    if err := c.channel.Cancel(c.tag, true); err != nil {
        return fmt.Errorf("AMQPInput cancel failed: %s", err)
    }

    if err := c.conn.Close(); err != nil {
        return fmt.Errorf("AMQPInput connection close error: %s", err)
    }

    defer c.logger.Warning.Printf("AMQP shutdown complete!\n")
    // wait for handle() to exit
    return <-c.done
}
