package communication

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func Send(url, method string, headers []string, body string) (*http.Response, []byte) {
	fmt.Printf("Sending rq: %#v %#v %#v %#v \n", url, method, headers, body)
	// create request body
	var reqBody *strings.Reader

	if method != "GET" && method != "HEAD" {
		reqBody = strings.NewReader(body)
	}

	// create a request object
	req, _ := http.NewRequest(
		method,
		url,
		reqBody,
	)

	// add a request header
	//req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	// send an HTTP using `req` object
	res, err := http.DefaultClient.Do(req)

	// check for response error
	if err != nil {
		log.Fatal("Error:", err)
	}

	// read response body
	data, _ := ioutil.ReadAll(res.Body)

	// close response body
	defer res.Body.Close()

	return res, data
}
