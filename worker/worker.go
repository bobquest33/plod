package worker

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/bitly/go-nsq"

	log "github.com/cihub/seelog"
	"github.com/sjwhitworth/plod/dao"
	"github.com/sjwhitworth/plod/domain"
	"github.com/sjwhitworth/plod/html"
)

var (
	once      sync.Once
	client    http.Client
	transport *http.Transport
	tlsConfig *tls.Config

	producer *nsq.Producer
)

const (
	NSQTopic = "urls"
)

func init() {
	once.Do(func() {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		client = http.Client{Transport: transport}

		var err error
		cfg := nsq.NewConfig()
		producer, err = nsq.NewProducer("192.168.59.103:4150", cfg)
		if err != nil {
			panic(err)
		}
	})
}

type Worker struct{}

func (w Worker) HandleMessage(msg *nsq.Message) error {
	var urls domain.URLPair
	body := msg.Body
	if err := json.Unmarshal(body, &urls); err != nil {
		return err
	}

	return w.crawl(urls)
}

func (w Worker) crawl(urls domain.URLPair) error {
	log.Tracef("Visiting %v from %v", urls.CurrentURL, urls.OriginURL)
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

	for _, link := range hyperlinks {
		uri := html.FixURL(string(link), string(urls.CurrentURL))

		// Bad URI, skip this one.
		if uri == "" {
			continue
		}

		visited := dao.DefaultCache.HaveVisited(uri)
		if !visited {
			d := domain.URLPair{
				OriginURL:  urls.CurrentURL,
				CurrentURL: uri,
			}

			dat, err := json.Marshal(&d)
			if err != nil {
				return err
			}

			if err := producer.Publish("urls", dat); err != nil {
				return err
			}
		}
	}

	// Store information in C* about the crawl
	return dao.Store(&dao.CrawlRecord{
		OriginPage:  string(urls.OriginURL),
		CurrentPage: string(urls.CurrentURL),
		Timestamp:   time.Now(),
		Body:        string(body),
	})
}

// Sends the first URL to kick things off!
func Initialise(url string) {
	d := domain.URLPair{"START", url}
	dat, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	if err := producer.Publish(NSQTopic, dat); err != nil {
		panic(err)
	}

}
