package dao

import "time"

// Stored domaijn record of crawling activities
type CrawlRecord struct {
	OriginPage  string
	CurrentPage string
	Timestamp   time.Time
	Body        string
	WorkerID    string
}

// A Storer can store information about the pages that we have visited
type Storer interface {
	Store(CrawlRecord) error
}

// Check if we have visited this website before
// We'll stick to using an in memory cache for now, but this could be easily swapped out for Redis
type Cache interface {
	HaveVisited(string) bool
	Set(string)
}
