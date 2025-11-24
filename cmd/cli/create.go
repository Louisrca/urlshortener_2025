package cli

import (
	"fmt"
	"log"
	"net/url"
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie via --url et affiche le code court généré.

Exemple :
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {

		// Lecture du flag --url
		urlStr, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Fatalf("Erreur lors de la lecture du flag --url : %v", err)
		}
		if urlStr == "" {
			fmt.Fprintln(os.Stderr, "ERREUR : le flag --url est requis.")
			os.Exit(1)
		}

		// Validation du format de l’URL
		if _, err := url.ParseRequestURI(urlStr); err != nil {
			fmt.Fprintf(os.Stderr, "ERREUR : URL invalide \"%s\" : %v\n", urlStr, err)
			os.Exit(1)
		}

		// Chargement de la configuration globale
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatal("FATAL : La configuration n'a pas été chargée correctement.")
		}

		// Connexion à la base de données via GORM
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL : Échec de la connexion à la base de données : %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL : Échec d'accès au driver SQL natif : %v", err)
		}
		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("WARN : Échec de la fermeture de la connexion DB : %v", err)
			}
		}()

		// Repositories + Services
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Création du lien court
		link, err := linkService.CreateLink(urlStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERREUR : Échec de la création de l'URL courte : %v\n", err)
			os.Exit(1)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.ShortCode)

		fmt.Println("URL courte créée avec succès ✔️")
		fmt.Printf("Code court : %s\n", link.ShortCode)
		fmt.Printf("URL complète : %s\n", fullShortURL)
	},
}

func init() {
	// Définition du flag --url
	CreateCmd.Flags().String("url", "", "L'URL longue à raccourcir")

	// Marquer le flag comme requis
	CreateCmd.MarkFlagRequired("url")

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(CreateCmd)
}
