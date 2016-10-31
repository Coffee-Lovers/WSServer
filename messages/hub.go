package messages

import (
	"github.com/streadway/amqp"
	"fmt"
	"log"
)

type Listener interface {
	Listen(chan <- Progress)
}

type RabbitMQ struct {
	syncChannel chan <- Progress
	Username string
	Password string
	Host string
	Port string
}

func (r RabbitMQ) Listen(channel chan Progress) {
	r.syncChannel = channel
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", r.Username, r.Password, r.Host, r.Port))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"coffeepot_progress", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,               // queue name
		"#",                // routing key
		"coffeepot_progress", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			if p, err := FromSerialized(d.Body); err == nil {
				channel <- p
			}
		}
	}()

	log.Print(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
