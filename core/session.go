package core

type session struct {
	bee          string
	hub          string
	publicKey    string
	seed         ISecureString
	busToken     ISecureString
	apiToken     ISecureString
	busAddresses []string
	hiveAddress  string
}

func newSession(bee, seed, hiveAddress string) *session {
	s := &session{bee: bee, hiveAddress: hiveAddress}
	s.SetSeed(seed)
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
