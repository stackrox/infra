// Package slack handles interfacing with the Slack API.
package slack

import (
	"sync"

	"github.com/slack-go/slack"
	"github.com/stackrox/infra/config"
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

func (s disabledSlack) PostMessage(options ...slack.MsgOption) error {
	return nil
}

func (s disabledSlack) PostMessageToUser(user *slack.User, options ...slack.MsgOption) error {
	return nil
}
func (s disabledSlack) LookupUser(email string) (*slack.User, bool) {
	return &slack.User{}, false
}

// New creates a new Slack client that uses the given token for
// authentication.
func New(cfg *config.SlackConfig) (Slacker, error) {
	// If the config was missing a Slack configuration, disable the integration
	// altogether.
	if cfg == nil {
		log.Infow("disabling Slack integration due to missing configuration")
		return &disabledSlack{}, nil
	}

	client := &slackClient{
		client:     slack.New(cfg.Token),
		channelID:  cfg.Channel,
		emailCache: make(map[string]*slack.User),
	}

	log.Infow("enabled Slack integration")

	return client, nil
}

func (s *slackClient) LookupUser(email string) (*slack.User, bool) {
	// TODO: do we still need all this debug logging?
	log.Debugw("lookup user by email", "email", email)
	s.lock.RLock()
	user, found := s.emailCache[email]
	if found {
		s.lock.RUnlock()
		log.Debugw("cache hit for email", "email", email)
		return user, found
	}
	s.lock.RUnlock()

	log.Debugw("get user by email", "email", email)
	user, err := s.client.GetUserByEmail(email)
	if err != nil {
		log.Warnw("get user error", "email", email, "error", err)
		return nil, false
	}
	log.Debugw("got user for email", "email", email)

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
