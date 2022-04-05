package main

import (	
	"flag"	
	"fmt"
	"os"	
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	
	"github.com/veriak/minion/config"	
	apiServer "github.com/veriak/minion/api"
)

var (	
	configFile     string
	cert           string
	key            string
	addr           string
	metricsAddr    string
	verbosityLevel int
)

func showHelp() {
	fmt.Printf("Usage:%s {params}\n", os.Args[0])
	fmt.Println("      -c {config file}")
	fmt.Println("      -cert {cert file}")
	fmt.Println("      -key {key file}")
	fmt.Println("      -a {listen addr}")
	fmt.Println("      -h (show help info)")
	fmt.Println("      -v {0-10} (verbosity level, default 0)")	
}

func parse() bool {
	flag.StringVar(&configFile, "c", "config.yml", "config file")
	flag.StringVar(&cert, "cert", "", "cert file")
	flag.StringVar(&key, "key", "", "key file")
	flag.StringVar(&addr, "a", "localhost", "address to use")
	flag.IntVar(&verbosityLevel, "v", -1, "verbosity level, higher value - more logs")		
	help := flag.Bool("h", false, "help info")
	flag.Parse()	
	
	if !config.Load(configFile) {
		return false
	}

	if *help {
		return false
	}
	return true
}

func main() {
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05-0700"
	
	if !parse() {
		showHelp()
		os.Exit(-1)
	}
	
	if verbosityLevel < 0 {
		verbosityLevel = config.Get().Log.Level
	}
	
	zerolog.SetGlobalLevel(zerolog.Level(verbosityLevel))	

	serv := apiServer.NewAppServer(config.Get())
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT)
	select {
	case err := <-serv.ListenAndServe():
		panic(err)
	case <-sigCh:
		//cleanup()
		fmt.Println("Shutdown service...")
		os.Exit(1)
	}	
}
