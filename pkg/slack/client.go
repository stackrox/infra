// Package slack handles interfacing with the Slack API.
package slack

import (
	"sync"

	"github.com/slack-go/slack"
	"github.com/stackrox/infra/pkg/config"
	"github.com/stackrox/infra/pkg/logging"
)

// Slacker represents a type that can interact with the Slack API.
type Slacker interface {
	PostMessage(options ...slack.MsgOption) error
	PostMessageToUser(user *slack.User, options ...slack.MsgOption) error
	LookupUser(email string) (*slack.User, bool)
}

var (
	log = logging.CreateProductionLogger()

	_ Slacker = (*slackClient)(nil)
	_ Slacker = (*disabledSlack)(nil)
)

type slackClient struct {
	client     *slack.Client
	channelID  string
	emailCache map[string]*slack.User
	lock       sync.RWMutex
}

type disabledSlack struct{}

func (s disabledSlack) PostMessage(_ ...slack.MsgOption) error {
	return nil
}

func (s disabledSlack) PostMessageToUser(_ *slack.User, _ ...slack.MsgOption) error {
	return nil
}
func (s disabledSlack) LookupUser(_ string) (*slack.User, bool) {
	return &slack.User{}, false
}

// New creates a new Slack client that uses the given token for
// authentication.
func New(cfg *config.SlackConfig) (Slacker, error) {
	// If the config was missing a Slack configuration, disable the integration
	// altogether.
	if cfg == nil {
		log.Log(logging.INFO, "disabling Slack integration due to missing configuration")
		return &disabledSlack{}, nil
	}

	client := &slackClient{
		client:     slack.New(cfg.Token),
		channelID:  cfg.Channel,
		emailCache: make(map[string]*slack.User),
	}

	log.Log(logging.INFO, "enabled Slack integration")

	return client, nil
}

func (s *slackClient) LookupUser(email string) (*slack.User, bool) {
	s.lock.RLock()
	user, found := s.emailCache[email]
	if found {
		s.lock.RUnlock()
		return user, found
	}
	s.lock.RUnlock()

	user, err := s.client.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "users_not_found" {
			log.Log(logging.DEBUG, "slack user not found by email", "email", email)
		} else {
			log.Log(logging.WARN, "slack generic get user by email error", "email", email, "error", err)
		}
		return nil, false
	}

	s.lock.RLock()
	defer s.lock.RUnlock()
	s.emailCache[user.Profile.Email] = user
	return user, true
}

func (s *slackClient) PostMessage(options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(s.channelID, options...)
	return err
}

func (s *slackClient) PostMessageToUser(user *slack.User, options ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(user.ID, options...)
	return err
}
