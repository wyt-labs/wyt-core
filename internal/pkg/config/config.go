package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/wyt-labs/wyt-core/pkg/log"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

var (
	RootPath = ""

	Version = ""

	BuildTime = ""

	CommitID = ""
)

const (
	MarketDriverTypeCoincap = "coincap"
	MarketDriverTypeBinance = "binance"
	MarketDriverTypeOkx     = "okx"
)

func DefaultConfig(rootPath string) *Config {
	return &Config{
		RootPath: rootPath,
		App:      App{},
	}
}

func Load() (*Config, error) {
	cfg, err := func() (*Config, error) {
		rootPath, err := LoadRootPathFromEnv()
		if err != nil {
			return nil, err
		}

		cfg := DefaultConfig(rootPath)
		existConfig := IsConfigExist(cfg)
		if existConfig {
			if err := ReadConfig(cfg); err != nil {
				return nil, err
			}
		}
		if cfg.App.AccessDomain == "" {
			cfg.App.AccessDomain = fmt.Sprintf("%s:%d", util.GetLocalIP(), cfg.HTTP.Port)
		}
		return cfg, nil
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}
	return cfg, nil
}

func LoadRootPathFromEnv() (string, error) {
	rootPath := os.Getenv(rootPathEnvName)
	var err error
	if len(rootPath) == 0 {
		rootPath, err = homedir.Expand(rootPathEnvName)
	}
	return rootPath, err
}

func ReadConfig(config *Config) error {
	viper.SetConfigFile(filepath.Join(config.RootPath, cfgFileName))
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WYT_CORE")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	if err := viper.Unmarshal(config, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(";"),
		)),
	); err != nil {
		return err
	}
	// not support viper.Unmarshal
	config.Log.MaxSize = int64(viper.GetSizeInBytes("log.max_size"))

	return nil
}

func InitLogger(ctx context.Context, config *Config) (*logrus.Logger, error) {
	logger, err := log.New(
		ctx,
		config.Log.Level,
		filepath.Join(config.RootPath, logsDirName),
		config.Log.Filename,
		config.Log.MaxSize,
		config.Log.MaxAge.ToDuration(),
		config.Log.RotationTime.ToDuration(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init logger")
	}
	return logger, nil
}

func PrintSystemInfo(rootPath string, writer func(format string, args ...any)) {
	writer("%s version: %s", AppName, Version)
	writer("System version: %s", runtime.GOOS+"/"+runtime.GOARCH)
	writer("Golang version: %s", runtime.Version())
	writer("App build time: %s", BuildTime)
	writer("Git commit id: %s", CommitID)
	if rootPath != "" {
		writer("Config path: %s", rootPath)
	}
}

func WritePid(rootPath string) error {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)
	if err := os.WriteFile(filepath.Join(rootPath, pidFileName), []byte(pidStr), 0755); err != nil {
		return errors.Wrap(err, "failed to write pid file")
	}
	return nil
}

func RemovePID(rootPath string) error {
	return os.Remove(filepath.Join(rootPath, pidFileName))
}

func WriteDebugInfo(rootPath string, debugInfo any) error {
	p := filepath.Join(rootPath, debugFileName)
	_ = os.Remove(p)

	raw, err := json.Marshal(debugInfo)
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, raw, 0755); err != nil {
		return errors.Wrap(err, "failed to write debug info file")
	}
	return nil
}

func IsConfigExist(config *Config) bool {
	return util.FileExist(filepath.Join(config.RootPath, cfgFileName))
}

func WriteConfig(config *Config) error {
	raw, err := MarshalConfig(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(config.RootPath, cfgFileName), []byte(raw), 0755); err != nil {
		return err
	}
	return nil
}

func MarshalConfig(config *Config) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	e := toml.NewEncoder(buf)
	e.SetIndentTables(true)
	e.SetArraysMultiline(true)
	err := e.Encode(config)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
