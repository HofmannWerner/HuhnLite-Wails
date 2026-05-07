package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	appDataDir := ""
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataDir = filepath.Join(configDir, "HuhnLite")
	}

	cwd, _ := os.Getwd()

	execPath, err := os.Executable()
	var bundleDir string
	if err == nil {
		bundleDir = filepath.Dir(execPath)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
	}

	fmt.Println("UserConfigDir: ", appDataDir)
	fmt.Println("CWD: ", cwd)
	fmt.Println("ExecPath: ", execPath)
	fmt.Println("BundleDir: ", bundleDir)

	defaultDB := "HuhnLite.db"
	if appDataDir != "" {
		defaultDB = filepath.Join(appDataDir, "HuhnLite.db")
	}

	if bundleDir != "" {
		statPath := filepath.Join(bundleDir, "HuhnLite.db")
		if _, err := os.Stat(statPath); err == nil {
			defaultDB = statPath
			fmt.Println("Found DB in BundleDir: ", statPath)
		} else {
			fmt.Println("DB not in BundleDir: ", statPath, err)
		}
	}
	if cwd != "" && cwd != bundleDir {
		statPath := filepath.Join(cwd, "HuhnLite.db")
		if _, err := os.Stat(statPath); err == nil {
			defaultDB = statPath
			fmt.Println("Found DB in CWD: ", statPath)
		} else {
			fmt.Println("DB not in CWD: ", statPath, err)
		}
	}

	fmt.Println("FINAL DB PATH: ", defaultDB)
}
