package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type LinkService struct {
	linkRepo repository.LinkRepository
}

func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{linkRepo: linkRepo}
}

// GenerateShortCode génère un short code sécurisé
func (s *LinkService) GenerateShortCode(length int) (string, error) {
	code := make([]byte, length)

	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}

	return string(code), nil
}

func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {

	const maxRetries = 5
	var shortCode string

	for i := 0; i < maxRetries; i++ {

		code, err := s.GenerateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		// 👇 ICI était le bug : il manquait totalement cet appel !
		_, err = s.linkRepo.GetLinkByShortCode(code)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// code libre → bingo !
			shortCode = code
			break
		}

		if err != nil {
			// Vraie erreur DB
			return nil, fmt.Errorf("database error checking code uniqueness: %w", err)
		}

		// Collision détectée
		log.Printf("Short code '%s' already exists, retrying (%d/%d)...",
			code, i+1, maxRetries)
	}

	// Si jamais aucun code n'a été trouvé
	if shortCode == "" {
		return nil, errors.New("failed to generate a unique short code after several attempts")
	}

	// Création du lien
	link := &models.Link{
		LongURL:   longURL,
		ShortCode: shortCode,
		CreatedAt: time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, fmt.Errorf("failed to create link in database: %w", err)
	}

	return link, nil
}

func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	return link, nil
}

func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch link: %w", err)
	}

	count, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clicks: %w", err)
	}

	return link, count, nil
}
