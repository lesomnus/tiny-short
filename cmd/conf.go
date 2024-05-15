package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/lesomnus/tiny-short/bybit"
	"github.com/lesomnus/tiny-short/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	path string

	Secret SecretConfig `yaml:"secret"`

	Coins []bybit.Coin `yaml:"coins"`
	Move  MoveConfig   `yaml:"move"`

	Log   LogConfig   `yaml:"log"`
	Debug DebugConfig `yaml:"debug"`
}

type SecretConfig struct {
	Type           bybit.SecretType `yaml:"type"`
	ApiKeyFile     string           `yaml:"api_key_file"`
	PrivateKeyFile string           `yaml:"private_key_file"`

	Store struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
	}
}

type MoveConfig struct {
	Enabled bool     `yaml:"enabled"`
	From    []string `yaml:"from"`
	To      string   `yaml:"to"`

	from []bybit.AccountInfo
	to   bybit.AccountInfo
}

type LogConfig struct {
	Enabled bool     `yaml:"enabled"`
	Format  string   `yaml:"format"` // "text" | "json"
	Output  []string `yaml:"output"` // filepath | "$STDOUT" | "$STDERR"
}

type DebugConfig struct {
	Enabled         bool `yaml:"enabled"`
	SkipTransaction bool `yaml:"skip_transaction"`
	SkipTransfer    bool `yaml:"skip_transfer"`
}

func (c *LogConfig) NewLogger() (*slog.Logger, error) {
	if !c.Enabled {
		return log.Discard, nil
	}

	ws := []io.Writer{}
	for _, o := range c.Output {
		switch o {
		case "$STDOUT":
			ws = append(ws, os.Stdout)
		case "$STDERR":
			ws = append(ws, os.Stderr)
		default:
			f, err := os.OpenFile(o, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return nil, fmt.Errorf("open %s: %w", o, err)
			}
			ws = append(ws, f)
		}
	}

	o := io.MultiWriter(ws...)

	var logger *slog.Logger
	switch c.Format {
	case "text":
		logger = slog.New(slog.NewTextHandler(o, nil))

	case "json":
		logger = slog.New(slog.NewJSONHandler(o, nil))

	default:
		panic("unreachable")
	}

	return logger, nil
}

func ReadConfig(path string) (*Config, error) {
	conf := &Config{path: path}

	data, err := os.ReadFile(conf.path)
	if err != nil {
		return nil, fmt.Errorf("read config at %s: %w", conf.path, err)
	}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal config at %s: %w", conf.path, err)
	}
	if !conf.Move.Enabled {
		conf.Move.From = nil
		conf.Move.To = ""
	}

	conf.Log.Output = removeDuplicate(conf.Log.Output)

	errs := []error{}
	if !fileExists(conf.Secret.ApiKeyFile) {
		errs = append(errs, errors.New(".api_key_file: file not exist"))
	}
	if !fileExists(conf.Secret.PrivateKeyFile) {
		errs = append(errs, errors.New(".private_key_file: file not exist"))
	}
	if !slices.Contains([]string{"text", "json"}, conf.Log.Format) {
		errs = append(errs, fmt.Errorf(`.log.format must be one of "text" or "json": %s`, conf.Log.Format))
	}
	if conf.Move.Enabled && conf.Move.To == "" {
		errs = append(errs, fmt.Errorf(".move.to must be set if .move.enabled is true"))
	}
	if slices.Contains(conf.Move.From, "") {
		errs = append(errs, fmt.Errorf(".move.from cannot contain empty string"))
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return conf, nil
}

func ParseArgs(args []string) (*Config, error) {
	conf := &Config{}

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&conf.path, "conf", ".tiny-short.yaml", "path to a config file")
	flags.Parse(args[1:])

	data, err := os.ReadFile(conf.path)
	if err != nil {
		return nil, fmt.Errorf("read config at %s: %w", conf.path, err)
	}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal config at %s: %w", conf.path, err)
	}
	if !conf.Move.Enabled {
		conf.Move.From = nil
		conf.Move.To = ""
	}

	conf.Log.Output = removeDuplicate(conf.Log.Output)

	errs := []error{}
	if !fileExists(conf.Secret.ApiKeyFile) {
		errs = append(errs, errors.New(".api_key_file: file not exist"))
	}
	if !fileExists(conf.Secret.PrivateKeyFile) {
		errs = append(errs, errors.New(".private_key_file: file not exist"))
	}
	if !slices.Contains([]string{"text", "json"}, conf.Log.Format) {
		errs = append(errs, fmt.Errorf(`.log.format must be one of "text" or "json": %s`, conf.Log.Format))
	}
	if conf.Move.Enabled && conf.Move.To == "" {
		errs = append(errs, fmt.Errorf(".move.to must be set if .move.enabled is true"))
	}
	if slices.Contains(conf.Move.From, "") {
		errs = append(errs, fmt.Errorf(".move.from cannot contain empty string"))
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return conf, nil
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}

	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
