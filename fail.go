package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
)

var failCommand = cli.Command{
	Name:  "fail",
	Usage: "get a failed git repository",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "nsqd", Usage: "nsqd address"},
		cli.StringFlag{Name: "topic", Usage: "topic to receive failures from"},
		cli.StringFlag{Name: "channel", Value: "default", Usage: "channel to identify as"},
	},
	Action: failAction,
}

func failAction(context *cli.Context) {
	consumer, err := nsq.NewConsumer(context.String("topic"), context.String("channel"), nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	consumer.SetLogger(log.New(ioutil.Discard, "", log.Flags()), nsq.LogLevelInfo)
	consumer.AddConcurrentHandlers(&failHandler{consumer}, 1)
	if err := consumer.ConnectToNSQD(context.String("nsqd")); err != nil {
		log.Fatal(err)
	}
	<-consumer.StopChan
}

type failHandler struct {
	stopper interface {
		Stop()
	}
}

func (h *failHandler) HandleMessage(m *nsq.Message) error {
	var f *failure
	if err := json.Unmarshal(m.Body, &f); err != nil {
		return err
	}
	fmt.Printf("URL:%s\nCHANNEL: %s\n\n", f.Url, f.Channel)
	fmt.Printf("LOG:\n%s", f.Log)
	h.stopper.Stop()
	return nil
}
