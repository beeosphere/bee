package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beeosphere/bee/core/mediary"
)

type HttpClient struct {
	baseUri string
	http    *http.Client
	session *session
}

func NewHttpClient(options *BeeOptions, session *session) *HttpClient {

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	clientWithInterceptor := mediary.Init().
		WithPreconfiguredClient(client).
		AddInterceptors(func(req *http.Request, handler mediary.Handler) (*http.Response, error) {
			// req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session.apiToken.Get()))

			// fmt.Println("INTERCEPTOR...")

			return handler(req)
		}).
		Build()

	return &HttpClient{
		baseUri: options.HiveUri,
		http:    clientWithInterceptor,
		session: session,
	}
}

func (c *HttpClient) Get(ctx context.Context, uri string, data interface{}) error {

	absUrl := uri
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		absUrl = fmt.Sprintf("%s/%s", c.baseUri, strings.TrimPrefix(uri, "/"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, absUrl, nil)

	// req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.session.apiToken.Get()))
	res, err := c.http.Do(req)
	if err != nil {
		return err.(*url.Error).Err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		if data != nil {
			if err = json.Unmarshal(body, data); err != nil {
				return err
			}
		}
		return nil
	}
	// fmt.Println(res.Status)
	return errors.New(string(body))
}

// var baseUri string

// func newHttpClient(options *BeeOptions, session *session) *http.Client {

// 	baseUri = options.ServerUri

// 	var netTransport = &http.Transport{
// 		Dial: (&net.Dialer{
// 			Timeout: 5 * time.Second,
// 		}).Dial,
// 		TLSHandshakeTimeout: 5 * time.Second,
// 	}

// 	client := &http.Client{
// 		Timeout:   time.Second * 10,
// 		Transport: netTransport,
// 	}

// 	clientWithInterceptor := mediary.Init().
// 		WithPreconfiguredClient(client).
// 		AddInterceptors(func(req *http.Request, handler mediary.Handler) (*http.Response, error) {

// 			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", session.PrivateKey()))
// 			return handler(req)
// 		}).
// 		Build()

// 	return clientWithInterceptor
// }

// func dumpInterceptor(req *http.Request, handler mediary.Handler) (*http.Response, error) {

// 	fmt.Println("INTERCEPTOR PRE EXECUTED!!")

// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "TOKEN"))

// 	res, err := handler(req)

// 	fmt.Println("INTERCEPTOR POST EXECUTED!!")
// 	return res, err
// }
