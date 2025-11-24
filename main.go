package main

import (
	"github.com/axellelanca/urlshortener/cmd"
	_ "github.com/axellelanca/urlshortener/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	_ "github.com/axellelanca/urlshortener/cmd/server" // Importe le package 'server' pour que ses init() soient exécutés
)

func main() {
	cmd.Execute()
}
