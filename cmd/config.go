package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"sync"
)

// ======

type ConfigHTTP struct {
	Listen string `yaml:"listen"`
}

type ConfigHTTPS struct {
	Listen string `yaml:"listen"`
	Key    string `yaml:"key"`
	Pem    string `yaml:"pem"`
}

type ConfigDomainEntry struct {
	Endpoints []string `yaml:"endpoints"`
	Methods   []string `yaml:"methods"`
	Action    string   `yaml:"action"`
}

type ConfigAPI struct {
	Endpoint string `yaml:"endpoint"`
	Prefix   string `yaml:"prefix"`
}

type ConfigProxyEntryTracker struct {
	Enabled     bool   `yaml:"enabled"`
	AllowMirror bool   `yaml:"allowMirror"`
	MirrorId    string `yaml:"mirrorId"`
}

type ConfigProxyAddHeaders struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type ConfigProxyEntry struct {
	Id             string                  `yaml:"id"`
	XForwardedFor  bool                    `yaml:"XForwardedFor"`
	LocalForwarder bool                    `yaml:"localForwarder"`
	Filters        []string                `yaml:"filters"`
	Tracker        ConfigProxyEntryTracker `yaml:"tracker"`
	AddHeaders     []ConfigProxyAddHeaders `yaml:"addHeaders"`
}

type Config struct {
	HTTP    ConfigHTTP          `yaml:"http"`
	HTTPS   ConfigHTTPS         `yaml:"https"`
	Domains []ConfigDomainEntry `yaml:"domains"`
	API     ConfigAPI           `yaml:"api"`
	Proxy   []ConfigProxyEntry  `yaml:"proxy"`

	sync.RWMutex
}

func loadConfig(fn string) (conf Config, err error) {
	f, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(f, &conf)
	return
}

func (c *Config) getDomainEndpoint(domain string, method string) (id string, ok bool) {
	c.RLock()
	defer c.RUnlock()

	for _, x := range c.Domains {
		for _, dn := range x.Endpoints {
			if dn == domain {
				for _, mt := range x.Methods {
					if strings.ToLower(mt) == strings.ToLower(method) {
						return x.Action, true
					}
				}
			}
		}
	}

	return "", false
}

func (c *Config) getProxyAction(id string) (e ConfigProxyEntry, ok bool) {
	c.RLock()
	defer c.RUnlock()
	for _, x := range c.Proxy {
		if x.Id == id {
			return x, true
		}
	}
	return ConfigProxyEntry{}, false
}
