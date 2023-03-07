// Package slack handles interfacing with the Slack API.
package slack

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/stackrox/infra/config"
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

	log.Printf("[INFO] Enabled Slack integration")

	return client, nil
}

func (s *slackClient) LookupUser(email string) (slack.User, bool) {
	s.lock.RLock()
	user, found := s.emailCache[email]
	if found {
		s.lock.RUnlock()
		return user, found
	}
	s.lock.RUnlock()

	log.Printf("[DEBUG] Lookup user: %s", email)
	users, err := s.client.GetUserByEmail(email)
	if err != nil {
		log.Printf("[WARN] Lookup user error: %s, %v", email, err)
		return nil, false
	}
	log.Printf("[DEBUG] Found user: %s", email)

	s.lock.RLock()
	defer s.lock.RUnlock()
	s.emailCache[user.Profile.Email] = user
	return user, true
}

func (s *slackClient) PostMessage(options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(s.channelID, options...)
	return err
}

func (s *slackClient) PostMessageToUser(user slack.User, options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(user.ID, options...)
	return err
}
