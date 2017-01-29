package config

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

func ReadConfig(filename string) (ServerConfig, error) {
	cfg := ServerConfig{}
	return cfg, cfg.Read(filename)
}

type ServerConfig struct {
	Bind []string                `hcl:"bind" json:"bind"`
	Cert []string                `hcl:"cert" json:"cert"`
	GRPC bool                    `hcl:"grpc" json:"grpc"`
	AppM []map[string]*AppConfig `hcl:"app" json:"app"`
	App  []*AppConfig            `hcl:"-" json:"-"`
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

type AppConfig struct {
	server *ServerConfig

	Name   string                    `hcl:"-" json:"-"`
	Host   string                    `hcl:"host" json:"host"`
	Bind   []string                  `hcl:"bind" json:"bind"`
	GRPC   *bool                     `hcl:"grpc" json:"grpc"`
	ProxyM []map[string]*ProxyConfig `hcl:"proxy" json:"proxy"`
	Proxy  []*ProxyConfig            `hcl:"-" json:"-"`
}

func (this *AppConfig) GetBind() []string {
	if len(this.Bind) == 0 {
		return this.server.Bind
	}

	return this.Bind
}

func (this *AppConfig) GetGRPC() bool {
	if this.GRPC == nil {
		return this.server.GRPC
	}

	return *this.GRPC
}

func (this *AppConfig) link() {
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

type ProxyConfig struct {
	app *AppConfig

	Name               string `hcl:"-" json:"-"`
	URI                string `hcl:"uri" json:"uri"`
	Host               string `hcl:"host" json:"host"`
	GRPC               *bool  `hcl:"grpc" json:"grpc"`
	Backend            string `hcl:"backend" json:"backend"`
	Policy             string `hcl:"policy" json:"policy"`
	TLS                bool   `hcl:"tls" json:"tls"`
	InsecureSkipVerify bool   `hcl:"insecure_skip_verify" json:"insecure_skip_verify"`
}

func (this *ProxyConfig) GetGRPC() bool {
	if this.GRPC == nil {
		return this.app.GetGRPC()
	}

	return *this.GRPC
}

func (this *ProxyConfig) link() {

}
