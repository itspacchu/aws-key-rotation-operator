package main

import (
	"flag"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/aws-key-rotation/cmd"
)

const VERSION string = "v1.1.0"

func main() {
	log.Infof("Started aws-key-rotation Operator %s", VERSION)
	log.Info("\\_ maintainer prashant.nandipati@zigram.tech")
	verbose := flag.Bool("v", false, "Enable Verbose")
	flag.StringVar(&cmd.Namespace, "n", "", "Namespace to use")
	flag.StringVar(&cmd.SecretName, "secret-name", "aws-key", "Secret Name to create")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
		log.SetPrefix("[Verbose]")
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
