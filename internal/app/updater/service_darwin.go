package updater

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// postUpdate updates Info.plist version strings after the binary is swapped in.
// Codesigning is deferred to restartWithCodesign so it runs after the process exits.
func postUpdate(exe, newVersion string) error {
	macosDir := filepath.Dir(exe)
	if filepath.Base(macosDir) != "MacOS" {
		return nil
	}
	contentsDir := filepath.Dir(macosDir)
	bundlePath := filepath.Dir(contentsDir)
	if !strings.HasSuffix(bundlePath, ".app") {
		return nil
	}

	plist := filepath.Join(contentsDir, "Info.plist")
	if err := exec.Command("plutil", "-replace", "CFBundleShortVersionString",
		"-string", newVersion, plist).Run(); err != nil {
		return fmt.Errorf("update CFBundleShortVersionString: %w", err)
	}
	if err := exec.Command("plutil", "-replace", "CFBundleVersion",
		"-string", newVersion, plist).Run(); err != nil {
		return fmt.Errorf("update CFBundleVersion: %w", err)
	}

	return nil
}

// restartWithCodesign launches a background shell that waits for the current
// process to exit, then ad-hoc signs the bundle, removes quarantine, and
// reopens the app. Must be called just before the process quits.
func restartWithCodesign(bundlePath string, pid int) {
	script := fmt.Sprintf(
		`while kill -0 %d 2>/dev/null; do sleep 0.1; done
codesign --force --deep --sign - %q
xattr -d com.apple.quarantine %q 2>/dev/null || true
open %q`,
		pid, bundlePath, bundlePath, bundlePath,
	)
	cmd := exec.Command("sh", "-c", script)
	_ = cmd.Start()
}

// getBundlePath returns the .app bundle path containing exe, or "" if not inside a bundle.
func getBundlePath(exe string) string {
	macosDir := filepath.Dir(exe)
	if filepath.Base(macosDir) != "MacOS" {
		return ""
	}
	contentsDir := filepath.Dir(macosDir)
	bundlePath := filepath.Dir(contentsDir)
	if !strings.HasSuffix(bundlePath, ".app") {
		return ""
	}
	return bundlePath
}
