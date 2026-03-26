package runtime

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
	"github.com/nats-io/nkeys"
	"golang.org/x/sys/unix"
)

type Authenticator interface {
	OpenSession(ctx context.Context) error
}

// TWO STEP AUTHENTICATOR

func ErrAuthLogin(err error) error { return fmt.Errorf("auth error: %w", err) }
func ErrKeyFormat(err error) error { return fmt.Errorf("private key format error: %w", err) }

type twoStepAuthenticator struct {
	session *core.Session
	log     models.Logger
	http    *core.HttpClient
}

func NewTwoStepAuthenticator(session *core.Session, httpClient *core.HttpClient) Authenticator {

	return &twoStepAuthenticator{session: session, log: session.Log, http: httpClient}
}

func (p *twoStepAuthenticator) OpenSession(ctx context.Context) error {

	// Extracts public key from seed
	keys, err := nkeys.FromSeed([]byte(p.session.Seed()))
	if err != nil {
		return ErrKeyFormat(err)
	}
	pubKey, err := keys.PublicKey()
	if err != nil {
		return ErrKeyFormat(err)
	}

	// Sends bee name and public key and gets nonce to probe identity
	var data struct {
		Nonce string `json:"nonce"`
	}
	counter := 0
	for {
		err = p.http.Get(ctx, fmt.Sprintf("api/agents/%s/login?pubkey=%s", p.session.Bee, pubKey), &data)
		if err != nil && isEConnRefused(err) {

			select {
			case <-ctx.Done():
				return errors.New("open session cancelled")
			case <-time.After(2 * time.Second):
				if counter == 0 {
					p.log.Warn("Hive not reachable. Trying to connect...")
				}
				counter += 1
			}
		} else {
			break
		}
	}

	// Signs and encodes received nonce
	signedNonce, err := keys.Sign([]byte(data.Nonce))
	if err != nil {
		return ErrKeyFormat(err) // TODO: check if this is the right error
	}
	encodedSignedNonce := base64.URLEncoding.EncodeToString(signedNonce)

	// Sends signed nonce and gets connection results (tokens, NATS seed servers, etc.)
	var connectResult struct {
		Bee      string   `json:"bee"`
		Hub      string   `json:"hub"`
		BusToken string   `json:"busToken"`
		ApiToken string   `json:"apiToken"`
		Servers  []string `json:"servers"`
	}

	err = p.http.Get(ctx, fmt.Sprintf("api/connect?pubkey=%s&nonce=%s", pubKey, encodedSignedNonce), &connectResult)
	if err != nil {
		return ErrAuthLogin(err)
	}

	p.session.PublicKey = pubKey
	p.session.Hub = connectResult.Hub
	p.session.SetBusToken(connectResult.BusToken)
	p.session.SetApiToken(connectResult.ApiToken)
	p.session.BusAddresses = connectResult.Servers

	return nil
}

func isEConnRefused(err error) bool {
	var netError *os.SyscallError
	if errors.As(err, &netError) {
		return netError.Syscall == "connect" && netError.Err == unix.ECONNREFUSED
	}
	return false
}

// EMBEDDED AUTHENTICATOR

type localAuthenticator struct {
	session *core.Session
}

func NewLocalAuthenticator(session *core.Session) Authenticator {
	return &localAuthenticator{session: session}
}

func (p *localAuthenticator) OpenSession(ctx context.Context) error {
	return nil
}
