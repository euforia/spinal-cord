package nurvs

import (
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/decoders"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/synapse"
	"github.com/streadway/amqp"
)

type AMQPNurv struct {
	Namespace string

	cfg     config.AMQPNurvConfig
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error

	logger *logging.Logger

	// initialized when Init is called
	_synapse synapse.ISynapse
	// initialized when Init is called
	_decoder decoders.IDecoder
}

//func NewAMQPNurv(namespace string, cfg map[string]interface{}, logger *logging.Logger) (*AMQPNurv, error) {
func NewAMQPNurv(cfg *config.NurvConfig, logger *logging.Logger) (*AMQPNurv, error) {
	var (
		err error
		c   = AMQPNurv{
			Namespace: cfg.Namespace,
			conn:      nil,
			channel:   nil,
			done:      make(chan error),
			logger:    logger,
		}
	)
	//logger.Trace.Printf("%#v\n", c)
	//c.cfg, err = c.checkConfig(cfg)
	aConfig, err := c.checkConfig(cfg)
	if err != nil {
		return &c, err
	}
	c.cfg = *aConfig
	//logger.Trace.Printf("-- %#v\n", c)
	if err = c.connect(); err != nil {
		return &c, err
	}

	queue, err := c.channel.QueueDeclare(
		c.cfg.QueueName, // name of the queue
		true,            // durable
		false,           // delete when usused
		false,           // exclusive
		false,           // noWait
		nil,             // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("Queue declaration failed: %s", err)
	}

	c.logger.Warning.Printf("Queue: %q; Messages: %d; Consumers: %d; Routing key: %s\n",
		queue.Name, queue.Messages, queue.Consumers, c.cfg.RoutingKey)

	// bind to multiples
	c.bindToExchanges()
	return &c, nil
}

func (a *AMQPNurv) Init(decoder decoders.IDecoder, syn synapse.ISynapse) error {
	a._decoder = decoder
	a._synapse = syn
	return nil
}

//func (c *AMQPNurv) Start(callback AMQPCallback, sock *zmq.Socket, defaultNamespace string) error {
func (c *AMQPNurv) Start() error {
	//c._synapse = syn
	deliveries, err := c.channel.Consume(
		c.cfg.QueueName,
		c.cfg.ConsumerTag, // consumerTag,
		false,             // noAck
		false,             // exclusive
		false,             // noLocal
		false,             // noWait
		nil,               // arguments
	)
	if err != nil {
		c.logger.Error.Printf("%s\n", err)
		return err
	}

	go c.handleDeliveries(deliveries)
	return nil
}

func (a *AMQPNurv) handleDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {

		event, err := a._decoder.Decode(d.Body)
		if err != nil {
			a.logger.Error.Printf("%v\n", err)
			continue
		}
		if event.Namespace == "" {
			event.Namespace = a.Namespace
		}

		a.logger.Info.Printf("amqp => namespace: %s; event: %s\n",
			event.Namespace, event.Type)

		a._synapse.Fire(event)

		d.Ack(false)
	}
	a.logger.Warning.Printf("'deliveries' channel closed\n")
	a.done <- nil
}

func (c *AMQPNurv) Stop() error {
	if err := c.channel.Cancel(c.cfg.ConsumerTag, true); err != nil {
		return fmt.Errorf("AMQPNurv cancel failed: %s", err)
	}
	// will close() the deliveries channel
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQPNurv connection close error: %s", err)
	}
	defer c.logger.Warning.Printf("AMQP shutdown complete!\n")
	// wait for handle() to exit
	return <-c.done
}

func (c *AMQPNurv) bindToExchanges() {
	success := 0
	for _, exch := range c.cfg.Exchanges {
		c.logger.Debug.Printf("Binding to exchange: %s...\n", exch)
		err := c.channel.QueueBind(
			c.cfg.QueueName,  // name of the queue
			c.cfg.RoutingKey, // routingKey
			exch,             // exchange sourceExchange
			false,            // noWait
			nil,              // arguments
		)
		if err != nil {
			c.logger.Error.Printf("Could not bind to queue: %s\n", err)
			continue
		}
		c.logger.Warning.Printf("Queue: %s <= bound to Exchange: %s\n", c.cfg.QueueName, exch)
		success++
	}
	if success == 0 {
		c.logger.Error.Fatal("Could not bind to any queues!")
	}
}

func (a *AMQPNurv) checkConfig(cfg *config.NurvConfig) (*config.AMQPNurvConfig, error) {
	//var (
	//	aCfg *config.AMQPNurvConfig
	//)

	aCfg, ok := cfg.TypeConfig.(*config.AMQPNurvConfig)
	if !ok {
		return aCfg, fmt.Errorf("type assertion failed: %s", cfg.TypeConfig)
	}
	//a.logger.Trace.Printf("--- %#v\n", aCfg)
	/*
		aCfg := config.AMQPNurvConfig{
			Exchanges: make([]string, 0),
		}
	*/
	/*
		if v, ok := cfgmap["uri"].(string); !ok {
			return aCfg, fmt.Errorf("Invalid config parameter: %s", cfgmap["uri"])
		} else {
			aCfg.URI = v
		}

		if v, ok := cfgmap["queue_name"].(string); !ok {
			return aCfg, fmt.Errorf("Invalid config parameter: %s", cfgmap["queue_name"])
		} else {
			aCfg.QueueName = v
		}

		if v, ok := cfgmap["routing_key"].(string); !ok {
			return aCfg, fmt.Errorf("Invalid config parameter: %s", cfgmap["routing_key"])
		} else {
			aCfg.RoutingKey = v
		}
		exchs, ok := cfgmap["exchanges"].([]interface{})
		if !ok {
			return aCfg, fmt.Errorf("Invalid config parameter: %s Must be array", cfgmap["exchanges"])
		}
		for _, e := range exchs {
			val, ok := e.(string)
			if !ok {
				return aCfg, fmt.Errorf("invalid config parameter %s", e)
			}
			aCfg.Exchanges = append(aCfg.Exchanges, val)
		}
	*/
	return aCfg, nil
}

func (c *AMQPNurv) connect() error {
	var err error
	c.logger.Debug.Printf("Dialing %q...\n", c.cfg.URI)
	c.conn, err = amqp.Dial(c.cfg.URI)
	if err != nil {
		return err
	}
	go func() {
		c.logger.Info.Printf("Closing - Error: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	c.logger.Debug.Printf("Connection established. Getting channel...\n")
	c.channel, err = c.conn.Channel()

	return err
}
