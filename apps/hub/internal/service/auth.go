package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/repository"
	"gorm.io/gorm"
)

const deviceCodeTTL = 10 * time.Minute

// ─── AgentTokenService ────────────────────────────────────────────────────────

type AgentTokenService struct {
	db         *gorm.DB
	tokenRepo  domain.AgentTokenRepository
}

func (s *AgentTokenService) Generate(ctx context.Context, orgID, name, createdByID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	plaintext := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256sum(plaintext)

	t := &domain.AgentToken{
		OrgID:       orgID,
		Name:        name,
		TokenHash:   hash,
		CreatedByID: createdByID,
	}
	if err := s.tokenRepo.Create(ctx, t); err != nil {
		return "", err
	}
	return plaintext, nil
}

func (s *AgentTokenService) Validate(ctx context.Context, token string) (string, string, error) {
	hash := sha256sum(token)
	t, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", "", fmt.Errorf("invalid token")
		}
		return "", "", err
	}
	_ = s.tokenRepo.TouchLastUsed(ctx, t.ID, time.Now())
	agentID := ""
	if t.AgentID != nil {
		agentID = *t.AgentID
	}
	return t.OrgID, agentID, nil
}

// ─── DeviceFlowService ────────────────────────────────────────────────────────

type DeviceFlowService struct {
	db          *gorm.DB
	deviceRepo  domain.DeviceCodeRepository
	tokenRepo   domain.AgentTokenRepository
	tokenSvc    *AgentTokenService
	frontendURL string
}

func (s *DeviceFlowService) Initiate(ctx context.Context) (*domain.DeviceCode, error) {
	deviceCode, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	userCode := randomUserCode()
	d := &domain.DeviceCode{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ExpiresAt:  time.Now().Add(deviceCodeTTL),
	}
	return d, s.deviceRepo.Create(ctx, d)
}

func (s *DeviceFlowService) Approve(ctx context.Context, userCode, orgID, userID string) error {
	d, err := s.deviceRepo.GetByUserCode(ctx, userCode)
	if err != nil {
		return err
	}
	if time.Now().After(d.ExpiresAt) {
		return fmt.Errorf("device code expired")
	}
	if d.ApprovedAt != nil {
		return fmt.Errorf("already approved")
	}

	// Generate the agent token via a temporary placeholder createdByID.
	// We use the org context; the approving user's ID isn't threaded here yet
	// but the token is scoped to the org.
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	plaintext := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256sum(plaintext)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := repository.New(tx)

		// Create the agent record now so it shows up in the dashboard immediately.
		agent := &domain.Agent{
			OrgID:  orgID,
			Name:   "pending",
			Status: domain.AgentStatusOffline,
		}
		if err := txRepos.Agent.Create(ctx, agent); err != nil {
			return err
		}

		t := &domain.AgentToken{
			OrgID:       orgID,
			AgentID:     &agent.ID,
			Name:        "paired-agent",
			TokenHash:   hash,
			CreatedByID: userID,
		}
		if err := txRepos.AgentToken.Create(ctx, t); err != nil {
			return err
		}
		return txRepos.DeviceCode.Approve(ctx, d.DeviceCode, orgID, plaintext)
	})
	return err
}

func (s *DeviceFlowService) Poll(ctx context.Context, deviceCode string) (string, bool, error) {
	d, err := s.deviceRepo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		return "", false, err
	}
	if time.Now().After(d.ExpiresAt) {
		return "", false, fmt.Errorf("expired")
	}
	if d.ApprovedAt == nil {
		return "", true, nil // pending
	}
	tok, err := s.deviceRepo.ConsumeToken(ctx, deviceCode)
	if err != nil {
		return "", false, err
	}
	return tok, false, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func sha256sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomUserCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	rand.Read(b)
	var sb strings.Builder
	for i, c := range b {
		if i == 4 {
			sb.WriteByte('-')
		}
		sb.WriteByte(chars[int(c)%len(chars)])
	}
	return sb.String()
}
