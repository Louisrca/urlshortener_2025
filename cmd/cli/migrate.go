package cli

import (
	"fmt"
	"log"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MigrateCmd représente la commande 'migrate'
var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Exécute les migrations de la base de données pour créer ou mettre à jour les tables.",
	Long: `Cette commande se connecte à la base de données configurée (SQLite)
et exécute les migrations automatiques de GORM pour créer les tables 'links' et 'clicks'
basées sur les modèles Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Chargement de la configuration globale
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatal("FATAL : La configuration n'a pas été chargée correctement.")
		}

		// Connexion à la base de données SQLite via GORM
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

		// Migration automatique des modèles GORM
		err = db.AutoMigrate(&models.Link{}, &models.Click{})
		if err != nil {
			log.Fatalf("FATAL : Échec de l'exécution des migrations : %v", err)
		}

		// Succès
		fmt.Println("Migrations de la base de données exécutées avec succès.")
	},
}

func init() {
	cmd2.RootCmd.AddCommand(MigrateCmd)
}
