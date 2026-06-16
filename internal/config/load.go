package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FontSize int  `yaml:"fontSize"`
	Diff     Diff `yaml:"diff"`
}

type Diff struct {
	AddBg     string `yaml:"addBg"`
	AddText   string `yaml:"addText"`
	DelBg     string `yaml:"delBg"`
	DelText   string `yaml:"delText"`
	DelStrike bool   `yaml:"delStrike"`
}

func defaultConfig() Config {
	return Config{
		FontSize: 14,
		Diff: Diff{
			AddBg:     "#d4f4dd",
			AddText:   "#1a7f37",
			DelBg:     "#ffd7d5",
			DelText:   "#cf222e",
			DelStrike: true,
		},
	}
}

func Load(dir string) (Config, error) {
	cfg := defaultConfig()
	files := []string{
		filepath.Join(dir, "setting.yaml"),
		filepath.Join(dir, "setting.local.yaml"),
	}
	for _, f := range files {
		if data, err := os.ReadFile(f); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return cfg, err
			}
		}
	}
	return cfg, nil
}

func Save(dir string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "setting.local.yaml")
	return os.WriteFile(path, data, 0644)
}
