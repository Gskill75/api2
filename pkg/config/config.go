package config

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`

	DB struct {
		Host     string `mapstructure:"host"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Port     string `mapstructure:"port"`
		DbName   string `mapstructure:"dbname"`
		SslMode  string `mapstructure:"sslmode"`
	} `mapstructure:"db"`

	Kubernetes struct {
		Url      string `mapstructure:"url"`
		Token    string `mapstructure:"token"`
		Insecure bool   `mapstructure:"insecure"`
		QPS      int    `mapstructure:"qps"`
		Burst    int    `mapstructure:"burst"`
		Timeout  int    `mapstructure:"timeout"`
	} `mapstructure:"kubernetes"`

	Harbor struct {
		Url      string `mapstructure:"url"`
		Token    string `mapstructure:"token"`
		Username string `mapstructure:"username"`
		Insecure bool   `mapstructure:"insecure"`
	} `mapstructure:"harbor"`

	Awx struct {
		Url      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Insecure bool   `mapstructure:"insecure"`
		Token    string `mapstructure:"token"`
		Bearer   string `mapstructure:"bearer"`
	} `mapstructure:"awx"`

	OIDC struct {
		Issuer   string `mapstructure:"issuer"`
		Audience string `mapstructure:"audience"`
	} `mapstructure:"oidc"`

	AWX struct {
		Url      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Insecure bool   `mapstructure:"insecure"`
	} `mapstructure:"awx"`
}

func Load(cmd *cobra.Command) (*Config, error) {
	var config Config
	var cfgFile string
	if cmd != nil {
		cfgFile, _ = cmd.Flags().GetString("config")
	}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc")
	}
	// Set up environment variable bindings
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Define default values
	viper.SetDefault("server.port", ":8080")

	if err := viper.ReadInConfig(); err != nil {
		klog.Warningf("No config file found, using defaults and environment: %v", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		klog.Fatalf("Failed to parse config: %v\n", err)
		os.Exit(1)
	}
	return &config, nil
}
