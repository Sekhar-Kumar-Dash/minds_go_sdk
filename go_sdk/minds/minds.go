package minds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	openai "github.com/sashabaranov/go-openai"
)

const DEFAULT_PROMPT_TEMPLATE = "{{input}}"

// Mind represents a MindsDB Mind.
type Mind struct {
	api            *RestAPI
	client         *Client
	Project        string                 `json:"project"`
	Name           string                 `json:"name"`
	ModelName      string                 `json:"model_name"`
	Provider       string                 `json:"provider"`
	PromptTemplate string                 `json:"prompt_template"`
	Parameters     map[string]interface{} `json:"parameters"`
	Datasources    []string               `json:"datasources"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// Update updates a Mind's configuration.
func (m *Mind) Update(updateOpts *UpdateMindOptions) error {
	data := make(map[string]interface{})

	if updateOpts.Datasources != nil {
		dsNames := make([]string, 0, len(updateOpts.Datasources))
		for _, ds := range updateOpts.Datasources {
			dsName, err := m.client.Minds._checkDatasource(ds)
			if err != nil {
				return fmt.Errorf("error checking datasource: %w", err)
			}
			dsNames = append(dsNames, dsName)
		}
		data["datasources"] = dsNames
	}

	if updateOpts.Name != nil {
		data["name"] = *updateOpts.Name
	}
	if updateOpts.ModelName != nil {
		data["model_name"] = *updateOpts.ModelName
	}
	if updateOpts.Provider != nil {
		data["provider"] = *updateOpts.Provider
	}

	parameters := m.Parameters
	if updateOpts.Parameters != nil {
		parameters = updateOpts.Parameters
	}
	if updateOpts.PromptTemplate != nil {
		parameters["prompt_template"] = *updateOpts.PromptTemplate
	}
	data["parameters"] = parameters

	resp, err := m.api.patch(fmt.Sprintf("/projects/%s/minds/%s", m.Project, m.Name), data)
	if err != nil {
		return fmt.Errorf("error updating mind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if updateOpts.Name != nil && *updateOpts.Name != m.Name {
		m.Name = *updateOpts.Name
	}
	return nil
}

type UpdateMindOptions struct {
	Name           *string
	ModelName      *string
	Provider       *string
	PromptTemplate *string
	Datasources    []interface{}
	Parameters     map[string]interface{}
}

func (m *Mind) AddDatasource(datasource interface{}) error {
	dsName, err := m.client.Minds._checkDatasource(datasource)
	if err != nil {
		return fmt.Errorf("error checking datasource: %w", err)
	}

	resp, err := m.api.post(
		fmt.Sprintf("/projects/%s/minds/%s/datasources", m.Project, m.Name),
		map[string]string{"name": dsName},
	)
	if err != nil {
		return fmt.Errorf("error adding datasource to mind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	updatedMind, err := m.client.Minds.Get(m.Name)
	if err != nil {
		return fmt.Errorf("error getting updated mind: %w", err)
	}
	m.Datasources = updatedMind.Datasources
	return nil
}

func (m *Mind) DelDatasource(datasource interface{}) error {
	var dsName string
	switch ds := datasource.(type) {
	case string:
		dsName = ds
	case *Datasource:
		dsName = ds.Name
	default:
		return fmt.Errorf("unknown datasource type: %T", datasource)
	}

	resp, err := m.api.delete(
		fmt.Sprintf("/projects/%s/minds/%s/datasources/%s", m.Project, m.Name, dsName),
	)
	if err != nil {
		return fmt.Errorf("error deleting datasource from mind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	updatedMind, err := m.client.Minds.Get(m.Name)
	if err != nil {
		return fmt.Errorf("error getting updated mind: %w", err)
	}
	m.Datasources = updatedMind.Datasources
	return nil
}

func (m *Mind) Completion(message string, useStream bool) (string, error) {
	parsedURL, err := url.Parse(m.api.BaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing API base URL: %w", err)
	}

	var llmHost string
	if parsedURL.Host == "mdb.ai" {
		llmHost = "llm.mdb.ai"
	} else {
		llmHost = "ai." + parsedURL.Host
	}

	parsedURL.Host = llmHost
	parsedURL.Path = ""

	clientConfig := openai.DefaultConfig(m.api.APIKey)
	clientConfig.BaseURL = parsedURL.String()
	openAIClient := openai.NewClientWithConfig(clientConfig)

	ctx := context.Background()

	if useStream {
		stream, err := openAIClient.CreateChatCompletionStream(
			ctx,
			openai.ChatCompletionRequest{
				Model:    m.Name,
				Messages: []openai.ChatCompletionMessage{{Role: "user", Content: message}},
				Stream:   true,
			},
		)
		if err != nil {
			return "", fmt.Errorf("error creating chat completion stream: %w", err)
		}
		defer stream.Close()

		var fullResponse string
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", fmt.Errorf("error receiving chat completion stream: %w", err)
			}
			fullResponse += response.Choices[0].Delta.Content
		}
		return fullResponse, nil
	} else {
		response, err := openAIClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    m.Name,
				Messages: []openai.ChatCompletionMessage{{Role: "user", Content: message}},
				Stream:   false,
			},
		)
		if err != nil {
			return "", fmt.Errorf("error creating chat completion: %w", err)
		}
		return response.Choices[0].Message.Content, nil
	}
}

type Minds struct {
	api     *RestAPI
	client  *Client
	project string
}

func NewMinds(client *Client) *Minds {
	return &Minds{
		api:     client.api,
		client:  client,
		project: "mindsdb",
	}
}

func (ms *Minds) List() ([]*Mind, error) {
	resp, err := ms.api.get(fmt.Sprintf("/projects/%s/minds", ms.project))
	if err != nil {
		return nil, fmt.Errorf("error listing minds: %w", err)
	}
	defer resp.Body.Close()

	var data []*Mind
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding minds response: %w", err)
	}
	for _, mind := range data {
		mind.client = ms.client
	}
	return data, nil
}

func (ms *Minds) Get(name string) (*Mind, error) {
	resp, err := ms.api.get(fmt.Sprintf("/projects/%s/minds/%s", ms.project, name))
	if err != nil {
		return nil, fmt.Errorf("error getting mind: %w", err)
	}
	defer resp.Body.Close()

	var mind Mind
	if err := json.NewDecoder(resp.Body).Decode(&mind); err != nil {
		return nil, fmt.Errorf("error decoding mind response: %w", err)
	}

	mind.client = ms.client
	return &mind, nil
}

func (ms *Minds) _checkDatasource(ds interface{}) (string, error) {
	switch ds := ds.(type) {
	case string:
		return ds, nil
	case *Datasource:
		return ds.Name, nil
	case DatabaseConfig:
		if _, err := ms.client.Datasources.Get(ds.Name); err != nil {
			if _, ok := err.(*ObjectNotFound); ok {
				if _, err := ms.client.Datasources.Create(&ds, false); err != nil {
					return "", fmt.Errorf("error creating datasource: %w", err)
				}
			} else {
				return "", fmt.Errorf("error checking for existing datasource: %w", err)
			}
		}
		return ds.Name, nil
	default:
		return "", fmt.Errorf("unknown datasource type: %T", ds)
	}
}

type CreateMindOptions struct {
	ModelName      *string                `json:"model_name,omitempty"`
	Provider       *string                `json:"provider,omitempty"`
	PromptTemplate *string                `json:"prompt_template,omitempty"`
	Datasources    []interface{}          `json:"datasources,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
}

func (ms *Minds) Create(name string, opts *CreateMindOptions, replace bool) (*Mind, error) {
	if replace {
		_, err := ms.Get(name)
		if err == nil {
			err = ms.Drop(name)
			if err != nil {
				return nil, fmt.Errorf("error replacing mind: %w", err)
			}
		} else if _, ok := err.(*ObjectNotFound); !ok {
			return nil, fmt.Errorf("error checking for existing mind: %w", err)
		}
	}

	data := map[string]interface{}{
		"name": name,
	}

	if opts != nil {
		if opts.ModelName != nil {
			data["model_name"] = *opts.ModelName
		}
		if opts.Provider != nil {
			data["provider"] = *opts.Provider
		}

		if opts.Datasources != nil {
			dsNames := make([]string, 0, len(opts.Datasources))
			for _, ds := range opts.Datasources {
				dsName, err := ms._checkDatasource(ds)
				if err != nil {
					return nil, fmt.Errorf("error checking datasource: %w", err)
				}
				dsNames = append(dsNames, dsName)
			}
			data["datasources"] = dsNames
		}

		parameters := make(map[string]interface{})
		if opts.Parameters != nil {
			parameters = opts.Parameters
		}
		if opts.PromptTemplate != nil {
			parameters["prompt_template"] = *opts.PromptTemplate
		} else {
			parameters["prompt_template"] = DEFAULT_PROMPT_TEMPLATE
		}
		data["parameters"] = parameters
	} else {
		data["parameters"] = map[string]interface{}{
			"prompt_template": DEFAULT_PROMPT_TEMPLATE,
		}
	}

	resp, err := ms.api.post(fmt.Sprintf("/projects/%s/minds", ms.project), data)
	if err != nil {
		return nil, fmt.Errorf("error creating mind: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return ms.Get(name)
}

func (ms *Minds) Drop(name string) error {
	resp, err := ms.api.delete(fmt.Sprintf("/projects/%s/minds/%s", ms.project, name))
	if err != nil {
		return fmt.Errorf("error deleting mind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
