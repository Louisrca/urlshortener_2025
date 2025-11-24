package cmd

import (
	"log"

	"github.com/axellelanca/urlshortener/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Cfg est la configuration globale partagée entre toutes les commandes.
	Cfg *config.Config
)

// RootCmd est la commande racine de l'application CLI.
var RootCmd = &cobra.Command{
	Use:   "url-shortener",
	Short: "Un service de raccourcissement d'URLs avec API REST et CLI",
	Long: `url-shortener est une application complète de raccourcissement d'URL.
Elle offre une API REST, un serveur web, et des commandes CLI pour gérer les URLs.`,
}

// Execute lance l'application CLI.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalf("Erreur lors de l'exécution : %v", err)
	}
}

// init est appelée automatiquement avant main().
// Elle initialise la config globale et enregistre les sous-commandes.
func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig charge la configuration depuis config.yaml via Viper.
// Si le fichier est absent ou incorrect, Viper utilisera les valeurs par défaut définies dans LoadConfig().
func initConfig() {
	var err error
	Cfg, err = config.LoadConfig()
	if err != nil {
		log.Printf("⚠️ Problème lors du chargement de la configuration : %v. Valeurs par défaut utilisées.", err)
	}
}
