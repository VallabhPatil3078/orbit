package main

import "github.com/VallabhPatil3078/orbit/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
