package cmd

import (
	"errors"
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

	Coins    []bybit.Coin   `yaml:"coins"`
	Transfer TransferConfig `yaml:"transfer"`

	Log   LogConfig   `yaml:"log"`
	Misc  MiscConfig  `yaml:"misc"`
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

type AccountDescription struct {
	Nickname string `yaml:"nickname"`
	Username string `yaml:"username"`
}

type TransferConfig struct {
	Enabled bool                 `yaml:"enabled"`
	From    []AccountDescription `yaml:"from"`
	To      AccountDescription   `yaml:"to"`

	from []bybit.AccountInfo
	to   bybit.AccountInfo
}

type LogConfig struct {
	Enabled bool     `yaml:"enabled"`
	Format  string   `yaml:"format"` // "text" | "json"
	Output  []string `yaml:"output"` // filepath | "$STDOUT" | "$STDERR"
}

type MiscConfig struct {
	UseColorOutput string `yaml:"use_color_output"` // "auto" | "always" | "never"
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
	if !conf.Transfer.Enabled {
		conf.Transfer.From = nil
	}

	defaultV(&conf.Log.Format, "text")
	defaultV(&conf.Misc.UseColorOutput, "auto")

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
	if !slices.Contains([]string{"auto", "always", "never"}, conf.Misc.UseColorOutput) {
		conf.Misc.UseColorOutput = "auto"
	}
	if conf.Transfer.Enabled && conf.Transfer.To.Username == "" {
		errs = append(errs, fmt.Errorf(`".move.to.username" cannot be empty if ".move.enabled" is true`))
	}
	for _, v := range conf.Transfer.From {
		if v.Username == "" {
			errs = append(errs, fmt.Errorf(`".move.from[].username" cannot be empty`))
		}
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return conf, nil
}

func defaultV[T comparable](target *T, v T) {
	var zero T
	if *target == zero {
		*target = v
	}
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
