package redmine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the Redmine client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

const (
	testBaseURL = "https://www.redmine.org/projects/redmine/"
	testKey     = "Ohk1aiB0ahg6ooz3kai5we6caegheive"
)

// setup sets up a test HTTP server along with a redmine.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// Test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// Redmine client configured to use test server
	url, _ := url.Parse(server.URL)
	client = NewClient(nil, url.String(), testKey)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, testBaseURL, testKey)

	if got, want := c.BaseURL.String(), testBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}

func TestKeyAuthentication(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testHeader(t, r, "X-Redmine-API-Key", testKey)
		fmt.Fprint(w, "{}")
	})

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("client.NewRequest returned error: %v", err)
	}
	_, err = client.Do(req, nil)
	if err != nil {
		t.Errorf("client.Do returned error: %v", err)
	}
}

func TestPagination(t *testing.T) {
	type (
		want struct {
			nextPage  int
			prevPage  int
			firstPage int
			lastPage  int
		}
		test struct {
			body string
			want
		}
	)
	var (
		tests = []test{{
			body: `{ "total_count": 768, "offset": 750, "limit": 25 }`,
			want: want{31, 30, 1, 31},
		}, {
			body: `{ "total_count": 400, "offset": 400, "limit": 400 }`,
			want: want{1, 1, 1, 1},
		}, {
			body: `{ "total_count": 500, "offset": 499, "limit": 1 }`,
			want: want{500, 499, 1, 500},
		}, {
			body: `{ "total_count": 1, "offset": 1, "limit": 1 }`,
			want: want{1, 1, 1, 1},
		}, {
			body: `{ "total_count": 0, "offset": 0, "limit": 0 }`,
			want: want{0, 0, 0, 0},
		}}
		pagination = &Pagination{}
	)

	for _, test := range tests {
		br := bufio.NewReader(strings.NewReader("HTTP/1.1 200 OK\r\n" + "Content-Length: 1234\r\n" + "\r\n" + test.body))
		res, err := http.ReadResponse(br, &http.Request{Method: "GET"})
		if err != nil {
			t.Fatal(err)
		}

		r := newResponse(res)

		err = json.NewDecoder(r.Body).Decode(pagination)
		if err != nil {
			t.Fatal(err)
		}

		pagination.populateValues(r)

		if got, want := r.NextPage, test.want.nextPage; got != want {
			t.Errorf("Response NextPage is %v, want %v (%v)", got, want, test.body)
		}
		if got, want := r.PrevPage, test.want.prevPage; got != want {
			t.Errorf("Response PrevPage is %v, want %v (%v)", got, want, test.body)
		}
		if got, want := r.FirstPage, test.want.firstPage; got != want {
			t.Errorf("Response FirstPage is %v, want %v (%v)", got, want, test.body)
		}
		if got, want := r.LastPage, test.want.lastPage; got != want {
			t.Errorf("Response LastPage is %v, want %v (%v)", got, want, test.body)
		}
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %s, want %s", header, got, want)
	}
}
