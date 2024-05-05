package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/dan13ram/wpokt-oracle/app"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logLevel := strings.ToLower(os.Getenv("LOGGER_LEVEL"))
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	var yamlPath string
	var envPath string
	flag.StringVar(&yamlPath, "yaml", "", "path to yaml file")
	flag.StringVar(&envPath, "env", "", "path to env file")
	flag.Parse()

	var absYamlPath string = ""
	var err error
	if yamlPath != "" {
		absYamlPath, err = filepath.Abs(yamlPath)
		if err != nil {
			log.Fatal("[MAIN] Error getting absolute path for yaml file: ", err)
		}
		log.Debug("[MAIN] Yaml file: ", absYamlPath)
	}

	var absEnvPath string = ""
	if envPath != "" {
		absEnvPath, err = filepath.Abs(envPath)
		if err != nil {
			log.Fatal("[MAIN] Error getting absolute path for env file: ", err)
		}
		log.Debug("[MAIN] Env file: ", absEnvPath)
	}

	app.InitConfig(absYamlPath, absEnvPath)
	app.InitLogger()
	app.InitDB()

	log.Debug("[MAIN] Starting server")

	/*

		pokt.ValidateNetwork()
		eth.ValidateNetwork()

		healthcheck := app.NewHealthCheck()

		serviceHealthMap := make(map[string]models.ServiceHealth)

		if app.Config.HealthCheck.ReadLastHealth {
			if lastHealth, err := healthcheck.FindLastHealth(); err == nil {
				for _, serviceHealth := range lastHealth.ServiceHealths {
					serviceHealthMap[serviceHealth.Name] = serviceHealth
				}
			}
		}

		services := []app.Service{}
		var wg sync.WaitGroup

		for serviceName, NewService := range ServiceFactoryMap {
			health := models.ServiceHealth{}
			if lastHealth, ok := serviceHealthMap[serviceName]; ok {
				health = lastHealth
			}
			services = append(services, NewService(&wg, health))
		}

		services = append(services, app.NewHealthService(healthcheck, &wg))

		healthcheck.SetServices(services)

		wg.Add(len(services))

		for _, service := range services {
			go service.Start()
		}

		log.Info("[MAIN] Server started")

		gracefulStop := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)
		go waitForExitSignals(gracefulStop, done)
		<-done

		log.Debug("[MAIN] Stopping server gracefully")

		for _, service := range services {
			service.Stop()
		}

		wg.Wait()

	*/
	app.DB.Disconnect()
	log.Info("[MAIN] Server stopped")
}

func waitForExitSignals(gracefulStop chan os.Signal, done chan bool) {
	sig := <-gracefulStop
	log.Debug("[MAIN] Caught signal: ", sig)
	done <- true
}
