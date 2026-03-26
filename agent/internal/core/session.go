package core

import (
	"fmt"
	"strings"

	"github.com/beeosphere/bee/agent/models"
)

type Session struct {
	Bee          string
	Hub          string
	PublicKey    string
	BusAddresses []string
	HiveAddress  string
	// ConfigPath   string
	// ConfigWatch  bool
	seed     ISecureString
	busToken ISecureString
	apiToken ISecureString

	Manifest models.AgentManifest

	Log models.Logger
}

func NewSession(options *models.AgentOptions, manifest *models.AgentManifest, logger models.Logger) *Session {
	s := &Session{
		Bee:      options.Id,
		Log:      logger,
		Manifest: *manifest,
	}
	if options.HiveUri != "" {
		s.HiveAddress = formattedHiveUri(options.HiveUri)
	}
	// if options.Source != nil {
	// 	s.ConfigPath = options.Source.Path
	// 	s.ConfigWatch = options.Source.Watch
	// }
	if options.Key != "" {
		s.SetSeed(options.Key)
	}
	return s
}

func (s *Session) SetSeed(seed string) { s.seed = NewSecureString(seed) }
func (s *Session) Seed() string {
	if s.seed != nil {
		return s.seed.Get()
	}
	return ""
}

func (s *Session) SetBusToken(busToken string) { s.busToken = NewSecureString(busToken) }
func (s *Session) BusToken() string {
	if s.busToken != nil {
		return s.busToken.Get()
	}
	return ""
}

func (s *Session) SetApiToken(apiToken string) { s.apiToken = NewSecureString(apiToken) }
func (s *Session) ApiToken() string {
	if s.apiToken != nil {
		return s.apiToken.Get()
	}
	return ""
}

func (s *Session) IsEmbedded() bool {
	return false
	// return s.ConfigPath != "" && s.HiveAddress == ""
}

func formattedHiveUri(hive string) string {
	if strings.HasPrefix(hive, "http") {
		return hive
	}
	return fmt.Sprintf("http://%s", hive) // TODO: http or https...
}
