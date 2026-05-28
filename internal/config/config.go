package config

import (
	"errors"
	"flag"
	"os"
)

const (
	envDataPath = "RELATION_DATA"
	envAddr     = "RELATION_ADDR"

	defaultAddr = ":8080"
)

type Config struct {
	DataPath string
	Addr     string
}

func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("relation-stub", flag.ContinueOnError)

	dataFlag := fs.String("data", "", "path to seed JSON file (overrides "+envDataPath+")")
	addrFlag := fs.String("addr", "", "listen address, e.g. :8080 (overrides "+envAddr+")")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		DataPath: firstNonEmpty(*dataFlag, os.Getenv(envDataPath)),
		Addr:     firstNonEmpty(*addrFlag, os.Getenv(envAddr), defaultAddr),
	}

	if cfg.DataPath == "" {
		return nil, errors.New("seed data path is required (use --data or " + envDataPath + ")")
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
