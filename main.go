package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ssrDocument(pageHtml []byte) (string, string, error) {
	renderedPage := ""
	renderedHead := ""

	z := html.NewTokenizer(bytes.NewReader(pageHtml))

	var currentElement atom.Atom
	for {
		z.Next()

		token := z.Token()
		if token.Type == html.ErrorToken {
			if z.Err() == io.EOF {
				return renderedPage, renderedHead, nil
			} else {
				return renderedPage, renderedHead, z.Err()
			}
		}

		if token.Type == html.StartTagToken {
			if token.DataAtom == atom.Head || token.DataAtom == atom.Template {
				currentElement = token.DataAtom
				continue
			}
		}

		if currentElement == atom.Head {
			renderedHead += token.String()
		} else if currentElement == atom.Template {
			renderedPage += token.String()
		}

		if token.Type == html.EndTagToken {
			if currentElement == atom.Head || currentElement == atom.Template {
				currentElement = 0
			}
		}
	}
}

func ssrComponent(text []byte) (string, error) {
	return "", nil
}

func main() {
	app := &cli.App{
		Name:  "epic",
		Usage: "Make a website epically",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Epically serve your site",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Usage:       "port to listen on",
						Value:       4000,
						DefaultText: "4000",
					},
				},
				Action: func(cCtx *cli.Context) error {
					port := cCtx.Int("port")

					http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path == "/" {
							r.URL.Path = "/index.html"
						}
						url := path.Clean(r.URL.Path)
						if r.URL.Path[len(r.URL.Path)-1] == '/' {
							url += "/" // path.Clean() may remove trailing slash
						}
						maybeUrl := url
						if url[len(url)-1] != '/' && !strings.HasSuffix(url, ".html") {
							maybeUrl = url + ".html"
						}

						file := path.Join("./site/pages", url)
						maybeFile := path.Join("./site/pages", maybeUrl)

						// If no 'page/*.html' could be found, serve from static directory
						_, err := os.Stat(maybeFile)
						if err != nil {
							http.FileServer(http.Dir("./site/static")).ServeHTTP(w, r)
							return
						}

						documentHtml, err := os.ReadFile("./site/document.html")
						if err != nil {
							w.Write([]byte(err.Error()))
							return
						}
						pageHtml, err := os.ReadFile(file)
						if err != nil {
							w.Write([]byte(err.Error()))
							return
						}

						tmpl, err := template.New(file).Parse(string(documentHtml))
						if err != nil {
							w.Write([]byte(err.Error()))
							return
						}

						renderedPage, renderedHead, err := ssrDocument(pageHtml)
						if err != nil {
							w.Write([]byte(err.Error()))
							return
						}

						if err = tmpl.Execute(w, struct {
							Head string
							Page string
						}{
							Head: renderedHead,
							Page: renderedPage,
						}); err != nil {
							fmt.Println(err)
						}
					})

					fmt.Printf("Listening on :%d\n", port)
					return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
				},
			},
			{
				Name:  "generate",
				Usage: "Epically generate your site for static deployment",
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "complete",
				Usage: "complete a task on the list",
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
