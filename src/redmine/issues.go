package redmine

import "fmt"

// IssuesService handles communication with the issue related methods of the
// Redmine API.
//
// Redmine API docs: http://www.redmine.org/projects/redmine/wiki/Rest_Issues
type IssuesService struct {
	client *Client
}

// IssueResult represents a Redmine document with only one issue.
type IssueResult struct {
	Issue Issue `json:"issue"`
}

// IssuesResult represents a Redmine document with more than one issue.
type IssuesResult struct {
	Issues []Issue `json:"issues"`
	*Pagination
}

// Issue represents a Redmine issue.
type Issue struct {
	Number      *int    `json:"id,omitempty"`
	Subject     *string `json:"subject,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (i Issue) String() string {
	return Stringify(i)
}

// IssueRequest represents a request to create/edit an issue.
type IssueRequest struct{}

// IssueListOptions specifies the optional parameters to the IssuesService.List
// and IssuesService.ListByProject methods.
type IssueListOptions struct {
	ListOptions
}

// List fetches all the issues in Redmine
func (s *IssuesService) List(opt *IssueListOptions) ([]Issue, *Response, error) {
	u := "issues"
	return s.listIssues(u, opt)
}

// ListByProject fetches the issues in the specified project for the
// authenticated user.
func (s *IssuesService) ListByProject(project string, opt *IssueListOptions) ([]Issue, *Response, error) {
	u := fmt.Sprintf("projects/%v/issues", project)
	return s.listIssues(u, opt)
}

func (s *IssuesService) listIssues(u string, opt *IssueListOptions) ([]Issue, *Response, error) {
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	issues := new(IssuesResult)
	resp, err := s.client.Do(req, issues)
	if err != nil {
		return nil, resp, err
	}

	return issues.Issues, resp, err
}

// Get a single issue.
func (s *IssuesService) Get(number int) (*Issue, *Response, error) {
	u := fmt.Sprintf("issues/%d", number)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	issue := new(IssueResult)
	resp, err := s.client.Do(req, issue)
	if err != nil {
		return nil, resp, err
	}

	return &issue.Issue, resp, err
}

// Create a new issue.
func (s *IssuesService) Create(issue *IssueRequest) (*Response, error) {
	u := fmt.Sprintf("issues")
	req, err := s.client.NewRequest("POST", u, issue)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

// Edit an issue.
func (s *IssuesService) Edit(number int, issue *IssueRequest) (*Response, error) {
	u := fmt.Sprintf("issues/%d", number)
	req, err := s.client.NewRequest("PATCH", u, issue)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

// Delete an issue.
func (s *IssuesService) Delete(number int) (*Response, error) {
	u := fmt.Sprintf("issues/%d", number)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}
