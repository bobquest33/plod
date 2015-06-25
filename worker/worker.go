package worker

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/nu7hatch/gouuid"
	"github.com/sjwhitworth/plod/dao"
	"github.com/sjwhitworth/plod/domain"
	"github.com/sjwhitworth/plod/html"
)

var (
	once      sync.Once
	client    http.Client
	transport *http.Transport
	tlsConfig *tls.Config

	mu    sync.Mutex
	cache = make(map[string]string)
)

type Worker struct {
	ID       string
	WorkChan chan domain.URLPair
	sleep    time.Duration
}

func Spawn(q chan domain.URLPair) {
	once.Do(func() {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		client = http.Client{Transport: transport}
	})

	id, _ := uuid.NewV4()
	w := Worker{
		ID:       id.String()[:5],
		WorkChan: q,
		sleep:    1000 * time.Millisecond,
	}
	go w.work()
}

func (w Worker) work() {
	ticker := time.NewTicker(w.sleep)
	for {
		url := <-w.WorkChan
		err := w.crawl(url)
		if err != nil {
			log.Errorf("[worker-%v] Error crawling %v: %v", w.ID, url, err)
		}
		<-ticker.C
	}
}

func (w Worker) crawl(urls domain.URLPair) error {
	mu.Lock()
	count := 0
	cache[string(urls.CurrentURL)] = ""
	mu.Unlock()

	log.Tracef("[worker-%v] Visiting %v from %v", w.ID, urls.CurrentURL, urls.OriginURL)

	resp, err := client.Get(string(urls.CurrentURL))
	if err != nil {
		return err
	}

	links := html.Parse(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	for _, link := range links {
		uri := html.FixURL(string(link), string(urls.CurrentURL))

		if !visited(uri) {
			if uri != "" {
				go func() {
					w.WorkChan <- domain.URLPair{
						OriginURL:  urls.CurrentURL,
						CurrentURL: domain.URL(uri),
					}
					count++
				}()
			}
		}
	}

	log.Tracef("[worker-%v] Found %v links from %v, %v in queue", w.ID, count, urls.CurrentURL, len(w.WorkChan))

	return dao.Store(&dao.CrawlRecord{
		OriginPage:  string(urls.OriginURL),
		CurrentPage: string(urls.CurrentURL),
		Timestamp:   time.Now(),
		Body:        string(body),
	})
}

func visited(url string) bool {
	mu.Lock()
	_, visited := cache[url]
	mu.Unlock()

	return visited
}
