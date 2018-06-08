package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/vst"
	vstproto "github.com/arangodb/go-driver/vst/protocol"
	"log"
	"sort"
	"sync"
	"time"
)

type Book struct {
	Title   string `json:"title"`
	NoPages int    `json:"no_pages"`
}

var (
	nrConnections int    = 1
	endpoint      string = "http://127.0.0.1:8529"
	testcase      string = "postDocs"
	nrRequests    int    = 1000
	parallelism   int    = 1
	cleanup       bool   = true
	protocol      string = "HTTP" // can be "VST" as well
	usetls        bool   = false
)

func logStats(name string, times []time.Duration) {
	nr := len(times)
	if nr == 0 {
		return
	}
	sort.Slice(times, func(a, b int) bool {
		return int64(times[a]) < int64(times[b])
	})
	var sum int64 = 0
	for i := 0; i < nr; i++ {
		sum += int64(times[i])
	}
	log.Printf("Statistics for %s:", name)
	log.Printf("Number of samples: %d", nr)
	log.Printf("Average: %v", time.Duration(sum/int64(nr)))
	log.Printf("Median : %v", times[nr/2])
	log.Printf("90%%   : %v", times[(nr*90)/100])
	log.Printf("99%%   : %v", times[(nr*99)/100])
	log.Printf("99.9%% : %v", times[(nr*999)/1000])
	if nr >= 20 {
		log.Printf("Smallest: ")
		for i := 0; i < 10; i++ {
			log.Printf("%v ", times[i])
		}
		log.Printf("Largest: ")
		for i := 10; i > 0; i-- {
			log.Printf("%v ", times[nr-i])
		}
	}
}

func doPostDocs(col driver.Collection) {
	// Create documents
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(times []time.Duration, nrRequestsPerWorker int) {
		defer wg.Done()
		for i := 0; i < nrRequestsPerWorker; i++ {
			startTime := time.Now()
			book := Book{
				Title:   "Some small string",
				NoPages: i,
			}
			_, err := col.CreateDocument(nil, book)
			if err != nil {
				log.Fatalf("Failed to create document: %v", err)
			}
			endTime := time.Now()
			times[i] = endTime.Sub(startTime)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go worker(times[j*nrRequestsPerWorker:(j+1)*nrRequestsPerWorker],
			nrRequestsPerWorker)
	}

	wg.Wait()
	log.Printf("Hallo Welt")
	logStats("create document ops", times)
}

// For later:
//	// Read the document back
//	var result Book
//	if _, err := col.ReadDocument(nil, meta.Key, &result); err != nil {
//		log.Fatalf("Failed to read document: %v", err)
//	}
//	fmt.Printf("Read book '%+v'\n", result)
//
//	// Get books by using AQL
//	cur, err := db.Query(nil, "FOR b IN books RETURN b", nil)
//	if err != nil {
//		log.Fatalf("Failed to ask for query cursor: %v", err)
//	}
//	for {
//		meta, err = cur.ReadDocument(nil, &result)
//		if err != nil {
//			if driver.IsNoMoreDocuments(err) {
//				break
//			}
//			log.Fatalf("Failed to read doc from cursor: %v", err)
//		}
//		fmt.Printf("Read book from cursor '%+v'\n", result)
//	}

func main() {
	flag.IntVar(&nrConnections, "nrConnections", nrConnections, "number of connections")
	flag.StringVar(&endpoint, "endpoint", endpoint, "server endpoint")
	flag.StringVar(&testcase, "testcase", testcase, "test case")
	flag.IntVar(&nrRequests, "nrRequests", nrRequests, "number of requests")
	flag.IntVar(&parallelism, "parallelism", parallelism, "parallelism")
	flag.BoolVar(&cleanup, "cleanup", cleanup, "flag whether to perform cleanup")
	flag.StringVar(&protocol, "protocol", protocol, "protocol: HTTP or VST")
	flag.BoolVar(&usetls, "useTLS", usetls, "flag whether to use TLS")
	flag.Parse()

	fmt.Printf("Server endpoint: %s   ", endpoint)
	fmt.Printf("using %d connections\n", nrConnections)

	var conn driver.Connection
	var err error
	if protocol == "HTTP" {
		connConfig := http.ConnectionConfig{
			Endpoints: []string{endpoint},
			ConnLimit: nrConnections,
		}
		if usetls {
			connConfig.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		conn, err = http.NewConnection(connConfig)
		if err != nil {
			log.Fatalf("Failed to create HTTP connection: %v", err)
		}
	} else if protocol == "VST" {
		connConfig := vst.ConnectionConfig{
			Endpoints: []string{endpoint},
			Transport: vstproto.TransportConfig{
				ConnLimit: nrConnections,
			},
		}
		if usetls {
			connConfig.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		conn, err = vst.NewConnection(connConfig)
		if err != nil {
			log.Fatalf("Failed to create HTTP connection: %v", err)
		}
	} else {
		log.Fatalf("-protocol needs to be HTTP or VST")
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})

	db, err := c.Database(nil, "benchDB")
	if err != nil {
		// Create a database
		db, err = c.CreateDatabase(nil, "benchDB", nil)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
	}

	// Create collection
	col, err := db.Collection(nil, "test")
	if err != nil {
		col, err = db.CreateCollection(nil, "test", nil)
		if err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}
	}

	startTime := time.Now()
	switch testcase {
	case "postDocs":
		doPostDocs(col)
	}
	endTime := time.Now()

	log.Printf("Time for %d requests: %v\n", nrRequests, endTime.Sub(startTime))

	if cleanup {
		err = col.Remove(nil)
		if err != nil {
			log.Fatalf("Failed to drop collection: %v", err)
		}
		err = db.Remove(nil)
		if err != nil {
			log.Fatalf("Failed to drop database: %v", err)
		}
	}
}
