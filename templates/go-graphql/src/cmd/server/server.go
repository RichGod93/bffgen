package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/graphql/handler"
	"{{PROJECT_NAME}}/graph"
	"{{PROJECT_NAME}}/graph/generated"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	resolver := &graph.Resolver{
		{{ range .BackendServices }}
		{{ .Name | ToPascalCase }}Client: &http.Client{Timeout: 10 * time.Second},
		{{ .Name | ToPascalCase }}URL:    os.Getenv("{{ .Name | ToUpper }}_SERVICE_URL"),
		{{ end }}
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
