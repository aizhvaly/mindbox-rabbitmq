package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/streadway/amqp"
)

var (
	queue  int
	user   string
	passwd string
	host   string
)

func init() {

	pflag.IntVar(&queue, "q", 10, "number of queues")
	pflag.StringVar(&host, "h", "127.0.0.1:5672", "rabbitmq host:port")
	pflag.StringVar(&user, "u", "guest", "rabbitmq user")
	pflag.StringVar(&passwd, "p", "guest", "rabbitmq password")
}

// Consumer ...
func Consumer(queue, host, user, passwd string) error {

	conn, err := amqp.Dial("amqp://" + user + ":" + passwd + "@" + host + "/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}

func main() {

	pflag.Parse()

	chNoEnd := make(chan bool)
	chQ := make(chan string, queue)

	for i := 0; i < queue; i++ {
		chQ <- "q" + strconv.Itoa(i)
	}
	// user = os.Getenv("RABBITMQ_USER")
	// passwd = os.Getenv("RABBITMQ_PASSWD")

	for i := 0; i < queue; i++ {
		go func() {
			q := <-chQ
			fmt.Printf(q)
			if err := Consumer(q, host, user, passwd); err != nil {
				log.Fatal(err)
			}
		}()
	}
	<-chNoEnd
}
