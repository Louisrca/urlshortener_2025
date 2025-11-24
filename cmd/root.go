package cmd

import (
	"log"
"github.com/axellelanca/urlshortener/internal/cli"
	"github.com/axellelanca/urlshortener/internal/config"
	"github.com/spf13/cobra"
)

// Cfg : configuration globale accessible à toutes les commandes Cobra.
var Cfg *config.Config

// rootCmd représente la commande de base lorsque l'on appelle l'application sans sous-commande.
// C'est le point d'entrée principal pour Cobra.

// Execute est le point d'entrée principal pour l'application Cobra.
var RootCmd = &cobra.Command{
	Use:   "url-shortener",
	Short: "Un service de raccourcissement d'URLs avec API REST et CLI",
	Long: `'url-shortener' est une application complète pour gérer des URLs courtes.
Elle inclut un serveur API pour le raccourcissement et la redirection,
ainsi qu'une interface en ligne de commande pour l'administration. 

Utilisez 'url-shortener [command] --help' pour plus d'informations sur une commande.`,
}

// Execute est appelé depuis main.go
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalf("Erreur lors de l'exécution : %v", err)
	}
}

// init() est une fonction spéciale de Go qui s'exécute automatiquement
// avant la fonction main(). Elle est utilisée ici pour initialiser Cobra
// et ajouter toutes les sous-commandes.
func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig charge la configuration via Viper + ton système custom config.LoadConfig()
func initConfig() {
	var err error

	// LoadConfig() doit lire la config.yaml via Viper
	Cfg, err = config.LoadConfig()
	if err != nil {
		// Si LoadConfig() gère les erreurs critiques, on ne quitte pas ici.
		log.Printf("Attention: Problème lors du chargement de la configuration: %v. Utilisation éventuelle des valeurs par défaut.", err)
	}
}
