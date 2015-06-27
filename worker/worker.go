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
)

type Worker struct {
	ID       string
	WorkChan chan domain.URLPair
}

func Spawn(q chan domain.URLPair) {
	// Set up HTTP transport for TLS connection, one time only
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
		ID:       id.String()[:5], // We don't need the whole ID as we're not running that many workers..
		WorkChan: q,
	}
	go w.work()
}

func (w Worker) work() {
	for {
		url := <-w.WorkChan
		err := w.crawl(url)
		if err != nil {
			log.Errorf("[worker-%v] Error crawling %v: %v", w.ID, url, err)
		}
	}
}

func (w Worker) crawl(urls domain.URLPair) error {
	log.Tracef("[worker-%v] Visiting %v from %v", w.ID, urls.CurrentURL, urls.OriginURL)
	dao.DefaultCache.Set(string(urls.CurrentURL))

	resp, err := client.Get(string(urls.CurrentURL))
	if err != nil {
		return err
	}

	hyperlinks := html.ParseLinks(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	count := 0
	for _, link := range hyperlinks {
		uri := html.FixURL(string(link), string(urls.CurrentURL))

		// Bad URI, skip this one.
		if uri == "" {
			continue
		}

		visited := dao.DefaultCache.HaveVisited(uri)
		if !visited {
			// Insert in a non blocking fashion.
			go func() {
				w.WorkChan <- domain.URLPair{
					OriginURL:  urls.CurrentURL,
					CurrentURL: domain.URL(uri),
				}
				count++
			}()
		}
	}

	log.Tracef("[worker-%v] Found %v links from %v, %v in queue", w.ID, count, urls.CurrentURL, len(w.WorkChan))

	// Store information in C* about the crawl
	return dao.Store(&dao.CrawlRecord{
		OriginPage:  string(urls.OriginURL),
		CurrentPage: string(urls.CurrentURL),
		Timestamp:   time.Now(),
		Body:        string(body),
	})
}
