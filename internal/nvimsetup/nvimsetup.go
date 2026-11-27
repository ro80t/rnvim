// Package nvimsetup ensures a working nvim (+ user config) exists on a
// remote target: try installing it over the network via the target's
// package manager first; if that's not possible (no network, no supported
// package manager, no sudo, ...), fall back to pushing a local nvim binary
// and config. The caller can also force one strategy via Options.Strategy.
package nvimsetup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ro80t/rnvim/internal/transport"
)

type Options struct {
	Strategy     string // "auto" (default) | "install" | "push"
	LocalConfig  string // local nvim config dir to copy; "" skips config copy
	LocalNvimBin string // local nvim binary for push strategy; "" = look up PATH
}

// Result describes how to actually invoke nvim on the target.
type Result struct {
	Exec string
	Env  map[string]string
}

func Ensure(t transport.Transport, opt Options) (*Result, error) {
	switch opt.Strategy {
	case "", "auto":
		if res, err := install(t, opt.LocalConfig); err == nil {
			return res, nil
		}
		fmt.Fprintln(os.Stderr, "rnvim: network install unavailable, falling back to pushing local nvim + config")
		return push(t, opt)
	case "install":
		res, err := install(t, opt.LocalConfig)
		if err != nil {
			return nil, fmt.Errorf("install strategy failed: %w", err)
		}
		return res, nil
	case "push":
		return push(t, opt)
	default:
		return nil, fmt.Errorf("unknown strategy %q (want auto|install|push)", opt.Strategy)
	}
}

// installScript detects the target's package manager and installs neovim.
// It's a no-op if nvim is already present.
const installScript = `
set -e
SUDO=""
if [ "$(id -u)" != "0" ] && command -v sudo >/dev/null 2>&1; then SUDO="sudo"; fi
if command -v nvim >/dev/null 2>&1; then exit 0; fi
if command -v apt-get >/dev/null 2>&1; then
  $SUDO apt-get update && $SUDO apt-get install -y neovim
elif command -v dnf >/dev/null 2>&1; then
  $SUDO dnf install -y neovim
elif command -v apk >/dev/null 2>&1; then
  $SUDO apk add --no-cache neovim
elif command -v pacman >/dev/null 2>&1; then
  $SUDO pacman -Sy --noconfirm neovim
elif command -v brew >/dev/null 2>&1; then
  brew install neovim
else
  echo "rnvim: no supported package manager found" >&2
  exit 1
fi
`

// RemoteBase is where rnvim keeps everything it puts on a target, so the
// target's own ~/.config/nvim, ~/.local/share/nvim, ~/.vim etc. are never
// touched.
func RemoteBase(t transport.Transport) string {
	home, err := t.Home()
	if err != nil || home == "" {
		home = "/root"
	}
	return home + "/.rnvim"
}

// IsolatedEnv points nvim's XDG dirs into base and copies localConfig (if
// it exists) to base/config/nvim.
func IsolatedEnv(t transport.Transport, base, localConfig string) (map[string]string, error) {
	env := map[string]string{
		"XDG_CONFIG_HOME": base + "/config",
		"XDG_DATA_HOME":   base + "/data",
		"XDG_STATE_HOME":  base + "/state",
		"XDG_CACHE_HOME":  base + "/cache",
	}
	if localConfig == "" {
		return env, nil
	}
	if _, err := os.Stat(localConfig); err != nil {
		return env, nil // no local config to copy, not an error
	}
	// CopyTo nests into an existing dir instead of replacing it, so clear
	// the previous copy first.
	dst := base + "/config/nvim"
	if _, err := t.Run("rm -rf '" + dst + "'"); err != nil {
		return nil, err
	}
	if err := t.CopyTo(localConfig, dst); err != nil {
		return nil, fmt.Errorf("copy config: %w", err)
	}
	return env, nil
}

func install(t transport.Transport, localConfig string) (*Result, error) {
	if _, err := t.Run(installScript); err != nil {
		return nil, err
	}
	env, err := IsolatedEnv(t, RemoteBase(t), localConfig)
	if err != nil {
		return nil, err
	}
	return &Result{Exec: "nvim", Env: env}, nil
}

// push copies a local nvim binary + its runtime dir + the user's config
// onto the target and points the caller at them directly, bypassing any
// remote package manager.
func push(t transport.Transport, opt Options) (*Result, error) {
	nvimBin := opt.LocalNvimBin
	if nvimBin == "" {
		var err error
		nvimBin, err = exec.LookPath("nvim")
		if err != nil {
			return nil, fmt.Errorf("push strategy: local nvim not found in PATH (pass --nvim-bin): %w", err)
		}
	}
	resolved, err := filepath.EvalSymlinks(nvimBin)
	if err != nil {
		resolved = nvimBin
	}

	if err := checkArchMatch(t); err != nil {
		return nil, err
	}
	remoteBase := RemoteBase(t)

	// Standard nvim install layouts (official tarball, Homebrew, apt, ...)
	// put runtime files at <prefix>/share/nvim relative to <prefix>/bin/nvim.
	runtimeDir := filepath.Join(filepath.Dir(filepath.Dir(resolved)), "share", "nvim")
	if _, err := os.Stat(runtimeDir); err != nil {
		return nil, fmt.Errorf("push strategy: could not find nvim runtime dir next to %s (expected <prefix>/share/nvim); install nvim from the official release layout or pass --nvim-bin", resolved)
	}

	if err := t.CopyTo(resolved, remoteBase+"/bin/nvim"); err != nil {
		return nil, fmt.Errorf("copy nvim binary: %w", err)
	}
	if _, err := t.Run("chmod +x " + remoteBase + "/bin/nvim"); err != nil {
		return nil, err
	}
	if _, err := t.Run("rm -rf '" + remoteBase + "/share/nvim'"); err != nil {
		return nil, err
	}
	if err := t.CopyTo(runtimeDir, remoteBase+"/share/nvim"); err != nil {
		return nil, fmt.Errorf("copy nvim runtime: %w", err)
	}

	env, err := IsolatedEnv(t, remoteBase, opt.LocalConfig)
	if err != nil {
		return nil, err
	}
	env["VIMRUNTIME"] = remoteBase + "/share/nvim/runtime"

	return &Result{Exec: remoteBase + "/bin/nvim", Env: env}, nil
}

var unameOS = map[string]string{"linux": "Linux", "darwin": "Darwin", "windows": "Windows"}
var unameArch = map[string]string{"amd64": "x86_64", "arm64": "aarch64", "386": "i686"}

func checkArchMatch(t transport.Transport) error {
	remoteOS, remoteArch, err := t.Uname()
	if err != nil {
		return fmt.Errorf("push strategy: could not detect remote OS/arch: %w", err)
	}
	localOS, localArch := unameOS[runtime.GOOS], unameArch[runtime.GOARCH]
	if !strings.EqualFold(remoteOS, localOS) || !strings.EqualFold(remoteArch, localArch) {
		return fmt.Errorf("push strategy: local nvim is %s/%s but remote is %s/%s; they must match architectures (build/download a matching nvim binary and pass --nvim-bin, or use --strategy=install)",
			runtime.GOOS, runtime.GOARCH, remoteOS, remoteArch)
	}
	return nil
}
