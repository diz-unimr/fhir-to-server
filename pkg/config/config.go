package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
	"strings"
	"time"
)

type AppConfig struct {
	App   App   `mapstructure:"app"`
	Kafka Kafka `mapstructure:"kafka"`
	Fhir  Fhir  `mapstructure:"fhir"`
}

type App struct {
	Name     string `mapstructure:"name"`
	LogLevel string `mapstructure:"log-level"`
}

type Kafka struct {
	BootstrapServers string   `mapstructure:"bootstrap-servers"`
	InputTopics      []string `mapstructure:"input-topics"`
	SecurityProtocol string   `mapstructure:"security-protocol"`
	Ssl              Ssl      `mapstructure:"ssl"`
}

type Ssl struct {
	CaLocation          string `mapstructure:"ca-location"`
	CertificateLocation string `mapstructure:"certificate-location"`
	KeyLocation         string `mapstructure:"key-location"`
	KeyPassword         string `mapstructure:"key-password"`
}

type Retry struct {
	Count   int `mapstructure:"count"`
	Timeout int `mapstructure:"timeout"`
	Wait    int `mapstructure:"wait"`
	MaxWait int `mapstructure:"max-wait"`
}

type DateConfig struct {
	Value      *time.Time `mapstructure:"value"`
	Comparator string     `mapstructure:"comparator"`
}

type Filter struct {
	Date DateConfig `mapstructure:"date"`
}

type Fhir struct {
	Server Server `mapstructure:"server"`
	Retry  Retry  `mapstructure:"retry"`
	Filter Filter `mapstructure:"filter"`
}

type Server struct {
	BaseUrl string `mapstructure:"base-url"`
	Auth    *Auth  `mapstructure:"auth"`
}

type Auth struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func LoadConfig(path string) (config AppConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	decoderOpts := func(m *mapstructure.DecoderConfig) {
		m.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			StringToTimeHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	err = viper.Unmarshal(&config, decoderOpts)
	return
}

func StringToTimeHookFunc() mapstructure.DecodeHookFunc {
	loc, _ := time.LoadLocation("Europe/Berlin")
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.ParseInLocation("2006-01-02", data.(string), loc)
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}
