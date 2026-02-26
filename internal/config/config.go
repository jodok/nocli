package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	BaseURL      string `json:"base_url,omitempty"`
	TokenV2      string `json:"token_v2,omitempty"`
	NotionUserID string `json:"notion_user_id,omitempty"`
	ActiveUserID string `json:"active_user_id,omitempty"`
	Cookie       string `json:"cookie,omitempty"`
}

func Read(path string) (File, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return File{}, nil
	}

	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			// Backward compatibility: if default ~/.nocli.json does not exist,
			// transparently try the legacy ~/.notion.json.
			if p == DefaultPath() {
				if legacy := LegacyPath(); legacy != "" {
					if lb, lerr := os.ReadFile(legacy); lerr == nil {
						var legacyCfg File
						if jerr := json.Unmarshal(lb, &legacyCfg); jerr == nil {
							return legacyCfg, nil
						}
					}
				}
			}
			return File{}, nil
		}
		return File{}, fmt.Errorf("read config %s: %w", p, err)
	}

	var cfg File
	if err := json.Unmarshal(b, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", p, err)
	}
	return cfg, nil
}

func ResolvePath(input string) string {
	p := strings.TrimSpace(input)
	if p != "" {
		if expanded := expandHomePath(p); expanded != "" {
			return expanded
		}
		return p
	}
	return DefaultPath()
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".nocli.json"
	}
	return filepath.Join(home, ".nocli.json")
}

func LegacyPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".notion.json"
	}
	return filepath.Join(home, ".notion.json")
}

func expandHomePath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}

func Write(path string, cfg File) error {
	p := ResolvePath(path)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil && filepath.Dir(p) != "." {
		return fmt.Errorf("create config dir: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	b = append(b, '\n')

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("commit config: %w", err)
	}
	if err := os.Chmod(p, 0o600); err != nil {
		return fmt.Errorf("chmod config: %w", err)
	}
	return nil
}
