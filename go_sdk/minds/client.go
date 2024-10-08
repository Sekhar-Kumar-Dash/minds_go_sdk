package minds

// Client is the main entry point for interacting with the MindsDB API.
type Client struct {
	api         *RestAPI
	Datasources *Datasources
	Minds       *Minds
}

// NewClient creates a new MindsDB client.
func NewClient(apiKey string, baseURL string) *Client {
	// Create RestAPI instance with default base URL if not provided.
	api := NewRestAPI(apiKey, baseURL)

	client := &Client{
		api: api,
	}

	// Initialize Datasources and Minds with the client instance.
	client.Datasources = NewDatasources(api)
	client.Minds = NewMinds(client)

	return client
}
