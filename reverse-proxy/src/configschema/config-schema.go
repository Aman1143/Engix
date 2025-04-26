package configschema

type WorkerMessageSchema struct {
	RequestType string              `json:"requestType"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	URL         string              `json:"url"` 
	RemoteAddr  string              `json:"remoteAddr"`
}
type WorkerResponse struct {
	WorkerID string `json:"workerID"`
	Status   int    `json:"status"`
	Body     string `json:"body"`
	Error    string `json:"error,omitempty"`
}



type ServerSchema struct {
	Listen    int              `json:"listen" validate:"required"`
	Worker    int              `json:"worker" validate:"required"`
	UpStream  []UpstreamNode   `json:"upstreams" validate:"required,dive"`
	Headers   []HeaderRule     `json:"headers"`
	Rules     []RoutingRule    `json:"rules" validate:"required,dive"`
}

type UpstreamNode struct {
	ID  string `json:"id" validate:"required"`
	URL string `json:"url" validate:"required"`
}

type HeaderRule struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type RoutingRule struct {
	Path      string   `json:"path" validate:"required"`
	Upstreams []string `json:"upstreams" validate:"required"`
}

type RootConfigSchema struct {
	Server ServerSchema `json:"server" validate:"required"`
}

 