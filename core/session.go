package core

// log "github.com/sirupsen/logrus"

type session struct {
	bee          string
	hub          string
	publicKey    string
	seed         ISecureString
	busToken     ISecureString
	apiToken     ISecureString
	busAddresses []string
	hiveAddress  string
	configPath   string
	configWatch  bool
}

func newSession(options *AgentOptions) *session {
	s := &session{
		bee: options.Id,
	}
	if options.HiveUri != "" {
		s.hiveAddress = formattedHiveUri(options.HiveUri)
	}
	if options.Source != nil {
		s.configPath = options.Source.Path
		s.configWatch = options.Source.Watch
	}
	if options.Key != "" {
		s.SetSeed(options.Key)
	}
	return s
}

func (s *session) SetSeed(seed string) { s.seed = NewSecureString(seed) }
func (s *session) Seed() string {
	if s.seed != nil {
		return s.seed.Get()
	}
	return ""
}

func (s *session) SetBusToken(busToken string) { s.busToken = NewSecureString(busToken) }
func (s *session) BusToken() string {
	if s.busToken != nil {
		return s.busToken.Get()
	}
	return ""
}

func (s *session) SetApiToken(apiToken string) { s.apiToken = NewSecureString(apiToken) }
func (s *session) ApiToken() string {
	if s.apiToken != nil {
		return s.apiToken.Get()
	}
	return ""
}

func (s *session) IsEmbedded() bool {
	return s.configPath != "" && s.hiveAddress == ""
}
