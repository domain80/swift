package main

import (
	"embed"
	"log"
	"os"
	"swift/node"
)

//go:embed ui
var UI_DIR embed.FS

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	node := node.NewNode(infoLog, errorLog, UI_DIR)
	node.Start()
}
