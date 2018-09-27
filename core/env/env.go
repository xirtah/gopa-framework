/*
Copyright 2016 Medcl (m AT medcl.net)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/yaml"
	"github.com/xirtah/gopa-framework/core/config"
	log "github.com/xirtah/gopa-framework/core/logger/seelog"
	"github.com/xirtah/gopa-framework/core/util"
)

// Env is environment object of an application
type Env struct {

	// static configs
	SystemConfig *config.SystemConfig

	// dynamic configs
	RuntimeConfig *config.RuntimeConfig

	IsDebug bool
}

// Environment create a new env instance from a config
func Environment(configFile string) *Env {

	env := Env{}
	sysConfig := loadSystemConfig(configFile)
	env.SystemConfig = &sysConfig

	var err error
	env.RuntimeConfig, err = env.loadRuntimeConfig(configFile)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	if env.IsDebug {
		log.Debug(util.ToJson(env, true))
	}

	return &env
}

var moduleConfig map[string]*config.Config
var pluginConfig map[string]*config.Config
var startTime = time.Now().UTC()

var (
	defaultSystemConfig = config.SystemConfig{
		ClusterConfig: config.ClusterConfig{
			Name: "gopa",
		},
		NetworkConfig: config.NetworkConfig{
			Host: "127.0.0.1",
		},
		NodeConfig: config.NodeConfig{
			Name: util.PickRandomName(),
		},
		PathConfig: config.PathConfig{
			Data: "data",
			Log:  "log",
			Cert: "cert",
		},

		LoggingConfig: config.LoggingConfig{
			LogLevel: "info",
		},

		APIBinding:         "127.0.0.1:8001",
		HTTPBinding:        "127.0.0.1:9001",
		ClusterBinding:     "127.0.0.1:13001",
		AllowMultiInstance: true,
		MaxNumOfInstance:   5,
		TLSEnabled:         false,
	}
)

func loadSystemConfig(cfgFile string) config.SystemConfig {
	cfg := defaultSystemConfig
	cfgFilePath := getConfigPath(cfgFile)
	cfg.ConfigFile = cfgFilePath
	if util.IsExist(cfgFilePath) {
		config, err := yaml.NewConfigWithFile(cfgFilePath, ucfg.PathSep("."))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = config.Unpack(&cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	os.MkdirAll(cfg.GetWorkingDir(), 0777)
	os.MkdirAll(cfg.PathConfig.Log, 0777)
	return cfg
}

var (
	defaultRuntimeConfig = config.RuntimeConfig{}
)

func getConfigPath(cfgFile string) string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath + "/" + cfgFile
}

func (env *Env) loadRuntimeConfig(cfgFile string) (*config.RuntimeConfig, error) {

	var configFile string
	//Use system config file definition if it already exists
	if env.SystemConfig != nil && len(env.SystemConfig.ConfigFile) > 0 {
		configFile = env.SystemConfig.ConfigFile
	} else {
		configFile = getConfigPath(cfgFile)
	}

	filename, _ := filepath.Abs(configFile)
	var cfg config.RuntimeConfig

	if util.FileExists(filename) {
		log.Debug("load configFile:", filename)
		cfg, err := config.LoadFile(filename)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		config := defaultRuntimeConfig

		if err := cfg.Unpack(&config); err != nil {
			log.Error(err)
			return nil, err
		}

		pluginConfig = parseModuleConfig(config.Plugins)
		moduleConfig = parseModuleConfig(config.Modules)

	} else {
		log.Debug("no config file was found")

		cfg = defaultRuntimeConfig
	}

	return &cfg, nil
}

func parseModuleConfig(cfgs []*config.Config) map[string]*config.Config {
	result := map[string]*config.Config{}

	for _, cfg := range cfgs {
		log.Trace(getModuleName(cfg), ",", cfg.Enabled(true))
		name := getModuleName(cfg)
		result[name] = cfg
	}

	return result
}

//GetModuleConfig return specify module's config
func GetModuleConfig(name string) *config.Config {
	cfg := moduleConfig[strings.ToLower(name)]
	return cfg
}

//GetPluginConfig return specify plugin's config
func GetPluginConfig(name string) *config.Config {
	cfg := pluginConfig[strings.ToLower(name)]
	return cfg
}

func getModuleName(c *config.Config) string {
	cfgObj := struct {
		Module string `config:"name"`
	}{}

	if c == nil {
		return ""
	}
	if err := c.Unpack(&cfgObj); err != nil {
		return ""
	}

	return cfgObj.Module
}

// EmptyEnv return a empty env instance
func EmptyEnv() *Env {
	system := defaultSystemConfig
	return &Env{SystemConfig: &system, RuntimeConfig: &config.RuntimeConfig{}}
}

//GetStartTime returns the time the application was started
func GetStartTime() time.Time {
	return startTime
}

//GetAppName returns the name of the application
func (env *Env) GetAppName() string {
	//TODO: Rather than using the cluster name consider adding a new field in the config which is the application name
	return env.SystemConfig.ClusterConfig.Name
}
