package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/celestiaorg/torch/config"
	handlers "github.com/celestiaorg/torch/pkg/http"
	"github.com/celestiaorg/torch/pkg/k8s"
)

func init() {
	setupLogging()
}

func setupLogging() {
	// Set the default log level
	log.SetLevel(log.InfoLevel)

	// Set the custom formatter
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			directory := path.Base(path.Dir(f.File))
			return "", directory + "/" + filename + ":" + strconv.Itoa(f.Line)
		},
	})

	// Enable reporting the file and line
	log.SetReportCaller(true)

	// Read the LOG_LEVEL environment variable
	logLevel := os.Getenv("LOG_LEVEL")

	// Adjust the log level based on the environment variable
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		// If LOG_LEVEL is not set or has an unrecognized value, use the default Info level
		log.Warn("LOG_LEVEL not defined in the env vars, using default 'info'")
	}

	log.Info("LOG_LEVEL: ", log.GetLevel())
}

// ParseFlags parses the command-line flags and reads the configuration file.
func ParseFlags() config.MutualPeersConfig {
	// Define the flag for the configuration file path
	configFile := flag.String("config-file", "", "Path to the configuration file")

	// Parse the flags
	flag.Parse()

	// Read the configuration file
	file, err := os.ReadFile(*configFile)
	if err != nil {
		log.Error("Config file doesn't exist...", err)
	}

	cfg := config.MutualPeersConfig{}

	// Unmarshal the YAML into a struct
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Error("Cannot unmarshal the config file...", err)
	}

	return cfg
}

func PrintName() {
	torch := `
================================ 
  __________  ____  ________  __
 /_  __/ __ \/ __ \/ ____/ / / /
  / / / / / / /_/ / /   / /_/ / 
 / / / /_/ / _, _/ /___/ __  /  
/_/  \____/_/ |_|\____/_/ /_/
================================
`
	fmt.Println(torch)
}

func main() {
	PrintName()
	// Parse the command-line flags and read the configuration file
	log.Info("Running on namespace: ", k8s.GetCurrentNamespace())
	cfg := ParseFlags()

	handlers.Run(cfg)
}
