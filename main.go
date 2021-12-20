package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var queueName = "hello"

func init() {
	godotenv.Load()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connect() *amqp.Connection {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	url := fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	return conn
}

func getChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func getQueue(ch *amqp.Channel, name string) amqp.Queue {
	q, err := ch.QueueDeclare(
		name,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q
}

func produceMessage(ch *amqp.Channel, q amqp.Queue, body string) {
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}

func consumeMessages(ch *amqp.Channel, q amqp.Queue) <-chan amqp.Delivery {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	return msgs
}

func produce() {
	msgs := []string{"Welcome", "to", "RabbitMQ", "Hello", "World", "example"}

	conn := connect()
	defer conn.Close()

	ch := getChannel(conn)
	defer ch.Close()

	q := getQueue(ch, queueName)

	for _, msg := range msgs {
		produceMessage(ch, q, msg)
	}
}

func consume() {
	conn := connect()
	defer conn.Close()

	ch := getChannel(conn)
	defer ch.Close()

	q := getQueue(ch, queueName)

	msgs := consumeMessages(ch, q)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func main() {
	produce()
	consume()
}
