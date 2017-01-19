package proxy

import (
	"io/ioutil"
	// "sort"

	"github.com/hashicorp/hcl"
)

const (
	Sep      = ","
	Wildcard = "*"
)

func ReadConfig(filename string) (ServerConfig, error) {
	cfg := ServerConfig{}
	return cfg, cfg.Read(filename)
}

type ServerConfig struct {
	Bind []string                `hcl:"bind" json:"bind"`
	Cert []string                `hcl:"cert" json:"cert"`
	GRPC bool                    `hcl:"grpc" json:"grpc"`
	AppM []map[string]*appConfig `hcl:"app" json:"app"`
	App  []*appConfig            `hcl:"-" json:"-"`
}

func (this *ServerConfig) Read(filename string) error {
	in, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := hcl.Unmarshal(in, this); err != nil {
		return err
	}

	this.link()

	return nil
}

func (this *ServerConfig) link() {
	for _, m := range this.AppM {
		for name, app := range m {
			app.link()
			app.server = this
			app.Name = name
			if app.Host == "" {
				app.Host = name
			}

			this.App = append(this.App, app)
		}
	}
}

type appConfig struct {
	server *ServerConfig

	Name   string                    `hcl:"-" json:"-"`
	Host   string                    `hcl:"host" json:"host"`
	Bind   []string                  `hcl:"bind" json:"bind"`
	GRPC   *bool                     `hcl:"grpc" json:"grpc"`
	ProxyM []map[string]*proxyConfig `hcl:"proxy" json:"proxy"`
	Proxy  []*proxyConfig            `hcl:"-" json:"-"`
}

func (this *appConfig) GetBind() []string {
	if len(this.Bind) == 0 {
		return this.server.Bind
	}

	return this.Bind
}

func (this *appConfig) GetGRPC() bool {
	if this.GRPC == nil {
		return this.server.GRPC
	}

	return *this.GRPC
}

func (this *appConfig) link() {
	for _, m := range this.ProxyM {
		for name, proxy := range m {
			proxy.link()
			proxy.app = this
			proxy.Name = name
			if proxy.URI == "" {
				proxy.URI = name
			}

			this.Proxy = append(this.Proxy, proxy)
		}
	}
}

type proxyConfig struct {
	app *appConfig

	Name               string `hcl:"-" json:"-"`
	URI                string `hcl:"uri" json:"uri"`
	Host               string `hcl:"host" json:"host"`
	GRPC               *bool  `hcl:"grpc" json:"grpc"`
	Backend            string `hcl:"backend" json:"backend"`
	Policy             string `hcl:"policy" json:"policy"`
	TLS                bool   `hcl:"tls" json:"tls"`
	InsecureSkipVerify bool   `hcl:"insecure_skip_verify" json:"insecure_skip_verify"`
}

func (this *proxyConfig) GetGRPC() bool {
	if this.GRPC == nil {
		return this.app.GetGRPC()
	}

	return *this.GRPC
}

func (this *proxyConfig) link() {

}
