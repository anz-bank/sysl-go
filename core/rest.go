package core

type RestResult struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}
