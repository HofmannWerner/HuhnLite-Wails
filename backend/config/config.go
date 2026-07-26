package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Mandant            int    `json:"mandant"`
	Mode               string `json:"mode"`          // "standalone" or "server"
	DBEngine           string `json:"db_engine"`     // "sqlite" or "mysql"
	DBConnectionString string `json:"db_connection"` // e.g. "HuhnLite.db" or "user:pass@tcp(127.0.0.1:3306)/dbname"
	DBConnectionTest   string `json:"db_connection_test"`
	Test               int    `json:"test"`
	Port               int    `json:"port"`          // HTTP Port for server mode or standalone Gin server
	System             int    `json:"system"`        // 1 = Erlaube Bearbeiten von System-Einträgen
	AutoBackup         int    `json:"autobackup"`    // 0 = kein backup, 1 = Beim Start, 2 = Beim Programmende, 3 = Start & Ende
	BackupTimeStr      string `json:"backuptime"`    // e.g. "1200,20:00"
	WaitTimeStr        string `json:"waittime"`      // e.g. "00:01" (hh:mm)
	LauncherPort       int    `json:"launcher_port"`
	Language           string `json:"language"`
	HasCLILanguage     bool   `json:"-"`
	ConfigFilePath     string `json:"-"`
}

type safeWriter struct {
	w io.Writer
}

func (sw *safeWriter) Write(p []byte) (n int, err error) {
	n, _ = sw.w.Write(p)
	return n, nil
}

func LoadConfig() Config {
	cwd, _ := os.Getwd()
	execPath, errExec := os.Executable()
	var bundleDir string
	if errExec == nil {
		bundleDir = filepath.Dir(execPath)
		// Falls wir in einem macOS .app Bundle sind (Contents/MacOS/...)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			// Gehe 3 Ebenen hoch: .../HuhnLite-Wails.app -> Verzeichnis, in dem die .app liegt
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
		fmt.Printf("BundleDir: %s\n", bundleDir)
	}

	appDataDir := ""
	var logFile *os.File
	var logPath string
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataDir = filepath.Join(configDir, "HuhnLite")
		if err := os.MkdirAll(appDataDir, 0755); err == nil {
			logPath = filepath.Join(appDataDir, "app.log")
			var errFile error
			logFile, errFile = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if errFile != nil {
				logPath = filepath.Join(appDataDir, fmt.Sprintf("app_%d.log", os.Getpid()))
				logFile, _ = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			}
		} else {
			fmt.Printf("ERROR creating AppDataDir: %v\n", err)
		}
	} else {
		fmt.Printf("ERROR getting UserConfigDir: %v\n", err)
	}

	// Lokal im Verzeichnis der Executable (bzw. CWD als Fallback) loggen
	localLogDir := bundleDir
	if localLogDir == "" {
		localLogDir = cwd
	}
	localLogPath := filepath.Join(localLogDir, "app.log")
	logFileLocal, errLocal := os.OpenFile(localLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if errLocal != nil {
		localLogPath = filepath.Join(localLogDir, fmt.Sprintf("app_%d.log", os.Getpid()))
		logFileLocal, _ = os.OpenFile(localLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}

	var writers []io.Writer
	writers = append(writers, &safeWriter{w: os.Stdout})
	
	if logFile != nil {
		writers = append(writers, logFile)
	}
	if logFileLocal != nil {
		writers = append(writers, logFileLocal)
	}
	
	var multiWriter io.Writer = io.MultiWriter(writers...)
	log.SetOutput(multiWriter)
	log.Printf("--- App gestartet (PID: %d) ---", os.Getpid())
	if logPath != "" {
		log.Printf("Log-Datei AppData: %s", logPath)
	}
	log.Printf("Log-Datei Lokal: %s", localLogPath)

	parent := filepath.Dir(cwd)

	// Check if running from Program Files or macOS Applications (read-only for standard users)
	isProgramFiles := false
	if bundleDir != "" {
		lowerBundleDir := strings.ToLower(bundleDir)
		if strings.Contains(lowerBundleDir, "program files") || strings.Contains(lowerBundleDir, "programmdateien") || strings.Contains(lowerBundleDir, "programme") || strings.Contains(lowerBundleDir, "/applications") {
			isProgramFiles = true
		}
	}

	if isProgramFiles && appDataDir != "" {
		sourceDirForCopy := bundleDir
		// If running in macOS bundle, copy from Contents/Resources
		if errExec == nil && filepath.Base(filepath.Dir(execPath)) == "MacOS" && filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			sourceDirForCopy = filepath.Join(filepath.Dir(filepath.Dir(execPath)), "Resources")
		}

		// Files to copy from sourceDirForCopy to appDataDir if they don't exist in appDataDir
		filesToCopy := []string{
			"HuhnLite.db",
			"HuhnLite_test.db",
			"HuhnLite_prod.db",
			"settings.json",
			"settings_mariadb.json",
			"settings_server.json",
			"settings_server_mariadb.json",
		}
		for _, f := range filesToCopy {
			src := filepath.Join(sourceDirForCopy, f)
			dst := filepath.Join(appDataDir, f)
			if _, err := os.Stat(src); err == nil {
				if _, errDst := os.Stat(dst); os.IsNotExist(errDst) {
					log.Printf("Copying %s from bundle to AppData: %s", f, dst)
					if errCopy := copyFile(src, dst); errCopy != nil {
						log.Printf("ERROR copying %s: %v", f, errCopy)
					}
				}
			}
		}
	}

	// 1. Fallback: Application Support Verzeichnis
	defaultDB := "HuhnLite.db"
	if appDataDir != "" {
		defaultDB = filepath.Join(appDataDir, "HuhnLite.db")
	}

	// 2. Bevorzuge eine bestehende HuhnLite.db neben der App (Portable Mode)
	if bundleDir != "" && !isProgramFiles {
		fullPath := filepath.Join(bundleDir, "HuhnLite.db")
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Found DB at BundleDir: %s\n", fullPath)
			defaultDB = fullPath
		} else {
			fmt.Printf("DB NOT found at BundleDir: %s (%v)\n", fullPath, err)
		}
	}
	// 3. Bevorzuge eine bestehende HuhnLite.db im aktuellen Terminal-Ordner (z.B. für wails dev)
	if cwd != "" && cwd != bundleDir {
		fullPath := filepath.Join(cwd, "HuhnLite.db")
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Found DB at CWD: %s\n", fullPath)
			defaultDB = fullPath
		} else {
			fmt.Printf("DB NOT found at CWD: %s (%v)\n", fullPath, err)
		}
	}
	fmt.Printf("Selected Default DB: %s\n", defaultDB)

	// Parse CLI arguments early so CLI Mandant/Port overrides are known upfront
	ov := parseCLIArgs(os.Args[1:])

	// Default configuration
	cfg := Config{
		Mode:               "standalone",
		DBEngine:           "sqlite",
		DBConnectionString: defaultDB,
		DBConnectionTest:   "",
		Test:               0,
		Port:               8080,
		LauncherPort:       8080,
		Language:           "de",
		System:             0,
		AutoBackup:         -1,
	}

	// Name der Einstellungsdatei basierend auf dem Programm-Namen bestimmen
	configFiles := []string{"settings.json"}

	if execPath != "" {
		fullPath := strings.ToLower(execPath)
		log.Printf("ExecPath: %s", execPath)
		if strings.Contains(fullPath, "server") && strings.Contains(fullPath, "mariadb") {
			configFiles = []string{"settings_server_mariadb.json", "settings.json"}
			log.Printf("Server-MariaDB-Modus erkannt, priorisiere settings_server_mariadb.json")
		} else if strings.Contains(fullPath, "mariadb") {
			configFiles = []string{"settings_mariadb.json", "settings.json"}
			log.Printf("MariaDB-Modus erkannt, priorisiere settings_mariadb.json")
		} else if strings.Contains(fullPath, "server") {
			configFiles = []string{"settings_server.json", "settings.json"}
			log.Printf("Server-Modus erkannt, priorisiere settings_server.json")
		}
	}

	var loaded bool
	var rawMap map[string]interface{}
	for _, configName := range configFiles {
		var paths []string
		if isProgramFiles && appDataDir != "" {
			paths = []string{
				filepath.Join(appDataDir, configName),
				filepath.Join(cwd, configName),
				filepath.Join(parent, configName),
			}
		} else {
			paths = []string{
				filepath.Join(cwd, configName),
				filepath.Join(parent, configName),
			}
			if appDataDir != "" {
				paths = append(paths, filepath.Join(appDataDir, configName))
			}
			if bundleDir != "" && !isProgramFiles {
				paths = append(paths, filepath.Join(bundleDir, configName))
			}
		}
		if errExec == nil && filepath.Base(filepath.Dir(execPath)) == "MacOS" && filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			resourcesDir := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "Resources")
			paths = append(paths, filepath.Join(resourcesDir, configName))
		}

		for _, p := range paths {
			file, err := os.Open(p)
			if err == nil {
				defer file.Close()
				data, errRead := io.ReadAll(file)
				if errRead == nil {
					if errDec := json.Unmarshal(data, &cfg); errDec == nil {
						cfg.ConfigFilePath = p
						log.Printf("Konfiguration aus %s geladen (Engine: %s)", p, cfg.DBEngine)
						_ = json.Unmarshal(data, &rawMap)

						// Apply CLI Mandant override immediately if provided, otherwise use config file mandant
						if ov.Mandant != nil {
							cfg.Mandant = *ov.Mandant
							log.Printf("[CLI] Overriding file Mandant with CLI Mandant: %d", cfg.Mandant)
						}

						if cfg.Mandant > 0 {
							prodBase := "HuhnLite_prod.db"
							if cfg.DBConnectionString != "" {
								prodBase = filepath.Base(cfg.DBConnectionString)
							}
							testBase := "HuhnLite_test.db"
							if cfg.DBConnectionTest != "" {
								testBase = filepath.Base(cfg.DBConnectionTest)
							}

							settingsDir := filepath.Dir(p)
							if isProgramFiles && appDataDir != "" {
								settingsDir = appDataDir
							}
							cfg.DBConnectionString = filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant), prodBase)
							cfg.DBConnectionTest = filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant), testBase)

							if cfg.DBEngine == "sqlite" {
								mandantDir := filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant))
								if errMk := os.MkdirAll(mandantDir, 0755); errMk == nil {
									mandantProdPath := filepath.Join(mandantDir, prodBase)
									mandantTestPath := filepath.Join(mandantDir, testBase)
									srcProd := filepath.Join(settingsDir, prodBase)
									srcTest := filepath.Join(settingsDir, testBase)

									if _, errStat := os.Stat(mandantProdPath); os.IsNotExist(errStat) {
										if _, errSrc := os.Stat(srcProd); errSrc == nil {
											log.Printf("[Config] Copying base prod DB to %s", mandantProdPath)
											_ = copyFile(srcProd, mandantProdPath)
										}
									}
									if _, errStat := os.Stat(mandantTestPath); os.IsNotExist(errStat) {
										if _, errSrc := os.Stat(srcTest); errSrc == nil {
											log.Printf("[Config] Copying base test DB to %s", mandantTestPath)
											_ = copyFile(srcTest, mandantTestPath)
										}
									}
								}
							}

							testKey := fmt.Sprintf("test_%d", cfg.Mandant)
							systemKey := fmt.Sprintf("system_%d", cfg.Mandant)
							autobackupKey := fmt.Sprintf("autobackup_%d", cfg.Mandant)

							if tVal, ok := rawMap[testKey]; ok {
								if f, ok := tVal.(float64); ok {
									cfg.Test = int(f)
								} else if s, ok := tVal.(string); ok {
									if s == "1" || strings.ToLower(s) == "true" {
										cfg.Test = 1
									} else {
										cfg.Test = 0
									}
								} else if b, ok := tVal.(bool); ok {
									if b {
										cfg.Test = 1
									} else {
										cfg.Test = 0
									}
								}
							} else {
								cfg.Test = 0
							}

							if sVal, ok := rawMap[systemKey]; ok {
								if f, ok := sVal.(float64); ok {
									cfg.System = int(f)
								} else if s, ok := sVal.(string); ok {
									if s == "1" || strings.ToLower(s) == "true" {
										cfg.System = 1
									} else {
										cfg.System = 0
									}
								} else if b, ok := sVal.(bool); ok {
									if b {
										cfg.System = 1
									} else {
										cfg.System = 0
									}
								}
							} else {
								cfg.System = 0
							}

							if abVal, ok := rawMap[autobackupKey]; ok {
								if f, ok := abVal.(float64); ok {
									cfg.AutoBackup = int(f)
								} else if s, ok := abVal.(string); ok {
									var temp int
									if _, err := fmt.Sscanf(s, "%d", &temp); err == nil {
										cfg.AutoBackup = temp
									}
								}
							} else {
								cfg.AutoBackup = -1
							}

							backuptimeKey := fmt.Sprintf("backuptime_%d", cfg.Mandant)
							if btVal, ok := rawMap[backuptimeKey]; ok {
								if s, ok := btVal.(string); ok {
									cfg.BackupTimeStr = s
								}
							}
						} else {
							// Bei SQLite und relativem Pfad: Pfad relativ zur settings.json auflösen
							if cfg.DBEngine == "sqlite" && !filepath.IsAbs(cfg.DBConnectionString) {
								// Clean up Mac-style paths that might have been incorrectly treated as relative on Windows
								if strings.HasPrefix(cfg.DBConnectionString, "/Users/") || strings.HasPrefix(cfg.DBConnectionString, "Users/") {
									cfg.DBConnectionString = filepath.Base(cfg.DBConnectionString)
									log.Printf("Bereinigter Mac-Pfad zu: %s", cfg.DBConnectionString)
								}
								cfg.DBConnectionString = filepath.Join(filepath.Dir(p), cfg.DBConnectionString)
							}
							if cfg.DBEngine == "sqlite" && cfg.DBConnectionTest != "" && !filepath.IsAbs(cfg.DBConnectionTest) {
								if strings.HasPrefix(cfg.DBConnectionTest, "/Users/") || strings.HasPrefix(cfg.DBConnectionTest, "Users/") {
									cfg.DBConnectionTest = filepath.Base(cfg.DBConnectionTest)
								}
								cfg.DBConnectionTest = filepath.Join(filepath.Dir(p), cfg.DBConnectionTest)
							}

							if abVal, ok := rawMap["autobackup"]; ok {
								if f, ok := abVal.(float64); ok {
									cfg.AutoBackup = int(f)
								} else if s, ok := abVal.(string); ok {
									var temp int
									if _, err := fmt.Sscanf(s, "%d", &temp); err == nil {
										cfg.AutoBackup = temp
									}
								}
							} else {
								cfg.AutoBackup = -1
							}

							if btVal, ok := rawMap["backuptime"]; ok {
								if s, ok := btVal.(string); ok {
									cfg.BackupTimeStr = s
								}
							}
						}
						loaded = true
						break
					}
				}
			}
		}
		if loaded {
			break
		}
	}
	if !loaded {
		log.Printf("Warnung: settings.json nicht gefunden, verwende Defaults (DB: %s)", cfg.DBConnectionString)
	}

	// Command Line Parameter auswerten (z.B. Port=8081, mandant=1)
	ov = parseCLIArgs(os.Args[1:])
	if ov.Port != nil {
		cfg.Port = *ov.Port
		log.Printf("[CLI] Overriding Port: %d", cfg.Port)
	}
	if ov.Mode != nil {
		cfg.Mode = *ov.Mode
		log.Printf("[CLI] Overriding Mode: %s", cfg.Mode)
	}
	if ov.DBEngine != nil {
		cfg.DBEngine = *ov.DBEngine
		log.Printf("[CLI] Overriding DBEngine: %s", cfg.DBEngine)
	}
	if ov.Test != nil {
		cfg.Test = *ov.Test
		log.Printf("[CLI] Overriding Test: %d", cfg.Test)
	}
	if ov.Mandant != nil {
		cfg.Mandant = *ov.Mandant
		log.Printf("[CLI] Overriding Mandant: %d", cfg.Mandant)
	}
	if ov.WaitTimeStr != nil {
		cfg.WaitTimeStr = *ov.WaitTimeStr
		log.Printf("[CLI] Overriding WaitTime: %s", cfg.WaitTimeStr)
	}
	if ov.LauncherPort != nil {
		cfg.LauncherPort = *ov.LauncherPort
		log.Printf("[CLI] Overriding LauncherPort: %d", cfg.LauncherPort)
	}
	if ov.Language != nil {
		cfg.Language = *ov.Language
		cfg.HasCLILanguage = true
		log.Printf("[CLI] Overriding Language: %s", cfg.Language)
	}
	if cfg.Language == "" {
		cfg.Language = "de"
	}

	if ov.Mandant != nil {
		settingsDir := filepath.Dir(cfg.ConfigFilePath)
		if isProgramFiles && appDataDir != "" {
			settingsDir = appDataDir
		} else if cfg.ConfigFilePath == "" {
			if bundleDir != "" && !isProgramFiles {
				settingsDir = bundleDir
			} else {
				settingsDir = cwd
			}
		}

		prodBase := "HuhnLite_prod.db"
		testBase := "HuhnLite_test.db"

		cfg.DBConnectionString = filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant), prodBase)
		cfg.DBConnectionTest = filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant), testBase)

		if cfg.DBEngine == "sqlite" {
			mandantDir := filepath.Join(settingsDir, fmt.Sprintf("mandant_%d", cfg.Mandant))
			if errMk := os.MkdirAll(mandantDir, 0755); errMk == nil {
				mandantProdPath := filepath.Join(mandantDir, prodBase)
				mandantTestPath := filepath.Join(mandantDir, testBase)
				srcProd := filepath.Join(settingsDir, prodBase)
				srcTest := filepath.Join(settingsDir, testBase)

				if _, errStat := os.Stat(mandantProdPath); os.IsNotExist(errStat) {
					if _, errSrc := os.Stat(srcProd); errSrc == nil {
						log.Printf("[Config] Copying base prod DB to %s", mandantProdPath)
						_ = copyFile(srcProd, mandantProdPath)
					}
				}
				if _, errStat := os.Stat(mandantTestPath); os.IsNotExist(errStat) {
					if _, errSrc := os.Stat(srcTest); errSrc == nil {
						log.Printf("[Config] Copying base test DB to %s", mandantTestPath)
						_ = copyFile(srcTest, mandantTestPath)
					}
				}
			}
		}

		// Update mandant-specific settings from rawMap if CLI changed mandant
		if loaded && rawMap != nil {
			testKey := fmt.Sprintf("test_%d", cfg.Mandant)
			systemKey := fmt.Sprintf("system_%d", cfg.Mandant)
			autobackupKey := fmt.Sprintf("autobackup_%d", cfg.Mandant)
			backuptimeKey := fmt.Sprintf("backuptime_%d", cfg.Mandant)

			if tVal, ok := rawMap[testKey]; ok {
				if f, ok := tVal.(float64); ok {
					cfg.Test = int(f)
				} else if s, ok := tVal.(string); ok {
					if s == "1" || strings.ToLower(s) == "true" {
						cfg.Test = 1
					} else {
						cfg.Test = 0
					}
				} else if b, ok := tVal.(bool); ok {
					if b {
						cfg.Test = 1
					} else {
						cfg.Test = 0
					}
				}
			} else if ov.Test == nil {
				cfg.Test = 0
			}

			if sVal, ok := rawMap[systemKey]; ok {
				if f, ok := sVal.(float64); ok {
					cfg.System = int(f)
				} else if s, ok := sVal.(string); ok {
					if s == "1" || strings.ToLower(s) == "true" {
						cfg.System = 1
					} else {
						cfg.System = 0
					}
				} else if b, ok := sVal.(bool); ok {
					if b {
						cfg.System = 1
					} else {
						cfg.System = 0
					}
				}
			} else {
				cfg.System = 0
			}

			if abVal, ok := rawMap[autobackupKey]; ok {
				if f, ok := abVal.(float64); ok {
					cfg.AutoBackup = int(f)
				} else if s, ok := abVal.(string); ok {
					var temp int
					if _, err := fmt.Sscanf(s, "%d", &temp); err == nil {
						cfg.AutoBackup = temp
					}
				}
			} else {
				cfg.AutoBackup = -1
			}

			if btVal, ok := rawMap[backuptimeKey]; ok {
				if s, ok := btVal.(string); ok {
					cfg.BackupTimeStr = s
				}
			}
		}
	}

	return cfg
}

func SaveTestSetting(testVal int, engine string) error {
	configName := "settings.json"
	if engine == "mysql" {
		configName = "settings_mariadb.json"
	}

	cwd, _ := os.Getwd()
	execPath, _ := os.Executable()
	var bundleDir string
	if execPath != "" {
		bundleDir = filepath.Dir(execPath)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
	}
	parent := filepath.Dir(cwd)
	appDataDir := ""
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataDir = filepath.Join(configDir, "HuhnLite")
	}

	paths := []string{
		filepath.Join(cwd, configName),
		filepath.Join(parent, configName),
	}
	if appDataDir != "" {
		paths = append(paths, filepath.Join(appDataDir, configName))
	}
	if bundleDir != "" {
		paths = append(paths, filepath.Join(bundleDir, configName))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			// Read the file
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}

			// Unmarshal into a generic map
			var configMap map[string]interface{}
			if err := json.Unmarshal(data, &configMap); err != nil {
				continue
			}

			// Update test field
			var activeMandant int
			if m, ok := configMap["mandant"].(float64); ok {
				activeMandant = int(m)
			}
			if activeMandant > 0 {
				configMap[fmt.Sprintf("test_%d", activeMandant)] = testVal
			} else {
				configMap["test"] = testVal
			}

			// Marshal back
			newData, err := json.MarshalIndent(configMap, "", "  ")
			if err != nil {
				continue
			}

			// Write back
			err = os.WriteFile(p, newData, 0644)
			if err == nil {
				log.Printf("Successfully saved test=%d to %s", testVal, p)
			}
		}
	}
	return nil
}

// UnmarshalJSON implements a custom JSON unmarshaler to handle multiple types for the "system" field.
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		System interface{} `json:"system"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.System != nil {
		switch v := aux.System.(type) {
		case float64:
			c.System = int(v)
		case string:
			if v == "1" || strings.ToLower(v) == "true" {
				c.System = 1
			} else {
				c.System = 0
			}
		case bool:
			if v {
				c.System = 1
			} else {
				c.System = 0
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}

type cliOverrides struct {
	Port         *int
	Mandant      *int
	Mode         *string
	DBEngine     *string
	Test         *int
	WaitTimeStr  *string
	LauncherPort *int
	Language     *string
}

func tokenizeArgs(args []string) []string {
	var tokens []string
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		var current strings.Builder
		inQuotes := false
		var quoteChar rune

		for _, r := range arg {
			switch r {
			case '"', '\'':
				if inQuotes && quoteChar == r {
					inQuotes = false
				} else if !inQuotes {
					inQuotes = true
					quoteChar = r
				} else {
					current.WriteRune(r)
				}
			case ' ', '\t':
				if inQuotes {
					current.WriteRune(r)
				} else {
					if current.Len() > 0 {
						tokens = append(tokens, current.String())
						current.Reset()
					}
				}
			default:
				current.WriteRune(r)
			}
		}
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
		}
	}
	return tokens
}

func parseCLIArgs(args []string) cliOverrides {
	var ov cliOverrides
	log.Printf("[CLI] Raw command line args received (%d): %v", len(args), args)

	tokens := tokenizeArgs(args)

	for i := 0; i < len(tokens); i++ {
		raw := strings.TrimSpace(tokens[i])
		raw = strings.TrimRight(raw, ",;")

		// Check if argument is a pure number (e.g. positional mandant argument: "3" or "2")
		var numVal int
		if _, err := fmt.Sscanf(raw, "%d", &numVal); err == nil && numVal >= 0 && !strings.HasPrefix(raw, "-") && !strings.HasPrefix(raw, "/") {
			// If we get a pure number argument and mandant isn't set yet, treat as Mandant
			if ov.Mandant == nil && numVal > 0 && numVal < 1000 {
				mVal := numVal
				ov.Mandant = &mVal
				log.Printf("[CLI] Parsed positional numeric arg as Mandant: %d", mVal)
				continue
			} else if ov.Port == nil && numVal >= 1000 {
				pVal := numVal
				ov.Port = &pVal
				log.Printf("[CLI] Parsed positional numeric arg as Port: %d", pVal)
				continue
			}
		}

		trimmed := strings.TrimLeft(raw, "-/")
		if trimmed == "" {
			continue
		}

		if !strings.Contains(trimmed, "=") && strings.Contains(trimmed, ":") {
			trimmed = strings.Replace(trimmed, ":", "=", 1)
		}

		var key, valStr string
		if strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			key = strings.ToLower(strings.TrimSpace(parts[0]))
			valStr = strings.TrimRight(strings.TrimSpace(parts[1]), ",;")
		} else {
			key = strings.ToLower(strings.TrimSpace(trimmed))
			if i+1 < len(tokens) {
				nextRaw := strings.TrimRight(strings.TrimSpace(tokens[i+1]), ",;")
				if !strings.HasPrefix(nextRaw, "-") && !strings.HasPrefix(nextRaw, "/") && !strings.Contains(nextRaw, "=") {
					valStr = nextRaw
					i++
				}
			}
		}

		if key == "" {
			continue
		}

		switch key {
		case "port", "p":
			if pVal, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil && pVal > 0 {
				ov.Port = &pVal
				log.Printf("[CLI] Parsed Port: %d", pVal)
			}
		case "mandant", "m", "mandantennummer":
			if mVal, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil && mVal >= 0 {
				ov.Mandant = &mVal
				log.Printf("[CLI] Parsed Mandant: %d", mVal)
			}
		case "mode":
			s := strings.ToLower(valStr)
			ov.Mode = &s
			log.Printf("[CLI] Parsed Mode: %s", s)
		case "engine", "db_engine", "dbengine":
			s := strings.ToLower(valStr)
			ov.DBEngine = &s
			log.Printf("[CLI] Parsed DBEngine: %s", s)
		case "test":
			var tVal int
			if valStr == "1" || strings.ToLower(valStr) == "true" {
				tVal = 1
				ov.Test = &tVal
			} else if valStr == "0" || strings.ToLower(valStr) == "false" {
				tVal = 0
				ov.Test = &tVal
			} else if parsed, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
				tVal = parsed
				ov.Test = &tVal
			}
			if ov.Test != nil {
				log.Printf("[CLI] Parsed Test: %d", *ov.Test)
			}
		case "waittime", "wait_time", "wait":
			s := strings.TrimSpace(valStr)
			ov.WaitTimeStr = &s
			log.Printf("[CLI] Parsed WaitTime: %s", s)
		case "launcher-port", "launcherport", "lport":
			if lpVal, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil && lpVal > 0 {
				ov.LauncherPort = &lpVal
				log.Printf("[CLI] Parsed LauncherPort: %d", lpVal)
			}
		case "language", "lang", "sprache", "l":
			s := strings.ToLower(strings.TrimSpace(valStr))
			s = strings.Trim(s, " \t\r\n,;")
			if s != "" {
				ov.Language = &s
				log.Printf("[CLI] Parsed Language: %s", s)
			}
		}
	}

	return ov
}

