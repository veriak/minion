package config

import (
	"os"
	"bytes"
	"strings"

	"github.com/spf13/viper"	
	"github.com/go-playground/validator/v10"
	"github.com/fsnotify/fsnotify"	
    	"github.com/rs/zerolog/log"
)

type (
	Config struct {
		Log		Log			`mapstructure:"log" validate:"required"`
		Server		Server			`mapstructure:"server" validate:"required"`
		Minio		Minio			`mapstructure:"minio" validate:"required"`
		Thumbnailer	Thumbnailer		`mapstructure:"thumbnailer" validate:"required"`
		Observability	Observability		`mapstructure:"observability" validate:"required"`		
	}	
	
	Log struct {
		Level		int		`mapstructure:"level"`
	}
	
	Server struct {		
		Addr	string			`mapstructure:"addr" validate:"required"`
		Cert	string			`mapstructure:"cert" validate:"required"`
		Key	string			`mapstructure:"key" validate:"required"`
	}
	
	Minio struct {
		Addr			string			`mapstructure:"addr" validate:"required"`
		AccessKeyID		string			`mapstructure:"accessKeyID" validate:"required"`
		SecretAccessKey	string			`mapstructure:"secretAccessKey" validate:"required"`
		UseSSL			bool			`mapstructure:"useSSL"`
	}

	Thumbnailer struct {
		Width			int			`mapstructure:"width" validate:"required"`
		Height			int			`mapstructure:"height" validate:"required"`
	}
	
	Prometheus struct {
		Enabled	bool		`mapstructure:"enabled"`		
	}
	
	Observability struct {
		Prometheus	Prometheus	`mapstructure:"prometheus" validate:"required"`
	}
)

var (
	config Config
	logger = log.With().Str("service", "Minion").Logger()
)

const (
	portRangeLimit = 100
)	
	
func Get() *Config {
	return &config
}

func (c Config) Validate() error {
	return validator.New().Struct(c)
}

func Load(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}	

	viper.SetConfigFile(file)
	viper.SetConfigType("yaml")	
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")	
	viper.SetEnvPrefix(Namespace)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
		
	if err := viper.ReadConfig(bytes.NewReader([]byte(Default))); err != nil {
		logger.Error().Err(err).Msgf("error loading default configs")
	}
	
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Info().Msgf("Config file changed %s", file)
		reload(e.Name);
	})
	
	return reload(file);
}

func reload(file string) bool {
	err := viper.MergeInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Error().Err(err).Msgf("config file not found %s", file)
    		} else {
	    		logger.Error().Err(err).Msgf("config file read failed %s", file)
		}		
		return false
	}		
	
	err = viper.GetViper().UnmarshalExact(&config)
	if err != nil {
		logger.Error().Err(err).Msgf("config file loaded failed %s", file) 
		return false
	}

	if err = config.Validate(); err != nil {
		logger.Error().Err(err).Msgf("invalid configuration %s", file) 
	}        
	
	logger.Info().Msgf("Config file loaded %s", file)
	return true
}
