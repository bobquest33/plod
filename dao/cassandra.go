package dao

import (
	"github.com/hailocab/gocassa"
	"github.com/unravelin/arx/errors"
)

var (
	Keyspace = "plod"
	KS       gocassa.KeySpace
	t        gocassa.MultimapTable
)

// Initialise the keyspace in C*, and create the table
func Init() error {
	var err error
	hosts := []string{"192.168.59.103:9042"}
	conn, err := gocassa.Connect(hosts, "", "")
	if err != nil {
		panic(err)
	}

	// Init the keyspace
	err = conn.CreateKeySpace(Keyspace)
	if err != nil && !errors.CheckErr("Cannot add existing keyspace", err) {
		panic(err)
	}

	// KS variable used globally
	KS = conn.KeySpace(Keyspace)
	t = KS.MultimapTable("domainrecords", "OriginPage", "CurrentPage", &CrawlRecord{})

	if exists, _ := KS.Exists(t.Name()); !exists {
		return t.Create()
	}
	return nil
}

// Store the CrawlRecord in Cassandra.
func Store(cr *CrawlRecord) error {
	return t.Set(cr).Run()
}
