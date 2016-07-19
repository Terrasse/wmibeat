// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/paths"
)

type Config struct {
	Period    string `yaml:"period"`
	ConfigDir string `config:"config_dir"`
	Classes   []ClassConfig
}

type ClassConfig struct {
	Class       string   `config:"class"`
	Fields      []string `config:"fields"`
	WhereClause string   `config:"whereclause"`
}

var (
	DefaultConfig = Config{
		Period: "",
		Classes: []ClassConfig{
			ClassConfig{
				Class: "Win32_OperatingSystem",
				Fields: []string{
					"FreePhysicalMemory",
					"FreeSpaceInPagingFiles",
					"FreeVirtualMemory",
					"NumberOfProcesses",
					"NumberOfUsers",
				},
				WhereClause: "",
			},
			ClassConfig{
				Class: "Win32_PerfFormattedData_PerfDisk_LogicalDisk",
				Fields: []string{
					"Name",
					"FreeMegabytes",
					"PercentFreeSpace",
					"CurrentDiskQueueLength",
					"DiskReadsPerSec",
					"DiskWritesPerSec",
					"DiskBytesPerSec",
					"PercentDiskReadTime",
					"PercentDiskWriteTime",
					"PercentDiskTime",
				},
				WhereClause: `Name != "_Total"`,
			},
			ClassConfig{
				Class: "Win32_PerfFormattedData_PerfOS_Memory",
				Fields: []string{
					"CommittedBytes",
					"AvailableBytes",
					"PercentCommittedBytesInUse",
				},
				WhereClause: "",
			},
		},
	}
)

// getConfigFiles returns list of config files.
// In case path is a file, it will be directly returned.
// In case it is a directory, it will fetch all .yml files inside this directory
func getConfigFiles(path string) (configFiles []string, err error) {

	// Check if path is valid file or dir
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Create empty slice for config file list
	configFiles = make([]string, 0)

	if stat.IsDir() {
		files, err := filepath.Glob(path + "/*.yml")

		if err != nil {
			return nil, err
		}

		configFiles = append(configFiles, files...)

	} else {
		// Only 1 config file
		configFiles = append(configFiles, path)
	}

	return configFiles, nil
}

// mergeConfigFiles reads in all config files given by list configFiles and merges them into config
func mergeConfigFiles(configFiles []string, config *Config) error {

	for _, file := range configFiles {
		logp.Info("Additional configs loaded from: %s", file)

		tmpConfig := struct {
			Wmibeat Config
		}{}
		cfgfile.Read(&tmpConfig, file)

		config.Classes = tmpConfig.Wmibeat.Classes
	}

	return nil
}

// Fetches and merges all config files given by configDir. All are put into one config object
func (config *Config) FetchConfigs() error {

	configDir := config.ConfigDir

	// If option not set, do nothing
	if configDir == "" {
		return nil
	}

	// If configDir is relative, consider it relative to the config path
	configDir = paths.Resolve(paths.Config, configDir)

	// Check if optional configDir is set to fetch additional config files
	logp.Info("Additional config files are fetched from: %s", configDir)

	configFiles, err := getConfigFiles(configDir)

	if len(configFiles) > 0 {

		//if we have config files that allows wmi queries to be run, so clear out the default entries
		config.Classes = []ClassConfig{}

		if err != nil {
			log.Fatal("Could not use config_dir of: ", configDir, err)
			return err
		}

		err = mergeConfigFiles(configFiles, config)
		if err != nil {
			log.Fatal("Error merging config files: ", err)
			return err
		}
	}

	if len(config.Classes) == 0 {
		err := errors.New("No classes given. What wmi entries do you want me to watch?")
		log.Fatalf("%v", err)
		return err
	}

	return nil
}
