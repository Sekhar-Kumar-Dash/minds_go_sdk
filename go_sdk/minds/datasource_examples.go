package minds

// This code goes in: datasources/examples_test.go

var ExampleDS = &DatabaseConfig{
	Name:        "example_ds",
	Engine:      "postgres",
	Description: "Minds example database",
	ConnectionData: map[string]string{
		"user":     "demo_user",
		"password": "demo_password",
		"host":     "samples.mindsdb.com",
		"port":     "5432",
		"database": "demo",
		"schema":   "demo_data",
	},
}
