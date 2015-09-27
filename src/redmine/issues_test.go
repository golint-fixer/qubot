package redmine

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestIssuesService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
                "issues": [{
                    "id": 1,
                    "subject": "Foobar is badly broken",
                    "description": "According to one of our users, Foobar does not work!"
                }, {
                    "id": 2,
                    "subject": "Foobar is not compatible with humans",
                    "description": "This tool lacks a user interface."
                }]
            }
        }`)
	})

	issue, _, err := client.Issues.List(&IssueListOptions{})
	if err != nil {
		t.Errorf("Issues.List returned error: %v", err)
	}

	want := []Issue{{
		Number:      Int(1),
		Subject:     String("Foobar is badly broken"),
		Description: String("According to one of our users, Foobar does not work!"),
	}, {
		Number:      Int(2),
		Subject:     String("Foobar is not compatible with humans"),
		Description: String("This tool lacks a user interface."),
	}}
	if !reflect.DeepEqual(issue, want) {
		t.Errorf("Issues.Get returned %+v, want %+v", issue, want)
	}
}

func TestIssuesService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/issues/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
                "issue": {
                    "id": 1,
                    "subject": "Foobar is badly broken",
                    "description": "According to one of our users, Foobar does not work!"
                }
            }
        }`)
	})

	issue, _, err := client.Issues.Get(1)
	if err != nil {
		t.Errorf("Issues.Get returned error: %v", err)
	}

	want := &Issue{
		Number:      Int(1),
		Subject:     String("Foobar is badly broken"),
		Description: String("According to one of our users, Foobar does not work!"),
	}
	if !reflect.DeepEqual(issue, want) {
		t.Errorf("Issues.Get returned %+v, want %+v", issue, want)
	}
}
