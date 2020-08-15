package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/config"
	"github.com/go-chi/chi"
	"gopkg.in/yaml.v2"
)

func Serve(
	ctx context.Context,
	defaultConfig *config.DefaultConfig,
	newRouter func(cfg *config.DefaultConfig) (chi.Router, error),
) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Wrong number of arguments (usage: %s config)", os.Args[0])
	}

	config := os.Args[1]
	configData, err := ioutil.ReadFile(config)
	if err != nil {
		return err
	}
	if err = yaml.UnmarshalStrict(configData, defaultConfig); err != nil {
		return err
	}

	router, err := newRouter(defaultConfig)
	if err != nil {
		return err
	}

	addrConfig := defaultConfig.GenCode.Upstream.HTTP.Common
	serverAddress := fmt.Sprintf("%s:%d", addrConfig.HostName, addrConfig.Port)
	log.Println("Starting Server on " + serverAddress)
	return http.ListenAndServe(serverAddress, router)
}
