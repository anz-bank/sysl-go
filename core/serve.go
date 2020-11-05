package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/config/envvar"
	"github.com/anz-bank/sysl-go/health"
	"github.com/spf13/afero"

	"github.com/anz-bank/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type serveContextKey int

const (
	serveConfigFileSystemKey serveContextKey = iota
)

// ConfigFileSystemOnto adds a config filesystem to ctx.
func ConfigFileSystemOnto(ctx context.Context, fs afero.Fs) context.Context {
	return context.WithValue(ctx, serveConfigFileSystemKey, fs)
}

// Serve is deprecated and will be removed once downstream applications cease
// depending upon it. Generated code will no longer call this function.
// This is a shim for compatibility with code generated by sysl-go versions v0.122.0 & earlier.
func Serve(
	ctx context.Context,
	downstreamConfig, createService, serviceInterface interface{},
	newManagers func(ctx context.Context, cfg *config.DefaultConfig, serviceIntf interface{}, hooks *Hooks) (Manager, *GrpcServerManager, error),
) error {
	srv, err := NewServer(ctx, downstreamConfig, createService, serviceInterface, newManagers)
	if err != nil {
		return err
	}
	return srv.Start()
}

// NewServer returns an auto-generated service.
//nolint:funlen
func NewServer(
	ctx context.Context,
	downstreamConfig, createService, serviceInterface interface{},
	newManagers func(ctx context.Context, cfg *config.DefaultConfig, serviceIntf interface{}, hooks *Hooks) (Manager, *GrpcServerManager, error),
) (StoppableServer, error) {
	MustTypeCheckCreateService(createService, serviceInterface)
	customConfig := NewZeroCustomConfig(reflect.TypeOf(downstreamConfig), GetAppConfigType(createService))
	customConfig, err := LoadCustomConfig(ctx, customConfig)
	if err != nil {
		return nil, err
	}
	if customConfig == nil {
		return nil, fmt.Errorf("configuration is empty")
	}

	customConfigValue := reflect.ValueOf(customConfig).Elem()
	library := customConfigValue.FieldByName("Library").Interface().(config.LibraryConfig)
	admin := customConfigValue.FieldByName("Admin").Interface().(*config.AdminConfig)
	genCodeValue := customConfigValue.FieldByName("GenCode")
	development := customConfigValue.FieldByName("Development").Interface().(*config.DevelopmentConfig)
	appConfig := customConfigValue.FieldByName("App")
	upstream := genCodeValue.FieldByName("Upstream").Interface().(config.UpstreamConfig)
	downstream := genCodeValue.FieldByName("Downstream").Interface()

	defaultConfig := &config.DefaultConfig{
		Library:     library,
		Admin:       admin,
		Development: development,
		GenCode: config.GenCodeConfig{
			Upstream:   upstream,
			Downstream: downstream,
		},
	}

	createServiceResult := reflect.ValueOf(createService).Call(
		[]reflect.Value{reflect.ValueOf(ctx), appConfig},
	)
	if err := createServiceResult[2].Interface(); err != nil {
		return nil, err.(error)
	}
	serviceIntf := createServiceResult[0].Interface()
	hooksIntf := createServiceResult[1].Interface()

	server := &autogenServer{
		name: "nameless-autogenerated-app", // TODO source the application name from somewhere
	}

	pkgLoggerConfigs := []log.Config{
		log.SetVerboseMode(true),
	} // TODO expose this so it is configurable.

	var logrusLogger *logrus.Logger = nil // TODO do we need to expose this or can we delete it?

	ctx = InitialiseLogging(ctx, pkgLoggerConfigs, logrusLogger)
	// OK, we have a ctx that contains a logger now!

	manager, grpcManager, err := newManagers(ctx, defaultConfig, serviceIntf, hooksIntf.(*Hooks))
	if err != nil {
		return nil, err
	}

	server.restManager = manager
	server.grpcServerManager = grpcManager

	server.ctx = ctx

	return server, nil
}

// LoadCustomConfig populates the given zero customConfig value with configuration data.
func LoadCustomConfig(ctx context.Context, customConfig interface{}) (interface{}, error) {
	// TODO make this more flexible. It should be possible to resolve a config value
	// without needing to access os.Args or hit any kind of filesystem.
	if len(os.Args) != 2 {
		return nil, fmt.Errorf("Wrong number of arguments (usage: %s (config | -h | --help))", os.Args[0])
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Printf("Usage: %s config\n\n", os.Args[0])
		describeCustomConfig(os.Stdout, customConfig)
		fmt.Print("\n\n")
		return nil, nil
	}

	var fs afero.Fs
	if v := ctx.Value(serveConfigFileSystemKey); v != nil {
		fs = v.(afero.Fs)
	} else {
		fs = afero.NewOsFs()
	}

	configPath := os.Args[1]
	b := envvar.NewConfigReaderBuilder().WithFs(fs).WithConfigFile(configPath)

	// Use the environment variable prefix from the config file if provided
	env, err := b.Build().GetString("config.envPrefix")
	// Disable the feature if none is provided
	if len(env) > 0 && err == nil {
		log.Info(ctx, "config environment variable prefix set: "+env)
		b = b.AttachEnvPrefix(env)
	}

	err = b.Build().Unmarshal(customConfig)
	if err != nil {
		return nil, err
	}
	return customConfig, err
}

// NewZeroCustomConfig uses reflection to create a new type derived from DefaultConfig,
// but with new GenCode.Downstream and App fields holding the same types as
// downstreamConfig and appConfig. It returns a pointer to a zero value of that
// new type.
func NewZeroCustomConfig(downstreamConfigType, appConfigType reflect.Type) interface{} {
	defaultConfigType := reflect.TypeOf(config.DefaultConfig{})

	libraryField, has := defaultConfigType.FieldByName("Library")
	if !has {
		panic("config.DefaultType missing Library field")
	}

	adminField, _ := defaultConfigType.FieldByName("Admin")
	if !has {
		panic("config.DefaultType missing Admin field")
	}

	developmentField, has := defaultConfigType.FieldByName("Development")
	if !has {
		panic("config.DefaultType missing Development field")
	}

	genCodeType := reflect.TypeOf(config.GenCodeConfig{})

	upstreamField, has := genCodeType.FieldByName("Upstream")
	if !has {
		panic("config.DefaultType missing Upstream field")
	}

	return reflect.New(reflect.StructOf([]reflect.StructField{
		libraryField,
		adminField,
		{Name: "GenCode", Type: reflect.StructOf([]reflect.StructField{
			upstreamField,
			{Name: "Downstream", Type: downstreamConfigType, Tag: `mapstructure:"downstream"`},
		}), Tag: `mapstructure:"genCode"`},
		developmentField,
		{Name: "App", Type: appConfigType, Tag: `mapstructure:"app"`},
	})).Interface()
}

// MustTypeCheckCreateService checks that the given createService has an acceptable type, and panics otherwise.
func MustTypeCheckCreateService(createService, serviceInterface interface{}) {
	cs := reflect.TypeOf(createService)
	if cs.NumIn() != 2 {
		panic("createService: wrong number of in params")
	}
	if cs.NumOut() != 3 {
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

	var hooks Hooks
	if reflect.TypeOf(&hooks) != cs.Out(1) {
		panic(fmt.Errorf("createService: second out param must be of type *Hooks, not %v", cs.Out(1)))
	}

	var err error
	if reflect.TypeOf(&err).Elem() != cs.Out(2) {
		panic(fmt.Errorf("createService: third out param must be of type error, not %v", cs.Out(1)))
	}
}

// GetAppConfigType extracts the app's config type from createService.
// Precondition: MustTypeCheckCreateService(createService, serviceInterface) succeeded.
func GetAppConfigType(createService interface{}) reflect.Type {
	cs := reflect.TypeOf(createService)
	return cs.In(1)
}

func yamlEgComment(example, format string, args ...interface{}) string {
	return fmt.Sprintf("\033[1;31m%s \033[0;32m# "+format+"\033[0m", append([]interface{}{example}, args...)...)
}

func describeCustomConfig(w io.Writer, customConfig interface{}) {
	commonTypes := map[reflect.Type]string{
		reflect.TypeOf(config.CommonServerConfig{}):   "",
		reflect.TypeOf(config.CommonDownstreamData{}): "",
		reflect.TypeOf(config.TLSConfig{}):            "",
		reflect.TypeOf(common.SensitiveString{}):      yamlEgComment(`"*****"`, "sensitive string"),
	}

	fmt.Fprint(w, "\033[1mConfiguration file YAML schema\033[0m")

	commonTypeNames := make([]string, 0, len(commonTypes))
	commonTypesByName := make(map[string]reflect.Type, len(commonTypes))
	for ct := range commonTypes {
		name := fmt.Sprintf("%s.%s", ct.PkgPath(), ct.Name())
		commonTypeNames = append(commonTypeNames, name)
		commonTypesByName[name] = ct
	}
	sort.Strings(commonTypeNames)

	for _, name := range commonTypeNames {
		ct := commonTypesByName[name]
		if commonTypes[ct] == "" {
			delete(commonTypes, ct)
			fmt.Fprintf(w, "\n\n\033[1;32m%q.%s:\033[0m", ct.PkgPath(), ct.Name())
			describeYAMLForType(w, ct, commonTypes, 4)
			commonTypes[ct] = ""
		}
	}

	fmt.Fprintf(w, "\n\n\033[1mApplication Configuration\033[0m")
	describeYAMLForType(w, reflect.TypeOf(customConfig), commonTypes, 0)
}

//nolint:funlen
func describeYAMLForType(w io.Writer, t reflect.Type, commonTypes map[reflect.Type]string, indent int) {
	outf := func(format string, args ...interface{}) {
		parts := strings.SplitAfterN(format, "\n", 2)
		fmt.Fprintf(w, strings.Join(parts, strings.Repeat(" ", indent)), args...)
	}
	if alias, has := commonTypes[t]; has {
		if alias == "" {
			outf(" " + yamlEgComment(`{}`, "%q.%s", t.PkgPath(), t.Name()))
		} else {
			outf(" %s", alias)
		}
		return
	}
	switch reflect.New(t).Elem().Interface().(type) { //nolint:gocritic
	case logrus.Level:
		outf(" \033[1m%s\033[0m", logrus.StandardLogger().Level.String())
		return
	}
	switch t.Kind() {
	case reflect.Bool:
		outf(" \033[1mfalse\033[0m")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		outf(" \033[1m0\033[0m")
	case reflect.Float32, reflect.Float64:
		outf(" \033[1m0.0\033[0m")
	case reflect.Array, reflect.Slice:
		outf("\n-")
		describeYAMLForType(w, t.Elem(), commonTypes, indent+4)
	case reflect.Interface:
		outf(" " + yamlEgComment("{}", "any value"))
	case reflect.Map:
		outf("\n key: ")
		describeYAMLForType(w, t.Elem(), commonTypes, indent+4)
	case reflect.Ptr:
		describeYAMLForType(w, t.Elem(), commonTypes, indent)
	case reflect.String:
		outf(" \033[1m\"\"\033[0m")
	case reflect.Struct:
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			yamlTag := f.Tag.Get("yaml")
			yamlParts := strings.Split(yamlTag, ",")
			var name string
			if len(yamlParts) > 0 {
				name = yamlParts[0]
			} else {
				name = f.Name
			}
			outf("\n%s:", name)
			describeYAMLForType(w, f.Type, commonTypes, indent+4)
		}
	default:
		panic(fmt.Errorf("describeYAMLForType: Unhandled type: %v", t))
	}
}

// experimental fork of core.ServerParams

type autogenServer struct {
	ctx                context.Context
	name               string
	restManager        Manager
	grpcServerManager  *GrpcServerManager
	prometheusRegistry *prometheus.Registry
	servers            []StoppableServer
}

//nolint:funlen,gocognit
func (s *autogenServer) Start() error {
	// precondition: ctx must have been threaded through InitialiseLogging and hence contain a logger
	ctx := s.ctx

	// prepare the middleware
	mWare := prepareMiddleware(s.name, s.prometheusRegistry)

	// load health server
	var healthServer *health.Server = nil
	var err error
	if s.restManager != nil && s.restManager.LibraryConfig() != nil && s.restManager.LibraryConfig().Health {
		healthServer, err = health.NewServer()
		if err != nil {
			return err
		}
		s.grpcServerManager.EnabledGrpcHandlers = append(s.grpcServerManager.EnabledGrpcHandlers, healthServer)
	}

	var restIsRunning, grpcIsRunning bool

	servers := make([]StoppableServer, 0)

	// Make the listener function for the REST Admin server
	if s.restManager != nil && s.restManager.AdminServerConfig() != nil {
		log.Info(ctx, "found AdminServerConfig for REST")
		serverAdmin, err := configureAdminServerListener(ctx, s.restManager, s.prometheusRegistry, healthServer.HTTP, mWare.admin)
		if err != nil {
			return err
		}
		servers = append(servers, serverAdmin)
	} else {
		log.Info(ctx, "no AdminServerConfig for REST was found")
	}

	// Make the listener function for the REST Public server
	if s.restManager != nil && s.restManager.PublicServerConfig() != nil {
		log.Info(ctx, "found PublicServerConfig for REST")
		serverPublic, err := configurePublicServerListener(ctx, s.restManager, mWare.public)
		if err != nil {
			return err
		}
		servers = append(servers, serverPublic)
		restIsRunning = true
	} else {
		log.Info(ctx, "no PublicServerConfig for REST was found")
	}

	// Make the listener function for the gRPC Public server.
	if s.grpcServerManager != nil && s.grpcServerManager.GrpcPublicServerConfig != nil && len(s.grpcServerManager.EnabledGrpcHandlers) > 0 {
		log.Info(ctx, "found GrpcPublicServerConfig for gRPC")
		serverPublicGrpc := configurePublicGrpcServerListener(ctx, *s.grpcServerManager)
		servers = append(servers, serverPublicGrpc)
		grpcIsRunning = true
	} else {
		log.Info(ctx, "no GrpcPublicServerConfig for gRPC was found")
	}

	// Refuse to start and panic if neither of the public servers are enabled.
	if !restIsRunning && !grpcIsRunning {
		err := errors.New("REST and gRPC servers cannot both be nil")
		log.Error(ctx, err)
		panic(err)
	}

	s.servers = servers

	// Start all configured servers and block until the first one terminates.
	errChan := make(chan error, 1)
	for i := range servers {
		i := i                 // force capture
		server := s.servers[i] // force capture
		go func() {
			log.Infof(ctx, "starting sub-server %d of %d", i, len(servers))
			errChan <- server.Start()
		}()
	}

	if healthServer != nil {
		// Set health server to be ready
		healthServer.SetReady(true)
	}

	return <-errChan
}

// FIXME replace MultiError with some existing type that does this job better.
type MultiError struct {
	Msg    string
	Errors []error
}

func (e MultiError) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, e := range e.Errors {
		msgs[i] = e.Error()
	}
	return fmt.Sprintf("%s; sub-error(s): %s", e.Msg, strings.Join(msgs, "; "))
}

func (s *autogenServer) Stop() error {
	errQueue := make(chan error, len(s.servers))

	var wg sync.WaitGroup
	for i := range s.servers {
		i := i                 // force capture
		server := s.servers[i] // force capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Infof(s.ctx, "stopping sub-server %d of %d...", i, len(s.servers))
			err := server.Stop()
			log.Infof(s.ctx, "stopped sub-server %d of %d", i, len(s.servers))
			if err != nil {
				errQueue <- err
			}
		}()
	}
	wg.Wait()
	close(errQueue)
	errors := make([]error, 0)
	for err := range errQueue {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return MultiError{Msg: "error during stop", Errors: errors}
	}
	return nil
}

func (s *autogenServer) GracefulStop() error {
	errQueue := make(chan error, len(s.servers))

	var wg sync.WaitGroup
	for i := range s.servers {
		i := i                 // force capture
		server := s.servers[i] // force capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Infof(s.ctx, "graceful-stopping sub-server %d of %d...", i, len(s.servers))
			err := server.GracefulStop()
			log.Infof(s.ctx, "graceful-stopped sub-server %d of %d", i, len(s.servers))
			if err != nil {
				errQueue <- err
			}
		}()
	}
	wg.Wait()
	close(errQueue)
	errors := make([]error, 0)
	for err := range errQueue {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return MultiError{Msg: "error during graceful stop", Errors: errors}
	}
	return nil
}
