package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn          *amqp.Connection
	Channel       *amqp.Channel
	Ctx           context.Context
	Cancel        context.CancelFunc
	Queue         amqp.Queue
	mu            sync.RWMutex
	cfg           *models.Config
	connected     bool
	reconnecting  bool
	monitoringStarted bool
	messageQueue  []queuedMessage
	queueMu       sync.Mutex
	Exchange      string
	ExchangeType  string
}

type queuedMessage struct {
	message any
	queue   string
}

func (r *RabbitMQ) Connect(cfg *models.Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Connecting to RabbitMQ", utils.Field{Key: "host", Value: cfg.RMQ.Host}, utils.Field{Key: "port", Value: cfg.RMQ.Port}, utils.Field{Key: "queue", Value: cfg.RMQ.Queue})
	r.cfg = cfg
	if r.Ctx == nil || r.Cancel == nil {
		r.Ctx, r.Cancel = context.WithCancel(context.Background())
	}
	err := r.connect()
	if err != nil {
		logger.ErrorWithErr("Initial RabbitMQ connection failed", err)
		r.connected = false
	} else {
		logger.Info("Initial connection successful")
		r.connected = true
		r.flushMessageQueue()
	}
	if !r.monitoringStarted {
		r.monitoringStarted = true
		go r.monitorConnection()
	}
}

func (r *RabbitMQ) connect() error {
	var err error
	r.Conn, err = amqp.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s:%d/",
			r.cfg.RMQ.User,
			r.cfg.RMQ.Pass,
			r.cfg.RMQ.Host,
			r.cfg.RMQ.Port,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	r.Channel, err = r.Conn.Channel()
	if err != nil {
		r.Conn.Close()
		return fmt.Errorf("failed to open channel: %v", err)
	}
	if r.cfg.RMQ.Exchange != "" {
		r.Exchange = r.cfg.RMQ.Exchange
		r.ExchangeType = "topic"
		err = r.Channel.ExchangeDeclare(
			r.Exchange,
			r.ExchangeType,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %v", err)
		}
	}
	r.Queue, err = r.Channel.QueueDeclare(
		r.cfg.RMQ.Queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.Channel.Close()
		r.Conn.Close()
		return fmt.Errorf("failed to declare queue: %v", err)
	}
	logger := utils.GetLogger("rabbitmq")
	logger.Info("RabbitMQ connection established successfully", utils.Field{Key: "queue", Value: r.cfg.RMQ.Queue})
	return nil
}

func (r *RabbitMQ) monitorConnection() {
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Started monitoring RabbitMQ connection")
	for {
		select {
		case <-r.Ctx.Done():
			logger.Info("Context cancelled, stopping RabbitMQ monitoring")
			return
		default:
			r.mu.RLock()
			conn := r.Conn
			connected := r.connected
			r.mu.RUnlock()
			if !connected || conn == nil {
				logger.Warn("Not connected to RabbitMQ, attempting reconnect")
				r.reconnect()
				time.Sleep(5 * time.Second)
				continue
			}
			closeChan := conn.NotifyClose(make(chan *amqp.Error))
			select {
			case <-r.Ctx.Done():
				return
			case err := <-closeChan:
				logger.ErrorWithErr("RabbitMQ connection closed", err)
				r.mu.Lock()
				r.connected = false
				r.mu.Unlock()
				r.reconnect()
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Closing RabbitMQ connection")
	r.connected = false
	if r.Cancel != nil {
		r.Cancel()
	}
	if r.Channel != nil && !r.Channel.IsClosed() {
		r.Channel.Close()
		logger.Info("RabbitMQ channel closed")
	}
	if r.Conn != nil && !r.Conn.IsClosed() {
		r.Conn.Close()
		logger.Info("RabbitMQ connection closed")
	}
	logger.Info("RabbitMQ connection closed completely")
}

func (r *RabbitMQ) Publish(message any, queues ...string) {
	queue := r.Queue.Name
	if len(queues) > 0 {
		queue = queues[0]
	}
	r.mu.RLock()
	connected := r.connected
	r.mu.RUnlock()
	if !connected {
		logger := utils.GetLogger("rabbitmq")
		logger.Warn("Not connected to RabbitMQ, queuing message for queue", utils.Field{Key: "queue", Value: queue})
		r.queueMessage(message, queue)
		return
	}
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Publishing message to RabbitMQ queue", utils.Field{Key: "queue", Value: queue})
	if r.publishMessage(message, queue) != nil {
		r.queueMessage(message, queue)
	}
}

func (r *RabbitMQ) publishMessage(message any, queue string) error {
	body, _ := json.Marshal(message)
	r.mu.RLock()
	channel := r.Channel
	r.mu.RUnlock()
	if channel == nil {
		return fmt.Errorf("channel is nil")
	}
	err := channel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "utf-8",
			Body:            body,
		},
	)
	if err != nil {
		logger := utils.GetLogger("rabbitmq")
		logger.ErrorWithErr("Failed to publish message to RabbitMQ queue", err)
		r.mu.Lock()
		r.connected = false
		r.mu.Unlock()
		return err
	}
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Message published to RabbitMQ queue", utils.Field{Key: "queue", Value: queue})
	return nil
}

func (r *RabbitMQ) queueMessage(message any, queue string) {
	r.queueMu.Lock()
	defer r.queueMu.Unlock()
	r.messageQueue = append(r.messageQueue, queuedMessage{
		message: message,
		queue:   queue,
	})
	logger := utils.GetLogger("rabbitmq")
	logger.Debug("Queued message for RabbitMQ queue", utils.Field{Key: "queue", Value: queue}, utils.Field{Key: "message_type", Value: fmt.Sprintf("%T", message)})
}

func (r *RabbitMQ) flushMessageQueue() {
	r.queueMu.Lock()
	messages := make([]queuedMessage, len(r.messageQueue))
	copy(messages, r.messageQueue)
	r.messageQueue = nil
	r.queueMu.Unlock()
	if len(messages) > 0 {
		logger := utils.GetLogger("rabbitmq")
		logger.Debug("Flushing RabbitMQ message queue", utils.Field{Key: "count", Value: len(messages)})
		for _, msg := range messages {
			if r.publishMessage(msg.message, msg.queue) != nil {
				r.queueMessage(msg.message, msg.queue)
			}
		}
	}
}

func (r *RabbitMQ) reconnect() {
	r.mu.Lock()
	if r.reconnecting {
		r.mu.Unlock()
		return
	}
	r.reconnecting = true
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		r.reconnecting = false
		r.mu.Unlock()
	}()
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Attempting to reconnect to RabbitMQ")
	if r.Channel != nil && !r.Channel.IsClosed() {
		r.Channel.Close()
	}
	if r.Conn != nil && !r.Conn.IsClosed() {
		r.Conn.Close()
	}
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		select {
		case <-r.Ctx.Done():
			logger.Warn("Context cancelled during RabbitMQ reconnection")
			return
		default:
		}
		err := r.connect()
		if err == nil {
			r.mu.Lock()
			r.connected = true
			r.mu.Unlock()
			logger.Info("Successfully reconnected to RabbitMQ", utils.Field{Key: "attempts", Value: i + 1})
			r.flushMessageQueue()
			return
		}
		waitTime := time.Duration(min(i+1, 5)) * time.Second
		logger.Warn("Attempt failed to reconnect to RabbitMQ", utils.Field{Key: "attempt", Value: i + 1}, utils.Field{Key: "error", Value: err.Error()}, utils.Field{Key: "retry_delay", Value: waitTime})
		time.Sleep(waitTime)
	}
	logger.Error("Failed to reconnect to RabbitMQ", utils.Field{Key: "max_attempts", Value: maxRetries})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *RabbitMQ) PublishWithExchange(message any, routingKeys ...string) {
	body, _ := json.Marshal(message)
	if len(routingKeys) == 0 {
		utils.FailOnError(fmt.Errorf("routing key is required"), "Failed to publish a message with exchange")
	}
	exchange := r.Exchange
	if exchange == "" {
		utils.FailOnError(fmt.Errorf("exchange is not set"), "Failed to publish a message with exchange")
	}
	routingKey := routingKeys[0]
	err := r.Channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "utf-8",
			Body:            body,
		},
	)
	utils.LogError(err, "Failed to publish a message with exchange")
}

func (r *RabbitMQ) Consume(messageConsumer func(<-chan amqp.Delivery)) error {
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Starting RabbitMQ consumer for queue", utils.Field{Key: "queue", Value: r.Queue.Name})
	for !r.connected {
		logger.Info("Waiting for initial connection...")
		time.Sleep(2 * time.Second)
	}
	for {
		select {
		case <-r.Ctx.Done():
			logger.Info("Context cancelled, stopping RabbitMQ consumer")
			return nil
		default:
		}
		r.mu.RLock()
		connected := r.connected
		channel := r.Channel
		queueName := r.Queue.Name
		r.mu.RUnlock()
		if !connected || channel == nil {
			logger.Warn("Waiting for RabbitMQ connection to be established...")
			time.Sleep(1 * time.Second)
			continue
		}
		logger.Info("Starting consumer for RabbitMQ queue", utils.Field{Key: "queue", Value: queueName})
		msgs, err := channel.Consume(
			queueName,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			logger.ErrorWithErr("Failed to register RabbitMQ consumer", err)
			time.Sleep(2 * time.Second)
			continue
		}
		consumerDone := make(chan bool)
		go func() {
			messageConsumer(msgs)
			consumerDone <- true
			logger.Warn("Message channel closed for RabbitMQ consumer, will restart")
		}()
		connectionClosed := channel.NotifyClose(make(chan *amqp.Error))
		select {
		case <-r.Ctx.Done():
			return nil
		case <-consumerDone:
			logger.Warn("RabbitMQ consumer connection closed, restarting")
		case err := <-connectionClosed:
			logger.Info("Channel closed, restarting", utils.Field{Key: "error", Value: err.Error()})
		}
		time.Sleep(1 * time.Second)
	}
}

func (r *RabbitMQ) ConsumeWithExchange(messageConsumer func(<-chan amqp.Delivery), bindKeys ...string) error {
	if len(bindKeys) == 0 {
		return fmt.Errorf("routing key is required")
	}
	logger := utils.GetLogger("rabbitmq")
	for !r.connected {
		logger.Info("Waiting for initial connection...")
		time.Sleep(2 * time.Second)
	}
	for {
		select {
		case <-r.Ctx.Done():
			logger.Info("Context cancelled, stopping RabbitMQ exchange consumer")
			return nil
		default:
		}
		r.mu.RLock()
		connected := r.connected
		channel := r.Channel
		queueName := r.Queue.Name
		exchange := r.Exchange
		r.mu.RUnlock()
		if !connected || channel == nil || channel.IsClosed() {
			if channel != nil && channel.IsClosed() {
				logger.Warn("RabbitMQ channel is closed; triggering reconnect")
				r.mu.Lock()
				r.connected = false
				r.mu.Unlock()
				go r.reconnect()
			} else {
				logger.Warn("Waiting for RabbitMQ connection to be established...")
			}
			time.Sleep(1 * time.Second)
			continue
		}
		if exchange == "" {
			return fmt.Errorf("exchange is not set")
		}
		for _, key := range bindKeys {
			if err := channel.QueueBind(queueName, key, exchange, false, nil); err != nil {
				logger.ErrorWithErr("Failed to bind queue to exchange", err)
				r.mu.Lock()
				r.connected = false
				r.mu.Unlock()
				go r.reconnect()
				time.Sleep(2 * time.Second)
				continue
			}
		}
		logger.Info("Starting consumer for RabbitMQ exchange",
			utils.Field{Key: "exchange", Value: exchange},
			utils.Field{Key: "queue", Value: queueName},
			utils.Field{Key: "bind_keys", Value: bindKeys},
		)
		msgs, err := channel.Consume(
			queueName,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			logger.ErrorWithErr("Failed to register RabbitMQ exchange consumer", err)
			r.mu.Lock()
			r.connected = false
			r.mu.Unlock()
			go r.reconnect()
			time.Sleep(2 * time.Second)
			continue
		}
		consumerDone := make(chan bool)
		go func() {
			messageConsumer(msgs)
			consumerDone <- true
			logger.Warn("Message channel closed for RabbitMQ exchange consumer, will restart")
		}()
		connectionClosed := channel.NotifyClose(make(chan *amqp.Error))
		select {
		case <-r.Ctx.Done():
			return nil
		case <-consumerDone:
			logger.Warn("RabbitMQ exchange consumer closed; restarting")
			r.mu.Lock()
			r.connected = false
			r.mu.Unlock()
			go r.reconnect()
		case err := <-connectionClosed:
			if err != nil {
				logger.Info("Channel closed; restarting", utils.Field{Key: "error", Value: err.Error()})
			} else {
				logger.Info("Channel closed; restarting")
			}
			r.mu.Lock()
			r.connected = false
			r.mu.Unlock()
			go r.reconnect()
		}
		time.Sleep(1 * time.Second)
	}
}

var RMQ *RabbitMQ

func InitRabbitMQ(cfg *models.Config) {
	logger := utils.GetLogger("rabbitmq")
	logger.Info("Initializing RabbitMQ client")
	RMQ = &RabbitMQ{}
	RMQ.Connect(cfg)
	logger.Info("RabbitMQ client initialized")
}