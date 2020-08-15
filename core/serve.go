package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/anz-bank/sysl-go/config"
	"github.com/go-chi/chi"
	"gopkg.in/yaml.v2"
)

func Serve(
	ctx context.Context,
	downstreamConfig, createService, serviceInterface interface{},
	newRouter func(cfg *config.DefaultConfig, serviceIntf interface{}) (chi.Router, error),
) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Wrong number of arguments (usage: %s config)", os.Args[0])
	}

	configPath := os.Args[1]
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	customConfig := CreateConfig(downstreamConfig, createService, serviceInterface)
	if err = yaml.UnmarshalStrict(configData, customConfig); err != nil {
		return err
	}

	customConfigValue := reflect.ValueOf(customConfig).Elem()
	library := customConfigValue.FieldByName("Library").Interface().(config.LibraryConfig)
	genCodeValue := customConfigValue.FieldByName("GenCode")
	appConfig := customConfigValue.FieldByName("App")
	upstream := genCodeValue.FieldByName("Upstream").Interface().(config.UpstreamConfig)
	downstream := genCodeValue.FieldByName("Downstream").Interface()

	defaultConfig := &config.DefaultConfig{
		Library: library,
		GenCode: config.GenCodeConfig{
			Upstream:   upstream,
			Downstream: downstream,
		},
	}

	createServiceResult := reflect.ValueOf(createService).Call(
		[]reflect.Value{reflect.ValueOf(ctx), appConfig},
	)
	errIntf := createServiceResult[1].Interface()
	if errIntf != nil {
		return err.(error)
	}
	serviceIntf := createServiceResult[0].Interface()

	router, err := newRouter(defaultConfig, serviceIntf)
	if err != nil {
		return err
	}

	addrConfig := defaultConfig.GenCode.Upstream.HTTP.Common
	serverAddress := fmt.Sprintf("%s:%d", addrConfig.HostName, addrConfig.Port)
	log.Println("Starting Server on " + serverAddress)
	return http.ListenAndServe(serverAddress, router)
}

// CreateConfig uses reflection to create a new type derived from DefaultConfig,
// but with new GenCode.Downstream and App fields holding the same types as
// downstreaConfig and appConfig.
func CreateConfig(downstreamConfig, createService, serviceInterface interface{}) interface{} {
	defaultConfigType := reflect.TypeOf(config.DefaultConfig{})

	libraryField, has := defaultConfigType.FieldByName("Library")
	if !has {
		panic("config.DefaultType missing Library field")
	}

	genCodeType := reflect.TypeOf(config.GenCodeConfig{})

	upstreamField, has := genCodeType.FieldByName("Upstream")
	if !has {
		panic("config.DefaultType missing Upstream field")
	}

	downstreamConfigType := reflect.TypeOf(downstreamConfig)
	appConfigType := GetCreateServiceConfigType(createService, serviceInterface)

	return reflect.New(reflect.StructOf([]reflect.StructField{
		libraryField,
		{Name: "GenCode", Type: reflect.StructOf([]reflect.StructField{
			upstreamField,
			{Name: "Downstream", Type: downstreamConfigType, Tag: `yaml:"downstream"`},
		}), Tag: `yaml:"genCode"`},
		{Name: "App", Type: appConfigType, Tag: `yaml:"app"`},
	})).Interface()
}

func GetCreateServiceConfigType(createService, serviceInterface interface{}) reflect.Type {
	cs := reflect.TypeOf(createService)
	if cs.NumIn() != 2 {
		panic("createService: wrong number of in params")
	}
	if cs.NumOut() != 2 {
		panic("createService: wrong number of out params")
	}

	var ctx context.Context
	if reflect.TypeOf(&ctx).Elem() != cs.In(0) {
		panic(fmt.Errorf("createService: first in param must be of type context.Context, not %v", cs.In(0)))
	}

	serviceInterfaceType := reflect.TypeOf(serviceInterface)
	if serviceInterfaceType != cs.Out(0) {
		panic(fmt.Errorf("createService: second out param must be of type %v, not %v", serviceInterfaceType, cs.Out(0)))
	}

	var err error
	if reflect.TypeOf(&err).Elem() != cs.Out(1) {
		panic(fmt.Errorf("createService: second out param must be of type error, not %v", cs.Out(1)))
	}

	return cs.In(1)
}
