package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	defaultWorkMinutes  = 25
	defaultBreakMinutes = 5
	defaultWorkPhases   = 4
)

type Config struct {
	WorkMinutes  int `toml:"work_minutes"`
	BreakMinutes int `toml:"break_minutes"`
	WorkPhases   int `toml:"work_phases"`
}

func Default() Config {
	return Config{
		WorkMinutes:  defaultWorkMinutes,
		BreakMinutes: defaultBreakMinutes,
		WorkPhases:   defaultWorkPhases,
	}
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cadence", "config.toml"), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Default(), err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Default(), err
	}

	return normalize(cfg), nil
}

func LoadWithOverrides(workMinutes, breakMinutes int) (Config, error) {
	cfg, err := Load()
	cfg = ApplyOverrides(cfg, workMinutes, breakMinutes)
	return cfg, err
}

func ApplyOverrides(cfg Config, workMinutes, breakMinutes int) Config {
	if workMinutes > 0 {
		cfg.WorkMinutes = workMinutes
	}
	if breakMinutes > 0 {
		cfg.BreakMinutes = breakMinutes
	}
	return normalize(cfg)
}

func Save(cfg Config) error {
	cfg = normalize(cfg)

	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
}

func normalize(cfg Config) Config {
	if cfg.WorkMinutes <= 0 {
		cfg.WorkMinutes = defaultWorkMinutes
	}
	if cfg.BreakMinutes <= 0 {
		cfg.BreakMinutes = defaultBreakMinutes
	}
	if cfg.WorkPhases <= 0 {
		cfg.WorkPhases = defaultWorkPhases
	}
	return cfg
}
