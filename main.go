package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// AppConfig - Application-wide configuration handler structure
type AppConfig struct {
	Application struct {
		LogFile string `json:"LogFile"`
		OutFile string `json:"OutFile"`
	} `json:"application"`
	Server map[string]*struct {
		HostType      string   `json:"HostType"`
		DestHost      string   `json:"DestHost"`
		DestPort      int      `json:"DestPort"`
		DestNamespace string   `json:"DestNamespace"`
		HostUser      string   `json:"HostUser"`
		HostPass      string   `json:"HostPass"`
		SSHPrivateKey string   `json:"SSHPrivateKey"`
		ExecCommands  []string `json:"ExecCommands"`
	} `json:"servers"`
}

// AppResults - Structure of data for async multithreaded append
type AppResults struct {
	Metadata struct {
		ExecutionTime int64  `json:"ExecutionTime"`
		ExecutionHost string `json:"ExecutionHost"`
	} `json:"metadata"`
	Server map[string]map[string]resultsServerOut `json:"servers"`
}
type resultsServerOut struct {
	ExecStdout string `json:"stdout"`
	ExecStderr string `json:"stderr"`
}

var (
	// ApplicationConfig - Application-wide main settings
	ApplicationConfig *AppConfig
	// ApplicationLogger - Application-wide logger settings
	ApplicationLogger *log.Logger
	// ApplicationResult - Application-wide to-write data sctucture
	ApplicationResult AppResults
)

func coreLoadConfiguration(file string) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Output(3, "Configuration file not exist, create one first.")
			os.Exit(1)
		} else {
			log.Fatalln(err)
		}
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&ApplicationConfig)
	log.Output(3, "Configuration loaded.")

	if ApplicationConfig.Application.LogFile != "" {
		logFile, err := os.OpenFile(ApplicationConfig.Application.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		ApplicationLogger = log.New(logFile, "", log.Flags())
	} else {
		ApplicationLogger = log.New(os.Stdout, "", log.Flags())
	}
}

func coreWorkersRun() {
	ApplicationLogger.Output(3, "[DEBUG] Working on server map")
	var waitGroup sync.WaitGroup
	for current := range ApplicationConfig.Server {
		if ApplicationConfig.Server[current].HostType == "ssh2" {
			ApplicationLogger.Output(3, "[DEBUG] Executing thread Secure Shell ("+current+")")
			waitGroup.Add(1)
			go SSHWorker(current, &waitGroup)
		} else {
			ApplicationLogger.Output(3, "[WARNING] Could not execute thread for server "+current+": type \""+ApplicationConfig.Server[current].HostType+"\" not found.")
		}
	}
	waitGroup.Wait()
}

// ExecResultSave - Saves results of command execution to ApplicationResult.
func ExecResultSave(server string, stdin string, stdout string, stderr string) {
	// Create map if not exists.
	if len(ApplicationResult.Server) == 0 {
		ApplicationResult.Server = make(map[string]map[string]resultsServerOut)
		ApplicationResult.Server[server] = make(map[string]resultsServerOut)
	}
	ApplicationResult.Server[server][stdin] = resultsServerOut{stdout, stderr}
}

// coreWriteOutputToFile - Write execution results to file
func coreWriteOutputToFile() {
	outFile, err := os.OpenFile(ApplicationConfig.Application.OutFile, os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	jsonEncoder := json.NewEncoder(outFile)
	jsonEncoder.Encode(ApplicationResult)

}
func main() {
	timeStart := time.Now()
	log.Println("Initializing application RFetch...")
	coreLoadConfiguration("config.json")
	coreWorkersRun()
	ApplicationResult.Metadata.ExecutionTime = time.Now().Unix()
	ApplicationResult.Metadata.ExecutionHost, _ = os.Hostname()
	coreWriteOutputToFile()
	ApplicationLogger.Output(3, "RFetch finished in "+time.Since(timeStart).String())
	os.Exit(0)
}
