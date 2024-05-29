package main

import (
	"os"
	"qbart/pgvector/web"

	_ "embed"
)

var (
	//go:embed assets/images/logo.svg
	logo []byte

	//go:embed ui/style.css
	style []byte
)

func main() {
	srv := &web.Server{
		EmbeddableResources: web.EmbeddableResources{
			Favicon: logo,
			Logo:    logo,
			Style:   style,
		},
	}
	srv.Run(os.Args)
}
