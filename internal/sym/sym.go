// Package sym implements
package sym

import (
	"fmt"
	"os"
	"path/filepath"
)

const version = "1.0.0"

type Config struct {
	SymDir    string
	TargetDir string
	Verbose   bool
	Simulate  bool
	Delete    bool
	ReSym     bool
	Packages  []string
}

func ProcessPackage(config *Config, pkg string) error {
	pkgPath := filepath.Join(config.SymDir, pkg)

	// Check if package directory exists
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return fmt.Errorf("package directory does not exist: %s", pkgPath)
	}

	if config.ReSym {
		// Unstow first, then stow
		if err := unsymPackage(config, pkg, pkgPath); err != nil {
			return fmt.Errorf("failed to unsym during resym: %w", err)
		}
		return symPackage(config, pkg, pkgPath)
	} else if config.Delete {
		return unsymPackage(config, pkg, pkgPath)
	} else {
		return symPackage(config, pkg, pkgPath)
	}
}

func symPackage(config *Config, pkg string, pkgPath string) error {
	if config.Verbose {
		fmt.Printf("Stowing package: %s\n", pkg)
	}

	return filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the package root directory
		if path == pkgPath {
			return nil
		}

		// Get relative path within the package
		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return err
		}

		// Target path in the target directory
		targetPath := filepath.Join(config.TargetDir, relPath)

		if info.IsDir() {
			// Create directory if it doesn't exist
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				if config.Verbose {
					fmt.Printf("Creating directory: %s\n", targetPath)
				}
				if !config.Simulate {
					if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
					}
				}
			}
		} else {
			// Create symlink for files
			if err := createSymlink(config, path, targetPath); err != nil {
				return err
			}
		}

		return nil
	})
}

func unsymPackage(config *Config, pkg string, pkgPath string) error {
	if config.Verbose {
		fmt.Printf("Unstowing package: %s\n", pkg)
	}

	return filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the package root directory
		if path == pkgPath {
			return nil
		}

		// Get relative path within the package
		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return err
		}

		// Target path in the target directory
		targetPath := filepath.Join(config.TargetDir, relPath)

		if !info.IsDir() {
			// Remove symlink if it points to our file
			if err := removeSymlink(config, path, targetPath); err != nil {
				return err
			}
		}

		return nil
	})
}

func createSymlink(config *Config, srcPath, targetPath string) error {
	// Check if target already exists
	if _, err := os.Lstat(targetPath); err == nil {
		// Check if it's already the correct symlink
		if link, err := os.Readlink(targetPath); err == nil {
			if link == srcPath {
				if config.Verbose {
					fmt.Printf("Symlink already exists: %s -> %s\n", targetPath, srcPath)
				}
				return nil
			} else {
				return fmt.Errorf("target %s already exists and points to %s (not %s)",
					targetPath, link, srcPath)
			}
		} else {
			return fmt.Errorf("target %s already exists and is not a symlink", targetPath)
		}
	}

	// Create parent directories if needed
	targetDir := filepath.Dir(targetPath)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if config.Verbose {
			fmt.Printf("Creating parent directory: %s\n", targetDir)
		}
		if !config.Simulate {
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", targetDir, err)
			}
		}
	}

	if config.Verbose {
		fmt.Printf("Creating symlink: %s -> %s\n", targetPath, srcPath)
	}

	if !config.Simulate {
		if err := os.Symlink(srcPath, targetPath); err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w", targetPath, srcPath, err)
		}
	}

	return nil
}

func removeSymlink(config *Config, srcPath, targetPath string) error {
	// Check if target exists and is a symlink
	info, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		if config.Verbose {
			fmt.Printf("Target does not exist: %s\n", targetPath)
		}
		return nil
	}
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		if config.Verbose {
			fmt.Printf("Target is not a symlink: %s\n", targetPath)
		}
		return nil
	}

	// Check if symlink points to our source
	link, err := os.Readlink(targetPath)
	if err != nil {
		return err
	}

	if link != srcPath {
		if config.Verbose {
			fmt.Printf("Symlink %s points to %s (not %s), leaving alone\n",
				targetPath, link, srcPath)
		}
		return nil
	}

	if config.Verbose {
		fmt.Printf("Removing symlink: %s\n", targetPath)
	}

	if !config.Simulate {
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to remove symlink %s: %w", targetPath, err)
		}
	}

	return nil
}
