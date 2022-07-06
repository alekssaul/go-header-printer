package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lightstep/tracecontext.go/traceparent"
)

func main() {
	downstreamURL := os.Getenv("downstreamURL")
	mutateheader := false

	if os.Getenv("mutateheader") != "" {
		mutateheader = true
	}

	log.Printf("Downstream URL: %s \n", downstreamURL)
	log.Printf("Downstream enabled: %v\n", mutateheader)

	// http handler for healthz endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r, mutateheader, downstreamURL)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HttpHandler(w http.ResponseWriter, r *http.Request, mutateheader bool, downstreamURL string) {
	reqHeadersBytes, err := json.Marshal(r.Header)
	if err != nil {
		log.Println("Could not Marshal Req Headers")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tpHeader := r.Header.Get("traceparent")
	if tpHeader != "" && mutateheader == true {
		log.Println("Setting `traceparent` header to a random value ")
		tpHeader = mutateParentId(tpHeader)
		w.Header().Set("traceparent", tpHeader)

	}

	if downstreamURL != "" {
		log.Printf("Calling Downstream URL: %s", downstreamURL)
		err = callDownstream(downstreamURL, tpHeader)
	}

	// Set Response Code
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}

	w.Write(reqHeadersBytes)

}

// mutateParentId: takes traceparent string and returns a mutated version where parentid is replaced randomly
func mutateParentId(before string) (after string) {
	t, err := traceparent.ParseString(before)
	if err != nil {
		log.Printf("Could not parse the ParentID header %s", before)
		return before
	}

	// Generate new SpanID
	r := make([]byte, 8)
	rand.Read(r)
	copy(t.SpanID[:], r[:8])

	return t.String()
}

func callDownstream(downstreamURL string, traceparent string) (e error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", downstreamURL, nil)
	req.Header.Set("traceparent", downstreamURL)
	res, _ := client.Do(req)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Downstream Responded with Error code : %v", res.StatusCode)
	}
	return nil
}
