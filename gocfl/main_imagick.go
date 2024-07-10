//go:build imagick

package main

import (
	"github.com/ocfl-archive/gocfl/v2/gocfl/cmd"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func main() {
	imagick.Initialize()
	defer imagick.Terminate()
	cmd.Execute()

}
