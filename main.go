package main

import (
	"flag"
	"os"
	"os/signal"

	log "github.com/cihub/seelog"
	"github.com/sjwhitworth/plod/dao"
	"github.com/sjwhitworth/plod/domain"
	"github.com/sjwhitworth/plod/worker"
)

var (
	workers     = flag.Int("workers", 10, "Number of workers to spawn")
	startDomain = flag.String("start", "https://www.github.com", "The starting URL")
)

func main() {
	defer log.Flush()
	flag.Parse()

	log.Tracef("[Main] Initialising C*")
	if err := dao.Init(); err != nil {
		panic(err)
	}

	q := make(chan domain.URLPair)

	log.Tracef("Spawning %v workers", *workers)
	for i := 0; i < *workers; i++ {
		worker.Spawn(q)
	}

	q <- domain.URLPair{domain.URL("START"), domain.URL(*startDomain)}
	log.Tracef("[Main] Entering crawling loop")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block on an exit
	for {
		select {
		case <-c:
			log.Infof("[Main] Caught exit signal, signalling for workers to finish up cleanly")
			return
		}
	}
}
