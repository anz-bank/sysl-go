package core

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

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
		return fmt.Errorf("Wrong number of arguments (usage: %s (config | -h | --help))", os.Args[0])
	}

	customConfig := CreateConfig(downstreamConfig, createService, serviceInterface)
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Printf("Usage: %s config\n", os.Args[0])
		describeCustomConfig(os.Stdout, customConfig)
		fmt.Println("")
		return nil
	}

	configPath := os.Args[1]
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

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

func describeCustomConfig(w io.Writer, customConfig interface{}) {
	describeYAMLForType(w, reflect.TypeOf(customConfig), 0)
}

func describeYAMLForType(w io.Writer, t reflect.Type, indent int) {
	outf := func(format string, args ...interface{}) {
		parts := strings.SplitAfterN(format, "\n", 2)
		fmt.Fprintf(w, strings.Join(parts, strings.Repeat(" ", indent)), args...)
	}
	switch t.Kind() {
	case reflect.Bool:
		outf(" bool")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		outf(" int")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		outf(" uint")
	case reflect.Float32, reflect.Float64:
		outf(" float")
	case reflect.Array, reflect.Slice:
		outf("\n  - ")
		describeYAMLForType(w, t.Elem(), indent+2)
	case reflect.Interface:
		outf(" any")
	// case reflect.Map:
	case reflect.Ptr:
		describeYAMLForType(w, t.Elem(), indent)
	case reflect.String:
		outf(" string")
	case reflect.Struct:
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			outf("\n%s:", f.Name)
			describeYAMLForType(w, f.Type, indent+2)
		}
	default:
		panic(fmt.Errorf("describeYAMLForType: Unhandled type: %v", t))
	}
}
