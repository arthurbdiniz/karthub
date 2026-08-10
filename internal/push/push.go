package push

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/karthub/karthub/internal/repository"
)

type Service struct {
	repo       *repository.PushRepository
	vapidPub   string
	vapidPriv  string
	subscriber string
}

func NewService(repo *repository.PushRepository, vapidPub, vapidPriv, subscriber string) *Service {
	return &Service{
		repo:       repo,
		vapidPub:   vapidPub,
		vapidPriv:  vapidPriv,
		subscriber: subscriber,
	}
}

func (s *Service) VAPIDPublicKey() string {
	return s.vapidPub
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
}

func (s *Service) SendAll(ctx context.Context, payload Payload) {
	subs, err := s.repo.ListAll(ctx)
	if err != nil {
		slog.Error("listing push subscriptions", "error", err)
		return
	}
	s.sendToSubs(subs, payload)
}

func (s *Service) SendToEvent(ctx context.Context, eventID int64, payload Payload) {
	subs, err := s.repo.ListByEventID(ctx, eventID)
	if err != nil {
		slog.Error("listing push subscriptions by event", "error", err)
		return
	}
	s.sendToSubs(subs, payload)
}

func (s *Service) sendToSubs(subs []repository.PushSubscription, payload Payload) {
	data, _ := json.Marshal(payload)

	for _, sub := range subs {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(data, subscription, &webpush.Options{
			Subscriber:      s.subscriber,
			VAPIDPublicKey:  s.vapidPub,
			VAPIDPrivateKey: s.vapidPriv,
			TTL:             60,
		})
		if err != nil {
			slog.Error("sending push notification", "error", err, "endpoint", sub.Endpoint)
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
			_ = s.repo.Delete(context.Background(), sub.Endpoint)
		}
	}
}
