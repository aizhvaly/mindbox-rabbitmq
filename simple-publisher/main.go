package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/streadway/amqp"
)

var (
	data    = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	queue   int
	timeout int
	user    string
	passwd  string
	host    string
)

func init() {

	pflag.IntVar(&queue, "q", 10, "number of queues")
	pflag.IntVar(&timeout, "t", 0, "timeout")
	pflag.StringVar(&host, "h", "127.0.0.1:5672", "rabbitmq host:port")
	pflag.StringVar(&user, "u", "guest", "rabbitmq user")
	pflag.StringVar(&passwd, "p", "guest", "rabbitmq password")
}

// DataGen Generate random string
func DataGen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = data[rand.Intn(len(data))]
	}
	return string(b)
}

// MsgPublish Connect and generate load to RabbitMQ
func MsgPublish(queue, host, user, passwd string, timeout int) error {

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

	for {
		err = ch.Publish(
			"",    // exchange
			queue, // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(DataGen(len(data))),
			})
		if err != nil {
			return err
		}
		if timeout != 0 {
			time.Sleep(time.Duration(timeout) * time.Millisecond)
		}
	}
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
			if err := MsgPublish(q, host, user, passwd, timeout); err != nil {
				log.Fatal(err)
			}
		}()
	}
	<-chNoEnd
}
