package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/kafka"
	"go.uber.org/zap"
)

type NotificationService struct {
	cfg    config.Config
	log    *zap.Logger
	client *http.Client
}

func NewNotificationService(cfg config.Config, log *zap.Logger) *NotificationService {
	return &NotificationService{
		cfg: cfg,
		log: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *NotificationService) ConsumeAccessEvents(ctx context.Context, consumer *kafka.Consumer) error {
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			return err
		}
		var event AccessEventMessage
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			s.log.Error("decode access event", zap.Error(err))
			continue
		}
		if err := s.SendTelegram(ctx, event); err != nil {
			s.log.Error("send telegram", zap.Error(err))
			continue
		}
		s.log.Info("telegram notification sent", zap.String("event_id", event.EventID.String()))
	}
}

func (s *NotificationService) SendTelegram(ctx context.Context, event AccessEventMessage) error {
	if s.cfg.Telegram.BotToken == "" || s.cfg.Telegram.ChatID == "" {
		s.log.Info("telegram disabled; set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID to enable")
		return nil
	}
	payload := map[string]string{
		"chat_id": s.cfg.Telegram.ChatID,
		"text":    "AIPass " + string(event.EventType) + " " + string(event.Decision) + " user=" + event.UserID.String(),
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+s.cfg.Telegram.BotToken+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return ErrInvalidInput
	}
	return nil
}
