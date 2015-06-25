package dao

import (
	log "github.com/cihub/seelog"
	"github.com/hailocab/gocassa"
	"github.com/unravelin/arx/errors"
)

var (
	Keyspace = "plod"
	KS       gocassa.KeySpace
	t        gocassa.MultimapTable
)

func Init() error {
	var err error
	hosts := []string{"127.0.0.1"}
	conn, err := gocassa.Connect(hosts, "", "")
	if err != nil {
		panic(err)
	}

	// Init the keyspace
	err = conn.CreateKeySpace(Keyspace)
	if err != nil && !errors.CheckErr("Cannot add existing keyspace", err) {
		log.Critical(err)
	}

	// KS variable used globally
	KS = conn.KeySpace(Keyspace)
	t = KS.MultimapTable("domainrecords", "OriginPage", "CurrentPage", &CrawlRecord{})

	if exists, _ := KS.Exists(t.Name()); !exists {
		return t.Create()
	}
	return nil
}

func Store(cr *CrawlRecord) error {
	return t.Set(cr).Run()
}
