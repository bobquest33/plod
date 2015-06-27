package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/bitly/go-nsq"
	log "github.com/cihub/seelog"
	"github.com/sjwhitworth/plod/dao"
	"github.com/sjwhitworth/plod/worker"
)

var (
	handlers    = flag.Int("handlers", 10, "Number of concurrenct handlers")
	startDomain = flag.String("start", "https://www.github.com", "The starting URL")
	concurrency = flag.Int("concurrency", 5, "The theoretical maximum number of URLs we can crawl at a given time")
)

func main() {
	defer log.Flush()
	flag.Parse()

	log.Tracef("[Main] Initialising Cassandra")
	if err := dao.Init(); err != nil {
		panic(err)
	}

	log.Tracef("[Main] Running with %v handlers", *handlers)

	cfg := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("urls", "plod", cfg)
	if err != nil {
		panic(err)
	}

	w := worker.Worker{}
	consumer.ChangeMaxInFlight(*concurrency)
	consumer.AddConcurrentHandlers(w, *handlers)

	log.Tracef("[Main] Firing up NSQ")

	if err := consumer.ConnectToNSQD("192.168.59.103:4150"); err != nil {
		panic(err)
	}

	worker.Initialise(*startDomain)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block on an exit
	for {
		select {
		case <-c:
			log.Infof("[Main] Caught exit signal, bye!")
			return
		}
	}
}
