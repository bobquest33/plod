## Plod

<img src="http://www.1stfortoys.co.uk/toys-90/noddy-toys-pc-plod-6342-2896.jpg" height=300>

Plod is a concurrent, distributed web crawling service, which will happily plod around the network graph of the internet, following the hyperlinks on every page. It is designed to be run on a cluster, backed by a message queue. It will only visit links it hasn't seen before, and will persist the body of the page to Cassandra. 

The current implementation is backed by NSQ. However, it should be extremely simplistic to swap implementations as most message queues desire a simple `Handler` interface. A Dockerfile for NSQ and Cassandra is included within the repo. Simply change the configuration to use a larger set of machines for running on a cluster.

### Dependencies

* Docker (if you want to run it locally, and can't be bothered to set up Cassandra and NSQ on your machine)
* An internet connection (you seem to have one already)

### Installing

    go get github.com/sjwhitworth/plod
    docker-compose build
    docker-compose up -d
    go build
    plod

### Options

    -start="https://www.github.com": The starting URL
    -workers=10: Number of workers to spawn

### To do

* Intelligent throughput. Right now, each worker can only crawl a maximum of two URL's a second.
* Distributed deduping of URLs. I only check in a local in memory cache to verify if we have visited this page before. An interface is provided, which could easily be swapped out for Memcached, Redis and the like.
* Cleaning up of the HTML body. It is dumped raw into C* currently. Cleaning it up, removing tags, etc would save time on analysis further down the line.
* Real time analysis to steer direction. An interesting application would be to feed a RNN with information gleaned from the web.
