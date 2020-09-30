# jwtauth

This is sysl-go's vendored version of the jwtauth library.

See [docs](./docs) for more details on the role and desigh.

## Usage

### Middleware

* [jwthttp](./jwthttp)

### Customize authorisation

Authorisation refers to the process of checking a request as sufficient
permissions to be executed. There are all sorts of conditions that can affect
whether a request should be authorised, so this module provides an interface
you should implement for custom authorisation

```go
type Authoriser interface {
    Authorise(Claims) error
}

type AuthoriseFunc func(Claims) error
```

### Configure authentication

Configuration can be done via json, yaml and viper, and integrates directly into larger config structs.

```go
// Add Auth to your config
type AppConfig struct {
    Auth jwthttp.Config `json:"auth" yaml:"auth" mapstructure:"auth"`
}

// Initialise auth from config
func main() {
    // load config

    authClient := &http.Client{}
    auth := jwthttp.AuthFromConfig(&appConfig.Auth, func(string) *http.Client {return authClient})
    // This client func can be used to select a client on a per-issuer basis
}

// config.yaml
app:
  auth:
    headers: ["Authorisation"]
    issuers:
      - name: "mountebank"
        jwksUrl: "http://localhost:8888/.well-known/jwks.json"
        cacheTTL: "30m"
```

