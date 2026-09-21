package cmd

import (
	"os"
	"path/filepath"
	"runtime"
)

func defaultLocalConfigDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "nvim")
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "nvim")
		}
	}
	return filepath.Join(home, ".config", "nvim")
}
