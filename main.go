package main

import (
	log "github.com/cihub/seelog"
	"github.com/sjwhitworth/plod/dao"
	"github.com/sjwhitworth/plod/domain"
	"github.com/sjwhitworth/plod/worker"
)

func main() {
	log.Tracef("[Main] Initialising C*")
	if err := dao.Init(); err != nil {
		panic(err)
	}

	log.Tracef("[Main] Spawning workers")
	q := make(chan domain.URLPair, 10000)
	for i := 0; i < 3; i++ {
		worker.Spawn(q)
	}
	q <- domain.URLPair{domain.URL("START"), domain.URL("http://www.github.com")}

	log.Tracef("[Main] Spawned %v workers", 10)
	log.Tracef("[Main] Entering crawling loop!")
	select {}
}
