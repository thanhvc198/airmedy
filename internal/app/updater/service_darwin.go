package updater

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// postUpdate updates Info.plist version strings, re-signs the bundle ad-hoc,
// and removes the quarantine attribute after the binary is swapped in.
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

	// Re-sign ad-hoc so macOS accepts the new binary.
	if err := exec.Command("codesign", "--force", "--deep", "--sign", "-", bundlePath).Run(); err != nil {
		return fmt.Errorf("ad-hoc codesign: %w", err)
	}

	// Remove quarantine flag set by macOS on downloaded files.
	// Ignore error — flag may not be present.
	_ = exec.Command("xattr", "-d", "com.apple.quarantine", bundlePath).Run()

	return nil
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
