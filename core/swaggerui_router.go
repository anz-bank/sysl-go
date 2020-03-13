// +build !codeanalysis

package core

import (
	"net/http"
	"strings"

	"github.com/anz-bank/sysl-go-comms/core/build"
	"github.com/go-chi/chi"
	"gopkg.in/russross/blackfriday.v2"
)

func RouteSwaggerUI(swaggerFile http.FileSystem, r chi.Router) {
	r.Route("/-", func(root1 chi.Router) {
		root1.Route("/endpoints", func(root chi.Router) {
			fileServer(root, "/-/endpoints", "/swaggerui", build.React) //nolint:typecheck
			fileServer(root, "/-/endpoints", "/redoc", build.React)     //nolint:typecheck
		})

		root1.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
			if swaggerFile == nil {
				w.Header().Add("Content-Type", "text/html")
				out := []byte(`It seems that you do not have the swagger file generated or generate the embedded static for it.
				To generate the swagger file you need to use the command
				` +
					"\n```shell\n$ sysl export -f \"swagger\" -o gen/model/out.json --app-name app_name app.sysl\n```\n" +
					"To generate the static file, install `https://github.com/omeid/go-resources` and run the command\n ```shell\n$ cd gen/model\n$ resources -var=swagger.file -package=app_name -output=../app_name/swagger.go out.json\n```\n All files should be generated then!")
				if _, err := w.Write(blackfriday.Run(out)); err != nil {
					panic(err)
				}
				return
			}

			out, err := swaggerFile.Open("/out.json")
			if err != nil {
				panic(err)
			}
			statOut, err := out.Stat()
			if err != nil {
				panic(err)
			}
			buf := make([]byte, statOut.Size())
			if _, err = out.Read(buf); err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "application/json")
			if _, err = w.Write(buf); err != nil {
				panic(err)
			}
		})
	})
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem. Adapted from https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func fileServer(r chi.Router, basePath, path string, root http.FileSystem) {
	if root == nil {
		panic("Nothing to serve")
	}

	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(basePath+path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}

	path += "*"
	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := ""

		if r.URL.Query().Get("url") == "" {
			query = "?url=/-/swagger"
		}
		if strings.Contains(r.URL.Path, "endpoints/redoc") && r.URL.Query().Get("mode") != "redoc" {
			query += "&mode=redoc"
		}

		if query != "" {
			http.RedirectHandler(r.URL.Path+query, http.StatusMovedPermanently).ServeHTTP(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	}))
}
