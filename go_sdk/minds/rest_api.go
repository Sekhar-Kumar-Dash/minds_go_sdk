package minds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RestAPI provides methods for interacting with the MindsDB REST API.
type RestAPI struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewRestAPI creates a new RestAPI instance.
func NewRestAPI(apiKey string, baseURL string) *RestAPI {
	if baseURL == "" {
		baseURL = "https://mdb.ai"
	}
	if baseURL[len(baseURL)-1] != '/' {
		baseURL = baseURL + "/"
	}
	if baseURL[len(baseURL)-4:] != "/api/" {
		baseURL = baseURL + "api/"
	}

	return &RestAPI{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (r *RestAPI) _headers() http.Header {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+r.APIKey)
	return headers
}

func (r *RestAPI) get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", r.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = r._headers()
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_raiseForStatus(resp)
	return resp, nil
}

func (r *RestAPI) delete(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", r.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = r._headers()
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_raiseForStatus(resp)
	return resp, nil
}

func (r *RestAPI) post(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", r.BaseURL+url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header = r._headers()
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_raiseForStatus(resp)
	return resp, nil
}

func (r *RestAPI) patch(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", r.BaseURL+url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header = r._headers()
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_raiseForStatus(resp)
	return resp, nil
}

func _raiseForStatus(response *http.Response) error {
	if response.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(response.Body)
		return &ObjectNotFound{Message: string(body)}
	}

	if response.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(response.Body)
		return &Forbidden{Message: string(body)}
	}

	if response.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(response.Body)
		return &Unauthorized{Message: string(body)}
	}

	if response.StatusCode >= 400 && response.StatusCode < 600 {
		body, _ := io.ReadAll(response.Body)
		return &UnknownError{Message: fmt.Sprintf("%s: %s", response.Status, string(body))}
	}
	return nil
}
