package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
)

var pushCommand = cli.Command{
	Name:  "push",
	Usage: "push a CSV file of repositories onto the build queue",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "nsqd", Usage: "nsqd address"},
		cli.StringFlag{Name: "topic", Usage: "topic to push messages ontop"},
		cli.BoolFlag{Name: "verbose,v", Usage: "enable verbose output"},
	},
	Action: pushAction,
}

func pushAction(context *cli.Context) {
	file := context.Args().Get(0)
	if file == "" {
		log.Fatal("file not specified")
	}
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	producer, err := nsq.NewProducer(context.String("nsqd"), nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	s := bufio.NewScanner(f)
	i := 0
	for s.Scan() {
		i++
		u, err := url.Parse(strings.Replace(s.Text(), ".git", "", -1))
		if err != nil {
			log.Error(err)
			continue
		}
		if err := producer.Publish(context.String("topic"), []byte(fmt.Sprintf("%s%s", u.Host, u.Path))); err != nil {
			log.Error(err)
			continue
		}
		if context.Bool("verbose") {
			log.Printf("ADD: %d. %s", i, s.Text())
		}
	}
}
