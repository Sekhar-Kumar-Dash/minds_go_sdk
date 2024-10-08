package minds

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DatabaseConfig represents the configuration for a database data source.
type DatabaseConfig struct {
	Name           string            `json:"name"`
	Engine         string            `json:"engine"`
	Description    string            `json:"description"`
	ConnectionData map[string]string `json:"connection_data"`
	Tables         []string          `json:"tables,omitempty"`
}

// Datasource represents a data source connected to MindsDB.
type Datasource DatabaseConfig

// Datasources manages interactions with MindsDB data sources.
type Datasources struct {
	api *RestAPI
}

// NewDatasources creates a new Datasources instance.
func NewDatasources(client *RestAPI) *Datasources {
	return &Datasources{
		api: client,
	}
}

// Create creates a new data source.
func (d *Datasources) Create(dsConfig *DatabaseConfig, replace bool) (*Datasource, error) {
	if replace {
		// Attempt to retrieve the datasource, if it exists, delete it.
		_, err := d.Get(dsConfig.Name)
		if err == nil { // If no error, the datasource exists.
			err = d.Drop(dsConfig.Name)
			if err != nil {
				return nil, fmt.Errorf("error replacing datasource: %w", err)
			}
		} else if _, ok := err.(*ObjectNotFound); !ok {
			// If the error isn't ObjectNotFound, re-throw it.
			return nil, fmt.Errorf("error checking for existing datasource: %w", err)
		}
		// If the datasource didn't exist, ObjectNotFound is expected. Continue with creation.
	}

	resp, err := d.api.post("/datasources", dsConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating datasource: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Handle non-200 status codes (e.g., return an error)
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return d.Get(dsConfig.Name)
}

// List returns a list of all data sources.
func (d *Datasources) List() ([]*Datasource, error) {
	resp, err := d.api.get("/datasources")
	if err != nil {
		return nil, fmt.Errorf("error listing datasources: %w", err)
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding datasources response: %w", err)
	}

	dsList := []*Datasource{}
	for _, item := range data {
		// Skip non-SQL datasources for now (adjust as needed)
		if _, ok := item["engine"].(string); !ok {
			continue
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("error marshaling datasource item: %w", err)
		}

		var ds Datasource
		if err := json.Unmarshal(jsonData, &ds); err != nil {
			return nil, fmt.Errorf("error unmarshaling datasource item: %w", err)
		}
		dsList = append(dsList, &ds)
	}

	return dsList, nil
}

// Get retrieves a data source by name.
func (d *Datasources) Get(name string) (*Datasource, error) {
	resp, err := d.api.get("/datasources/" + name)
	if err != nil {
		return nil, fmt.Errorf("error getting datasource: %w", err)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding datasource response: %w", err)
	}

	// Skip non-SQL datasources for now (adjust as needed)
	if _, ok := data["engine"].(string); !ok {
		return nil, &ObjectNotSupported{Message: fmt.Sprintf("wrong type of datasource: %s", name)}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling datasource: %w", err)
	}

	var ds Datasource
	if err := json.Unmarshal(jsonData, &ds); err != nil {
		return nil, fmt.Errorf("error unmarshaling datasource: %w", err)
	}

	return &ds, nil
}

// Drop deletes a data source by name.
func (d *Datasources) Drop(name string) error {
	resp, err := d.api.delete("/datasources/" + name)
	if err != nil {
		return fmt.Errorf("error deleting datasource: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// Handle non-200 status codes (e.g., return an error)
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
