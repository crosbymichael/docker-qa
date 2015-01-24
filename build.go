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
	consumer.AddConcurrentHandlers(&handler{}, context.Int("c"))
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
}

func (h *handler) HandleMessage(m *nsq.Message) error {
	return exec.Command("docker", "build", "--force-rm", "-q", string(m.Body)).Run()
}
