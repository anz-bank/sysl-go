package status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anz-bank/sysl-go/handlerinitialiser"

	"github.com/anz-bank/sysl-go/config"

	"github.com/go-chi/chi"
)

type Service struct {
	BuildMetadata *BuildMetadata
	Config        *config.LibraryConfig
	Services      []handlerinitialiser.HandlerInitialiser
}

func WireRoutes(r chi.Router, s *Service) {
	r.Get("/", s.HandleGetStatus)
}

type BuildMetadata struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	BuildID    string `json:"build_id"`
	CommitSha  string `json:"commit_sha"`
	BranchName string `json:"branch_name"`
	TagName    string `json:"tag_name"`
}

func (m *BuildMetadata) String() string {
	return fmt.Sprintf(
		"%s version=%s build=%s commit=%s branch=%s tag=%s",
		m.Name,
		m.Version,
		m.BuildID,
		m.CommitSha,
		m.BranchName,
		m.TagName,
	)
}

type ServiceStatus struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type ResponseConfig struct {
	Core     *config.LibraryConfig `yaml:"core"`
	Services []ServiceStatus       `yaml:"services"`
}

type Response struct {
	BuildMetadata BuildMetadata  `json:"build_metadata"`
	Config        ResponseConfig `json:"config"`
	Status        string         `json:"status"`
}

func (s *Service) buildResponseConfig() ResponseConfig {
	rcfg := ResponseConfig{
		Core:     s.Config,
		Services: make([]ServiceStatus, 0),
	}

	for _, service := range s.Services {
		rcfg.Services = append(rcfg.Services, ServiceStatus{service.Name(), service.Config()})
	}
	return rcfg
}

func (s *Service) HandleGetStatus(rw http.ResponseWriter, r *http.Request) {
	response := Response{
		BuildMetadata: *s.BuildMetadata,
		Config:        s.buildResponseConfig(),
		Status:        "online",
	}

	buffer := bytes.Buffer{}
	enc := json.NewEncoder(&buffer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		panic(err) // Give up and let chi middleware deal with it
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	_, _ = rw.Write(buffer.Bytes())
	// Ignore write error, if any, as it is probably a client issue.
}
