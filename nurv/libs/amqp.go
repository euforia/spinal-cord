package libs

import(
    "fmt"
    "github.com/streadway/amqp"
    "github.com/euforia/spinal-cord/logging"
    zmq "github.com/pebbe/zmq3"
)


type AMQPCallback func(<-chan amqp.Delivery, chan error, *logging.Logger, *zmq.Socket, string)

type AMQPInput struct {
    AMQPConfig
    conn      *amqp.Connection
    channel   *amqp.Channel
    done      chan error
    logger    *logging.Logger
}

func (c *AMQPInput) connect() error {
    var err error
    c.logger.Debug.Printf("Dialing %q...\n", c.URI)
    c.conn, err = amqp.Dial(c.URI)
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
func NewAMQPInput(config *AMQPConfig, logger *logging.Logger) (*AMQPInput, error) {

    c := &AMQPInput{*config, nil, nil, make(chan error), logger,}

    err := c.connect()
    if err != nil {
        return c, err
    }
    queue, err := c.channel.QueueDeclare(
        c.QueueName, // name of the queue
        true,        // durable
        false,       // delete when usused
        false,       // exclusive
        false,       // noWait
        nil,         // arguments
    )

    if err != nil {
        return nil, fmt.Errorf("Queue Declare: %s", err)
    }

    c.logger.Warning.Printf("Queue: %q; Messages: %d; Consumers: %d; Routing key: %s\n",
        queue.Name, queue.Messages, queue.Consumers, c.RoutingKey)

    // bind to multiples
    c.BindToExchanges()
    return c, nil
}

func (c *AMQPInput) Start(callback AMQPCallback, sock *zmq.Socket, defaultNamespace string) error {
    deliveries, err := c.channel.Consume(
        c.QueueName,
        c.ConsumerTag,      // consumerTag,
        false,      // noAck
        false,      // exclusive
        false,      // noLocal
        false,      // noWait
        nil,        // arguments
    )
    if err != nil {
        return err
    }

    go callback(deliveries, c.done, c.logger, sock, defaultNamespace)
    return nil
}

func (c *AMQPInput) BindToExchanges() {
    success := 0
    for _, exch := range c.Exchanges {
        c.logger.Debug.Printf("Binding to exchange: %s...\n", exch)
        err := c.channel.QueueBind(
            c.QueueName,  // name of the queue
            c.RoutingKey, // routingKey
            exch,         // exchange sourceExchange
            false,        // noWait
            nil,          // arguments
        )
        if err != nil {
            c.logger.Error.Printf("Could not bind to queue: %s\n", err)
            continue
        }
        c.logger.Warning.Printf("Queue: %s <= bound to Exchange: %s\n", c.QueueName, exch)
        success++
    }
    if success == 0 {
        c.logger.Error.Fatal("Could not bind to any queues!")
    }
}

func (c *AMQPInput) Shutdown() error {
    if err := c.channel.Cancel(c.ConsumerTag, true); err != nil {
        return fmt.Errorf("AMQPInput cancel failed: %s", err)
    }
    // will close() the deliveries channel
    if err := c.conn.Close(); err != nil {
        return fmt.Errorf("AMQPInput connection close error: %s", err)
    }
    defer c.logger.Warning.Printf("AMQP shutdown complete!\n")
    // wait for handle() to exit
    return <-c.done
}
