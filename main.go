package main

//go:generate go run main.go openapi openapi.yaml

import (
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/spf13/cobra"
)

func main() {
	var api huma.API

	cli := humacli.New(func(hooks humacli.Hooks, options *struct{}) {
		router := http.NewServeMux()
		api = humago.New(router, huma.DefaultConfig("PyPI Server", "0.0.0.dev0"))

		hooks.OnStart(func() {
			http.ListenAndServe(fmt.Sprintf(":%d", 3000), router)
		})
	})

	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			outputFile := "openapi.yaml"
			if len(args) > 0 {
				outputFile = args[0]
			}

			b, _ := api.OpenAPI().YAML()
			fmt.Printf("Writing OpenAPI spec to %s\n", outputFile)
			if err := os.WriteFile(outputFile, b, 0644); err != nil {
				fmt.Printf("Error writing file: %v\n", err)
			}
		},
	})

	cli.Run()
}
