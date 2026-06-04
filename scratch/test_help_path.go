package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var pathsToCheck []string

	// 1. Check CWD
	if cwd, err := os.Getwd(); err == nil {
		pathsToCheck = append(pathsToCheck, filepath.Join(cwd, "HuhnLite-de.html"))
	}

	// 2. Check AppDataDir
	if configDir, err := os.UserConfigDir(); err == nil {
		pathsToCheck = append(pathsToCheck, filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite-de.html"))
		pathsToCheck = append(pathsToCheck, filepath.Join(configDir, "HuhnLite-de.html"))
	}

	// 3. Check LOCALAPPDATA
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, "HuhnLite-Wails", "HuhnLite-de.html"))
		pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, "HuhnLite-de.html"))
	}

	// 4. Check APPDATA
	if appData := os.Getenv("APPDATA"); appData != "" {
		pathsToCheck = append(pathsToCheck, filepath.Join(appData, "HuhnLite-Wails", "HuhnLite-de.html"))
		pathsToCheck = append(pathsToCheck, filepath.Join(appData, "HuhnLite-de.html"))
	}

	fmt.Println("Paths checked:")
	for _, p := range pathsToCheck {
		_, err := os.Stat(p)
		exists := err == nil
		fmt.Printf("- %s (exists: %v)\n", p, exists)
		if exists {
			fmt.Printf("  -> Parent Directory: %s\n", filepath.Dir(p))
		}
	}
}
