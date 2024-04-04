package main

import (
	"github.com/je4/gocfl/v2/gocfl/cmd"
	"os"
)

func init() {
	os.Setenv("SIEGFRIED_HOME", "c:/temp")
}

func main() {
	cmd.Execute()
}
