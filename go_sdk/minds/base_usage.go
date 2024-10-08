package minds

import (
	"fmt"
	"os"
)

func main() {
	apiKey := os.Getenv("MINDSDB_API_KEY")
	if apiKey == "" {
		panic("MINDSDB_API_KEY environment variable not set")
	}
	// --- Connect ---
	client := minds.NewClient(apiKey, "") // Use default base URL

	// Or use a custom server:
	// baseURL := "https://custom_cloud.mdb.ai/"
	// client := minds.NewClient(apiKey, baseURL)

	// --- Create Datasource ---
	postgresConfig := &minds.DatabaseConfig{
		Name:        "my_datasource",
		Description: "<DESCRIPTION-OF-YOUR-DATA>",
		Engine:      "postgres",
		ConnectionData: map[string]string{
			"user":     "demo_user",
			"password": "demo_password",
			"host":     "samples.mindsdb.com",
			"port":     "5432", // Note: Port is a string
			"database": "demo",
			"schema":   "demo_data",
		},
		Tables: []string{"<TABLE-1>", "<TABLE-2>"},
	}

	// --- Create Mind ---

	// With datasource at the same time:
	mind, err := client.Minds.Create(
		"mind_name",
		&minds.CreateMindOptions{
			Datasources: []interface{}{postgresConfig},
		},
		false, // replace
	)
	if err != nil {
		panic(err)
	}

	// Or separately:
	// datasource, err := client.Datasources.Create(postgresConfig, false)
	// if err != nil {
	// 	panic(err)
	// }
	// mind, err := client.Minds.Create(
	// 	"mind_name",
	// 	&minds.CreateMindOptions{
	// 		Datasources: []interface{}{datasource},
	// 	},
	// 	false, // replace
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// With a prompt template:
	// mind, err := client.Minds.Create(
	// 	"mind_name",
	// 	&minds.CreateMindOptions{
	// 		PromptTemplate: minds.StringPtr("You are a coding assistant"),
	// 	},
	// 	false, // replace
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// Or add to an existing mind:
	// mind, err := client.Minds.Get("mind_name")
	// if err != nil {
	// 	panic(err)
	// }
	// // By config:
	// if err := mind.AddDatasource(postgresConfig); err != nil {
	// 	panic(err)
	// }
	// // Or by datasource object:
	// if err := mind.AddDatasource(datasource); err != nil {
	// 	panic(err)
	// }

	// --- Managing Minds ---

	// Create or replace:
	// mind, err = client.Minds.Create(
	// 	"mind_name",
	// 	&minds.CreateMindOptions{
	// 		Datasources: []interface{}{postgresConfig},
	// 	},
	// 	true, // replace
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// Update:
	// err = mind.Update(&minds.UpdateMindOptions{
	// 	Name:        minds.StringPtr("mind_name"), // Required
	// 	Datasources: []interface{}{postgresConfig}, // Replaces current datasources
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// List:
	// mindsList, err := client.Minds.List()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Minds:", mindsList)

	// Get by name:
	// mind, err = client.Minds.Get("mind_name")
	// if err != nil {
	// 	panic(err)
	// }

	// Removing datasource:
	// if err := mind.DelDatasource("my_datasource"); err != nil { // Or pass the datasource object
	// 	panic(err)
	// }

	// Remove mind:
	// if err := client.Minds.Drop("mind_name"); err != nil {
	// 	panic(err)
	// }

	// --- Call Completion ---
	completion, err := mind.Completion("2+3", false) // Non-stream mode
	if err != nil {
		panic(err)
	}
	fmt.Println(completion)

	// Stream completion:
	// completionStream, err := mind.Completion("2+3", true)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(completionStream)

	// --- Managing Datasources ---

	// Create or replace:
	// datasource, err = client.Datasources.Create(postgresConfig, true)
	// if err != nil {
	// 	panic(err)
	// }

	// List:
	// datasources, err := client.Datasources.List()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Datasources:", datasources)

	// Get:
	// datasource, err = client.Datasources.Get("my_datasource")
	// if err != nil {
	// 	panic(err)
	// }

	// Remove:
	// if err := client.Datasources.Drop("my_datasource"); err != nil {
	// 	panic(err)
	// }
}
