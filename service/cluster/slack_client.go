package cluster

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

const (
	cacheUpdateInterval = time.Hour
)

// Slacker represents a type that can interact with the Slack API.
type Slacker interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	LookupUser(email string) (slack.User, bool)
}

var _ Slacker = (*slackClient)(nil)

type slackClient struct {
	*slack.Client
	emailCache map[string]slack.User
	lock       sync.RWMutex
}

// NewSlackClient creates a new Slack client that uses the given token for
// authentication.
func NewSlackClient(token string) (Slacker, error) {
	client := &slackClient{
		Client:     slack.New(token),
		emailCache: make(map[string]slack.User),
	}

	// Update the Slack user cache once, manually. If the initial attempt fails, bail out immediately.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.updateUserEmailCache(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to refresh Slack user cache")
	}

	log.Printf("[DEBUG] Fetched %d Slack users", len(client.emailCache))

	// Update the Slack user cache every hour, in the background. If any of these background attempts fail, log the
	// error and move along.
	go client.backgroundUpdateUserEmailCache()

	return client, nil
}

func (s *slackClient) LookupUser(email string) (slack.User, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	user, found := s.emailCache[email]
	return user, found
}

func (s *slackClient) updateUserEmailCache(ctx context.Context) error {
	users, err := s.GetUsersContext(ctx)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	for _, user := range users {
		if user.Profile.Email == "" {
			continue
		}
		s.emailCache[user.Profile.Email] = user
	}

	return nil
}

func (s *slackClient) backgroundUpdateUserEmailCache() {
	for {
		time.Sleep(cacheUpdateInterval)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := s.updateUserEmailCache(ctx); err != nil {
			log.Printf("[ERROR] Failed to refresh Slack user cache: %v", err)
		}
		cancel()
	}
}
