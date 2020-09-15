package version

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

var (
	Version   = "1.0.0"
	GitSHA    = "Not provided"
	BuildTime = "Not provided"

	showVersion bool
)

func Print() {
	if !flag.Parsed() {
		flag.Parse()
	}
	arg := os.Args[0]
	arg = strings.Replace(arg, "\\", "/", -1)
	_, exec := path.Split(arg)
	versionString := fmt.Sprintf("%s version: %s.%s, build: %v", exec, Version, GitSHA, BuildTime)

	if showVersion {
		fmt.Printf("%v\n", versionString)
		os.Exit(0)
	}

	fmt.Printf("%v, pid: %v\n", versionString, os.Getpid())
}

func init() {
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
}
