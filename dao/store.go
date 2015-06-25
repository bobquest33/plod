package dao

import "time"

type CrawlRecord struct {
	OriginPage  string
	CurrentPage string
	Timestamp   time.Time
	Body        string
	WorkerID    string
}

type Storer interface {
	Store(CrawlRecord) error
}
