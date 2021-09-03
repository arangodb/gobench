package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"

	//"net"
	"sort"
	"strconv"
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/arangodb/go-driver/vst"
	vstproto "github.com/arangodb/go-driver/vst/protocol"
	"golang.org/x/net/http2"
)

// Book is the basic data structure used for tests
type Book struct {
	Key     string `json:"_key,omitempty"`
	Title   string `json:"title"`
	NoPages int    `json:"no_pages"`
}

var (
	nrConnections int           = 1
	endpoint      string        = "http://127.0.0.1:8529"
	testcase      string        = "postDocs"
	replFactor    int           = 1
	nrRequests    int           = 1000
	parallelism   int           = 1
	delay         time.Duration = 0
	cleanup       bool          = true
	protocol      string        = "HTTP" // can be "VST" as well
	usetls        bool          = false
	username      string
	password      string
	outputFormat  string = "console" // can be "csv"

	submittedRequests int = 0 // the number of requests submitted
)

func logStats(name string, times []time.Duration) {
	if outputFormat == "console" {
		logStatsConsole(name, times)
	} else if outputFormat == "csv" {
		logStatsCSV(name, times)
	} else {
		log.Fatalf("unknown output format %s", outputFormat)
	}
}

// Output log stats as comma separated values, the columns are
//
// - test name
// - average time taken
// - median time
// - minimum
// - maximum
// - standard deviation
//
// all timings are in microseconds
//
func logStatsCSV(name string, times []time.Duration) {
	nr := len(times)
	if nr == 0 {
		return
	}
	sort.Slice(times, func(a, b int) bool {
		return int64(times[a]) < int64(times[b])
	})
	var sum time.Duration
	for _, d := range times {
		sum += d
	}
	var mean = (sum / time.Duration(nr)).Nanoseconds() / 1000
	var sqrdiff float64
	for _, d := range times {
		var tmp = float64(d.Nanoseconds()/1000.0 - mean)
		sqrdiff += tmp * tmp
	}
	var stddev = math.Sqrt(sqrdiff / float64(nr))
	fmt.Printf("%s,%v,%v,%v,%v,%.2f,%s\n",
		name,                           // test name
		mean,                           // mean
		times[nr/2].Nanoseconds()/1000, // median
		times[0].Nanoseconds()/1000,    // minimum
		times[nr-1].Nanoseconds()/1000, // maximum
		stddev,                         // standard deviation
		"")                             // collection label
}

func logStatsConsole(name string, times []time.Duration) {
	nr := len(times)
	if nr == 0 {
		return
	}
	sort.Slice(times, func(a, b int) bool {
		return int64(times[a]) < int64(times[b])
	})
	var sum time.Duration
	for _, d := range times {
		sum += d
	}
	log.Printf("Statistics for %s:", name)
	log.Printf("Samples : %d", nr)
	log.Printf("Average : %v", sum/time.Duration(nr))
	log.Printf("Median  : %v", times[nr/2])
	log.Printf("90%%     : %v", times[(nr*90)/100])
	log.Printf("99%%     : %v", times[(nr*99)/100])
	log.Printf("99.9%%   : %v", times[(nr*999)/1000])
	if nr >= 20 {
		s := ""
		for i := 0; i < 10; i++ {
			s = s + fmt.Sprintf(" %v", times[i])
		}
		log.Printf("Smallest:%s", s)
		s = ""
		for i := 10; i > 0; i-- {
			s = s + fmt.Sprintf(" %v", times[nr-i])
		}
		log.Printf("Largest:%s", s)
	}
}

func doPostDocs(col driver.Collection) {
	// Create documents
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			book := Book{
				Key:     "",
				Title:   "Some small string",
				NoPages: i,
			}
			_, err := col.CreateDocument(nil, book)
			if err != nil {
				log.Fatalf("Failed to create document: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				initTime)
		}(j)
	}

	wg.Wait()
	logStats("create document ops", times)
}

func doSeedDocs(col driver.Collection) {
	// Create documents with specific keys
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			book := Book{
				Key:     "K" + strconv.Itoa(base+i),
				Title:   "Some small string",
				NoPages: i,
			}
			_, err := col.CreateDocument(nil, book)
			if err != nil {
				log.Fatalf("Failed to create document: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("seed document ops", times)
}

func doReadDocs(col driver.Collection) {
	// Read seeded documents with specific keys
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			var book Book
			key := "K" + strconv.Itoa(base+i)
			if _, err := col.ReadDocument(nil, key, &book); err != nil {
				log.Fatalf("Failed to read document: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("read document ops", times)
}

func doReadSameDocs(col driver.Collection) {
	// Read always the same document
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			var book Book
			key := "K" + strconv.Itoa(base)
			if _, err := col.ReadDocument(nil, key, &book); err != nil {
				log.Fatalf("Failed to read document: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("read same document ops", times)
}

func doReplaceDocs(col driver.Collection) {
	// Will replace the seeded documents.
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			key := "K" + strconv.Itoa(base)
			book := Book{
				Key:     "K" + strconv.Itoa(base+i),
				Title:   "Some small string",
				NoPages: i,
			}
			if _, err := col.ReplaceDocument(nil, key, &book); err != nil {
				log.Fatalf("Failed to replace document: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("replace same document ops", times)
}

func doVersion(client driver.Client) {
	// Create documents with specific keys
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			// _, err := client.Version(driver.WithDetails(nil, true))
			_, err := client.Version(driver.WithDetails(nil, false))
			if err != nil {
				log.Fatalf("Error in /_api/version call: %v", err)
			}
			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("/_api/version", times)
}

func doInitThreeDiamondAQL(client driver.Client) (driver.Database, driver.Collection) {
	log.Printf("Setting up a database, collection and 100 documents...")
	// Prepare a new books collection in some database:
	db, err := client.Database(nil, "booksDB")
	if err != nil {
		// Create a database
		db, err = client.CreateDatabase(nil, "booksDB", nil)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
	}

	// Create collection
	col, err := db.Collection(nil, "books")
	if err != nil {
		col, err = db.CreateCollection(nil, "books", nil)
		if err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}
	} else {
		col.Truncate(nil)
	}

	// Write some books:
	for i := 0; i < 100; i++ {
		book := Book{
			Key:     "K" + strconv.Itoa(i),
			Title:   "Some small string",
			NoPages: i,
		}
		_, err := col.CreateDocument(nil, book)
		if err != nil {
			log.Fatalf("Failed to create document: %v", err)
		}
	}

	log.Printf("Done, let the race begin!")
	return db, col
}

func doReadThreeDiamondAQL(db driver.Database, col driver.Collection) {
	// Does a lot of three diamond AQL queries
	// Make nrRequests divisible by parallelism:
	nrRequestsPerWorker := nrRequests / parallelism
	nrRequests = nrRequestsPerWorker * parallelism
	submittedRequests += nrRequests
	times := make([]time.Duration, nrRequests, nrRequests)
	wg := sync.WaitGroup{}

	worker := func(innerTimes []time.Duration, base int, initDelay time.Duration) {
		time.Sleep(initDelay)
		for i := 0; i < len(innerTimes); i++ {
			startTime := time.Now()
			var book Book

			// Get books by using AQL
			cur, err := db.Query(nil, "FOR b1 IN books FOR b2 IN books FILTER b1._key == b2._key FOR b3 IN books FILTER b3._key == b1._key LIMIT 10 RETURN {_key: b1._key, title: b2.title, no_pages: b3.no_pages}", nil)
			if err != nil {
				log.Fatalf("Failed to ask for query cursor: %v", err)
			}
			for {
				_, err = cur.ReadDocument(nil, &book)
				if err != nil {
					if driver.IsNoMoreDocuments(err) {
						break
					}
					log.Fatalf("Failed to read doc from cursor: %v", err)
				}
			}

			endTime := time.Now()
			innerTimes[i] = endTime.Sub(startTime)
			time.Sleep(delay)
		}
	}

	for j := 0; j < parallelism; j++ {
		wg.Add(1)
		go func(jj int) {
			defer wg.Done()
			initTime := time.Duration(jj * int(delay) / parallelism)
			// Give non-overlapping slices to the workers which together cover
			// the whole of times:
			worker(times[jj*nrRequestsPerWorker:(jj+1)*nrRequestsPerWorker],
				jj*nrRequestsPerWorker, initTime)
		}(j)
	}

	wg.Wait()
	logStats("read three diamond AQL ops", times)
	if cleanup {
		err := col.Remove(nil)
		if err != nil {
			log.Fatalf("Failed to drop collection: %v", err)
		}
		err = db.Remove(nil)
		if err != nil {
			log.Fatalf("Failed to drop database: %v", err)
		}
	}
}

func main() {
	flag.IntVar(&nrConnections, "nrConnections", nrConnections, "number of connections")
	flag.StringVar(&endpoint, "endpoint", endpoint, "server endpoint")
	flag.StringVar(&testcase, "testcase", testcase, "test case")
	flag.IntVar(&replFactor, "replicationFactor", replFactor, "replication factor of collection")
	flag.IntVar(&nrRequests, "nrRequests", nrRequests, "number of requests")
	flag.IntVar(&parallelism, "parallelism", parallelism, "parallelism")
	flag.DurationVar(&delay, "delay", delay, "delay per thread between operations")
	flag.BoolVar(&cleanup, "cleanup", cleanup, "flag whether to perform cleanup")
	flag.StringVar(&protocol, "protocol", protocol, "protocol: HTTP or VST or HTTP2")
	flag.BoolVar(&usetls, "useTLS", usetls, "flag whether to use TLS")
	flag.StringVar(&username, "auth.user", username, "Authentication Username")
	flag.StringVar(&password, "auth.pass", password, "Authentication Password")
	flag.StringVar(&outputFormat, "outputFormat", outputFormat, "output format: console or csv")
	flag.Parse()

	if outputFormat != "console" && outputFormat != "csv" {
		log.Fatalf("-outputFormat needs to be console or csv")
	}

	// If we log to CSV we suppress Logger output and use fmt to print.
	if outputFormat == "csv" {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("Server endpoint: %s using %d connections", endpoint, nrConnections)
	log.Println()

	var conn driver.Connection
	var err error
	if protocol == "HTTP" {
		connConfig := http.ConnectionConfig{
			Endpoints: []string{endpoint},
			ContentType: driver.ContentTypeVelocypack,
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
	} else if protocol == "HTTP2" {
		var connConfig http.ConnectionConfig
		if usetls {
			connConfig = http.ConnectionConfig{
				Endpoints: []string{endpoint},
				ContentType: driver.ContentTypeVelocypack,
				ConnLimit: nrConnections,
				Transport: &http2.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		} else {
			connConfig = http.ConnectionConfig{
				Endpoints: []string{endpoint},
				ContentType: driver.ContentTypeVelocypack,
				ConnLimit: nrConnections,
				Transport: &http2.Transport{
					AllowHTTP:       true,
					DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
						return net.Dial(network, addr)
					},
				},
			}
		}
		conn, err = http.NewConnection(connConfig)
		if err != nil {
			log.Fatalf("Failed to create HTTP2 connection: %v", err)
		}
	} else {
		log.Fatalf("-protocol needs to be HTTP or VST or HTTP2")
	}

	clientConfig := driver.ClientConfig{
		Connection: conn,
	}

	if username != "" {
		clientConfig.Authentication = driver.BasicAuthentication(username, password)
	}

	c, err := driver.NewClient(clientConfig)

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
		opts := driver.CreateCollectionOptions{
			ReplicationFactor: replFactor,
		}
		col, err = db.CreateCollection(nil, "test", &opts)
		if err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}
	}

	startTime := time.Now()
	switch testcase {
	case "postDocs":
		doPostDocs(col)
	case "seedDocs":
		doSeedDocs(col)
	case "readDocs":
		doReadDocs(col)
	case "readSameDocs":
		doReadSameDocs(col)
	case "replaceDocs":
		doReplaceDocs(col)
	case "readThreeDiamondAQL":
		AQLdb, AQLcol := doInitThreeDiamondAQL(c)
		startTime = time.Now()
		doReadThreeDiamondAQL(AQLdb, AQLcol)
	case "all":
		AQLdb, AQLcol := doInitThreeDiamondAQL(c)
		startTime = time.Now()
		doPostDocs(col)
		doSeedDocs(col)
		doReadDocs(col)
		doReadSameDocs(col)
		doReplaceDocs(col)
		doReadThreeDiamondAQL(AQLdb, AQLcol)
	case "version":
		doVersion(c)
	}
	endTime := time.Now()

	log.Println()
	log.Printf("Time for %d requests: %v", submittedRequests, endTime.Sub(startTime))
	log.Printf("Reqs/s: %d", int(float64(submittedRequests)/(float64(endTime.Sub(startTime))/1000000000.0)))

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
