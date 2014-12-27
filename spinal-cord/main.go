package main

import (
	"flag"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/reactor"
	spinalio "github.com/euforia/spinal-cord/spinal-cord/io"
	"github.com/euforia/spinal-cord/web"

	"os"
)

func PrintVersion() {
	if PRE_RELEASE_VERSION == "" {
		fmt.Println(VERSION)
	} else {
		fmt.Printf("%s-%s\n", VERSION, PRE_RELEASE_VERSION)
	}
}

func initFlags() (*logging.Logger, string, *config.TaskWorkerConfig) {
	var (
		printVersion = flag.Bool("version", false, "Show version")
		logLevel     = flag.String("log-level", "trace", "Log level")

		workerMode                 = flag.Bool("worker", false, "Start in worker mode")
		workerHandlersDir          = flag.String("handlers-dir", "data", "Directory to store handlers.")
		workerSpinalCordUri string = "tcp://127.0.0.1:44444"

		configFile = flag.String("config", "spinal-cord.toml", "Configuration file")
	)
	flag.Parse()

	logger := logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	if *printVersion {
		PrintVersion()
		os.Exit(0)
	}

	err := logger.SetLogLevel(*logLevel)
	if err != nil {
		logger.Error.Fatalf("%s\n", err)
	}

	if *workerMode {
		if spUri := flag.Arg(flag.NArg() - 1); spUri != "" {
			workerSpinalCordUri = spUri
		}
	}
	return logger, *configFile, &config.TaskWorkerConfig{*workerMode, *workerHandlersDir, workerSpinalCordUri}
}

func startTaskWorker(cfg *config.TaskWorkerConfig, logger *logging.Logger) {
	worker, err := reactor.NewTaskWorker(cfg.SpinalCordUri, cfg.HandlersDir, logger)
	if err != nil {
		logger.Error.Fatalf("Failed to instantiate task worker: %s\n", err)
	}
	worker.Start()
	/* Read results and potentially write out to log file */
	for {
		workResult := <-worker.Results
		if v, ok := workResult["error"]; ok {
			logger.Error.Printf("RESULT ERROR: %s\n", v)
		}
		logger.Info.Printf("Result for handler: %s\n", workResult["handler"])
	}
}

func startSubreactor(cfg *config.SpinalCordConfig, logger *logging.Logger) {

	sreactor, err := reactor.NewSubReactor(cfg, logger)
	if err != nil {
		logger.Error.Printf("FAILED to start reactor: %s\n", err)
		return
	}
	sreactor.Start(cfg.Reactor.CreateSamples) // true = create samples
}

func loadConfig(cfgfile string, logger *logging.Logger) *config.SpinalCordConfig {
	cfg, err := config.LoadConfigFromTomlFile(cfgfile)
	if err != nil {
		logger.Error.Fatalf("%s", err)
	}
	if err := cfg.Validate(); err != nil {
		logger.Error.Fatalf("%s\n", err)
	}
	return cfg
}

func main() {

	LOGGER, CONFIG_FILE, WORKER_CONFIG := initFlags()

	if WORKER_CONFIG.WorkerMode {
		startTaskWorker(WORKER_CONFIG, LOGGER)
	} else {

		cfg := loadConfig(CONFIG_FILE, LOGGER)

		ioMgr := spinalio.NewIOManager(LOGGER)
		ioMgr.LoadIO(cfg, true)

		coreWebSvc := web.NewCoreWebService(cfg, LOGGER)
		coreWebSvc.Start()

		startSubreactor(cfg, LOGGER)
	}
}
