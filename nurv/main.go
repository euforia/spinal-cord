package main

import (
	"flag"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/nurv/nurvs"
	"os"
)

func PrintVersion() {
	if PRE_RELEASE_VERSION == "" {
		fmt.Println(VERSION)
	} else {
		fmt.Printf("%s-%s\n", VERSION, PRE_RELEASE_VERSION)
	}
}

func initFlags() (*logging.Logger, *config.NurvConfig) {

	var (
		logger       *logging.Logger = logging.NewStdLogger()
		nurvType                     = flag.String("type", "", "Type of nurv to spawn.  e.g. amqp")
		logLevel                     = flag.String("log-level", "info", "Log level")
		cfgFile                      = flag.String("config", "", "Path to configuration file")
		printVersion                 = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if err := logger.SetLogLevel(*logLevel); err != nil {
		logger.Error.Fatalf("%s\n", err)
	}

	if *printVersion {
		PrintVersion()
		os.Exit(0)
	}

	if *cfgFile == "" {
		fmt.Printf("\n nurv [options]\n\n")
		flag.PrintDefaults()
		fmt.Printf("\n Config file not specified! (-config)\n\n")
		os.Exit(1)
	}

	if *nurvType == "" {
		fmt.Printf("\n nurv [options]\n\n")
		flag.PrintDefaults()
		fmt.Printf("\n Nurv type not specified! (-type)\n\n")
		os.Exit(1)
	}

	cfg, err := config.LoadNurvConfigFromFile(*nurvType, *cfgFile)
	if err != nil {
		logger.Error.Fatalf("Could not load config: %s %s; %s\n", nurvType, cfgFile, err)
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = *logLevel
	}
	return logger, cfg
}

func main() {
	LOGGER, CONFIG := initFlags()

	amqpNurv, err := nurvs.LoadNurv(CONFIG, LOGGER)
	if err != nil {
		LOGGER.Error.Fatalf("Failed to load nurv: %s\n", err)
	}

	if err := amqpNurv.Start(); err != nil {
		LOGGER.Error.Fatalf("Failed to start nurv: %s", err)
	}

	/* block forever */
	quitChan := make(chan int)
	for {
		q := <-quitChan
		LOGGER.Info.Printf("Quitting... %d\n", q)
		amqpNurv.Stop()
		break
	}
}
