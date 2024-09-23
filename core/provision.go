package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nkeys"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func ErrAuthLogin(err error) error { return fmt.Errorf("auth error: %w", err) }
func ErrKeyFormat(err error) error { return fmt.Errorf("private key format error: %w", err) }

type provisioner struct {
	session *session
	http    *HttpClient
}

func newProvisioner(session *session, httpClient *HttpClient) *provisioner {

	return &provisioner{session: session, http: httpClient}
}

func (p *provisioner) openSession(ctx context.Context) error {

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
		err = p.http.Get(ctx, fmt.Sprintf("api/agents/%s/login?pubkey=%s", p.session.bee, pubKey), &data)
		if err != nil && isEConnRefused(err) {

			select {
			case <-ctx.Done():
				return errors.New("open session cancelled")
			case <-time.After(2 * time.Second):
				if counter == 0 {
					log.Warn("Hive not reachable. Trying to connect...")
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

	p.session.publicKey = pubKey
	p.session.hub = connectResult.Hub
	p.session.SetBusToken(connectResult.BusToken)
	p.session.SetApiToken(connectResult.ApiToken)
	p.session.busAddresses = connectResult.Servers

	return nil
}

func isEConnRefused(err error) bool {
	var netError *os.SyscallError
	if errors.As(err, &netError) {
		return netError.Syscall == "connect" && netError.Err == unix.ECONNREFUSED
	}
	return false
}
