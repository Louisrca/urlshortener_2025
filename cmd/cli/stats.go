package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"

	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

//shortCodeFlag stocke la valeur du flag --code

var shortCodeFlag = flag.String("code", "", "Le code court de l'URL pour laquelle récupérer les statistiques")

// StatsCmd représente la commande 'stats'
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Valider que le flag --code a été fourni.

		shortCodeFlag, err := cmd.Flags().GetString("code")
		if err != nil {
			log.Fatalf("Erreur lors de la lecture du flag --code : %v", err)
		}

		if shortCodeFlag == "" {
			fmt.Fprintln(os.Stderr, "ERREUR : le flag --code est requis.")
			os.Exit(1)
		}

		//Charger la configuration chargée globalement via cmd.cfg

		cfg := cmd2.Cfg

		//Initialiser la connexion à la BDD.
		// log.Fatalf si erreur

		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("WARN: Échec de la fermeture de la connexion à la base de données: %v", err)
			}
		}()

		// Initialiser les repositories et services nécessaires NewLinkRepository & NewLinkService
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Appeler GetLinkStats pour récupérer le lien et ses statistiques.

		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Fprintf(os.Stderr, "ERREUR : Aucun lien trouvé pour le code court \"%s\".\n", shortCodeFlag)
				os.Exit(1)
			} else {
				log.Fatalf("FATAL: Échec de la récupération des statistiques du lien: %v", err)
			}
		}

		fmt.Printf("Statistiques pour le code court: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

// init() s'exécute automatiquement lors de l'importation du package.
// Il est utilisé pour définir les flags que cette commande accepte.
func init() {
	//Définir le flag --code pour la commande stats.
	StatsCmd.Flags().String("code", "", "Le code court de l'URL pour laquelle récupérer les statistiques")

	// Marquer le flag comme requis
	StatsCmd.MarkFlagRequired("code")

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(StatsCmd)

}
