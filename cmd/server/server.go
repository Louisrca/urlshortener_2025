package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/api"
	"github.com/axellelanca/urlshortener/internal/config"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/monitor"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/axellelanca/urlshortener/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// RunServerCmd représente la commande 'run-server' de Cobra.
// C'est le point d'entrée pour lancer le serveur de l'application.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO : créer une variable qui stock la configuration chargée globalement via cmd.cfg
		// Ne pas oublier la gestion d'erreur et faire un fatalF
		cfg, err := config.GetConfig()
		if err != nil {
			log.Fatalf("FATAL: Échec du chargement de la configuration : %v", err)
		}

		// TODO : Initialiser la connexion à la BDD
		db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données : %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}
		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("WARN: Échec de la fermeture de la base de données : %v", err)
			}
		}()

		// TODO : Initialiser la connexion à la bBDD
		log.Println("Connexion à la base de données établie")

		// TODO : Initialiser les repositories.
		log.Println("Initialisation des repositories...")

		// Créez des instances de GormLinkRepository et GormClickRepository.
		linkRepo := repository.NewGormLinkRepository(db)
		clickRepo := repository.NewGormClickRepository(db)

		// Laissez le log
		log.Println("Repositories initialisés.")

		// TODO : Initialiser les services métiers.
		// Créez des instances de LinkService et ClickService, en leur passant les repositories nécessaires.
		log.Println("Initialisation des services métiers...")
		linkService := services.NewLinkService(linkRepo)
		clickService := services.NewClickService(clickRepo)

		// Laissez le log
		log.Println("Services métiers initialisés.")

		// TODO : Initialiser le channel ClickEventsChannel (api/handlers) des événements de clic et lancer les workers (StartClickWorkers).
		// Le channel est bufferisé avec la taille configurée.
		// Passez le channel et le clickRepo aux workers.
		clickEventsChannel := make(chan models.ClickEvent, cfg.Workers.ClickWorkerBufferSize)
		workers.StartClickWorkers(clickEventsChannel, clickRepo, cfg.Workers.ClickWorkerCount)

		// TODO : Remplacer les XXX par les bonnes variables
		log.Printf("Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			cfg.Workers.ClickWorkerBufferSize, cfg.Workers.ClickWorkerCount)

		// TODO : Initialiser et lancer le moniteur d'URLs.
		log.Println("Initialisation du moniteur d'URL...")
		// Utilisez l'intervalle configuré
		monitorInterval := time.Duration(cfg.Workers.UrlMonitorIntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval) // Le moniteur a besoin du linkRepo et de l'interval

		// TODO Lancez le moniteur dans sa propre goroutine.
		go urlMonitor.Start()

		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", monitorInterval)

		// TODO : Configurer le routeur Gin et les handlers API.
		// Passez les services nécessaires aux fonctions de configuration des routes.
		log.Println("Configuration des routes API...")
		router := gin.Default()
		api.ConfigureRoutes(router, linkService, clickService, clickEventsChannel)

		// Pas toucher au log
		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// TODO : Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer.
		// Pensez à logger des ptites informations...
		go func() {
			log.Println("Démarrage du serveur HTTP sur le port ", cfg.Server.Port)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("FATAL: Échec du démarrage du serveur HTTP : %v", err)
			}
		}()

		// Gére l'arrêt propre du serveur (graceful shutdown).
		// TODO Créez un channel pour les signaux OS (SIGINT, SIGTERM), bufferisé à 1.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Attendre Ctrl+C ou signal d'arrêt

		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu.
		<-quit
		log.Println("Signal d'arrêt reçu. Arrêt du serveur...")

		// Arrêt propre du serveur HTTP avec un timeout.
		log.Println("Arrêt en cours... Donnez un peu de temps aux workers pour finir.")
		time.Sleep(5 * time.Second)

		log.Println("Serveur arrêté proprement.")
	},
}

func init() {
	// TODO : ajouter la commande
	cmd2.RootCmd.AddCommand(RunServerCmd)
}
