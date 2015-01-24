package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
)

var buildCommand = cli.Command{
	Name:  "build",
	Usage: "build git repositores",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "nsqd", Usage: "nsqd address"},
		cli.StringFlag{Name: "topic", Usage: "topic to subscribe to"},
		cli.StringFlag{Name: "channel", Usage: "channel to identify as"},
		cli.DurationFlag{Name: "timeout", Value: 5 * time.Minute, Usage: "set the message build timeout"},
		cli.IntFlag{Name: "concurrency,c", Value: 2, Usage: "number of concurrent builds to process"},
		cli.StringFlag{Name: "error-topic", Usage: "topic to publish errors to"},
	},
	Action: buildAction,
}

func buildAction(context *cli.Context) {
	signals := make(chan os.Signal, 128)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	config := nsq.NewConfig()
	config.MsgTimeout = context.Duration("timeout")
	consumer, err := nsq.NewConsumer(context.String("topic"), context.String("channel"), config)
	if err != nil {
		log.Fatal(err)
	}
	producer, err := nsq.NewProducer(context.String("nsqd"), config)
	if err != nil {
		log.Fatal(err)
	}
	consumer.AddConcurrentHandlers(&handler{
		producer: producer,
		topic:    context.String("error-topic"),
	}, context.Int("c"))
	if err := consumer.ConnectToNSQD(context.String("nsqd")); err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-consumer.StopChan:
			return
		case <-signals:
			consumer.Stop()
		}
	}
}

type handler struct {
	// have a handle on the producer to report errors back in a way that we can view
	producer *nsq.Producer
	topic    string
}

func (h *handler) HandleMessage(m *nsq.Message) error {
	if err := exec.Command("docker", "build", "--force-rm", "-q", string(m.Body)).Run(); err != nil {
		if err := h.producer.Publish(h.topic, m.Body); err != nil {
			return err
		}
	}
	return nil
}
