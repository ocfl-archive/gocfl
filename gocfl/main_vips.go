//go:build vips

package main

import "github.com/ocfl-archive/gocfl/v2/gocfl/cmd"
import "github.com/davidbyttow/govips/v2/vips"

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()

	cmd.Execute()
}
