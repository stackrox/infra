// Package slack handles interfacing with the Slack API.
package slack

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/stackrox/infra/config"
)

const (
	cacheUpdateInterval = time.Hour
)

// Slacker represents a type that can interact with the Slack API.
type Slacker interface {
	PostMessage(options ...slack.MsgOption) error
	PostMessageToUser(user slack.User, options ...slack.MsgOption) error
	LookupUser(email string) (slack.User, bool)
}

var _ Slacker = (*slackClient)(nil)
var _ Slacker = (*disabledSlack)(nil)

type slackClient struct {
	client     *slack.Client
	channelID  string
	emailCache map[string]slack.User
	lock       sync.RWMutex
}

type disabledSlack struct{}

func (s disabledSlack) PostMessage(options ...slack.MsgOption) error {
	return nil
}

func (s disabledSlack) PostMessageToUser(user slack.User, options ...slack.MsgOption) error {
	return nil
}
func (s disabledSlack) LookupUser(email string) (slack.User, bool) {
	return slack.User{}, false
}

// New creates a new Slack client that uses the given token for
// authentication.
func New(cfg *config.SlackConfig) (Slacker, error) {
	// If the config was missing a Slack configuration, disable the integration
	// altogether.
	if cfg == nil {
		log.Printf("[INFO] Disabling Slack integration")
		return &disabledSlack{}, nil
	}

	client := &slackClient{
		client:     slack.New(cfg.Token),
		channelID:  cfg.Channel,
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

func (s *slackClient) PostMessage(options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(s.channelID, options...)
	return err
}

func (s *slackClient) PostMessageToUser(user slack.User, options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(user.ID, options...)
	return err
}

func (s *slackClient) updateUserEmailCache(ctx context.Context) error {
	users, err := s.client.GetUsersContext(ctx)
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
