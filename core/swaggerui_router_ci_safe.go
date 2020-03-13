// +build codeanalysis

package core

import (
	"net/http"

	"github.com/go-chi/chi"
)

func RouteSwaggerUI(swaggerFile http.FileSystem, r chi.Router) {}
