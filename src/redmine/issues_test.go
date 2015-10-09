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

	mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page":  "2",
			"limit": "2",
		})
		fmt.Fprint(w, `{
		"issues": [{
			"id": 3,
			"subject": "Foobar is badly broken",
			"description": "According to one of our users, Foobar does not work!"
		}, {
			"id": 4,
			"subject": "Foobar is not compatible with humans",
			"description": "This tool lacks a user interface."
		}],
		"total_count": 4,
		"offset": 2,
		"limit": 2
            }
        }`)
	})

	issues, resp, err := client.Issues.List(&IssueListOptions{
		ListOptions{Page: 2, PerPage: 2},
	})
	if err != nil {
		t.Errorf("Issues.List returned error: %v", err)
	}

	want := []Issue{{
		Number:      Int(3),
		Subject:     String("Foobar is badly broken"),
		Description: String("According to one of our users, Foobar does not work!"),
	}, {
		Number:      Int(4),
		Subject:     String("Foobar is not compatible with humans"),
		Description: String("This tool lacks a user interface."),
	}}
	if !reflect.DeepEqual(issues, want) {
		t.Errorf("Issues.Get returned %+v, want %+v", issues, want)
	}

	pwant := []int{2, 1, 1, 2}
	pgot := []int{resp.NextPage, resp.PrevPage, resp.FirstPage, resp.LastPage}
	if !reflect.DeepEqual(pgot, pwant) {
		t.Errorf("Issues.List returned %+v, want %+v", pgot, pwant)
	}
}

func TestIssuesService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{ "issue": {
			"id": 1,
			"subject": "Foobar is badly broken",
			"description": "According to one of our users, Foobar does not work!"
		}}`)
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
