/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	client "github.com/ory/hydra-client-go"
)

const (
	loginChallengeCookieName   = "bdd_test_cookie_login_challenge"
	consentChallengeCookieName = "bdd_test_cookie_consent_challenge"
	skipConsentCookieName      = "skipConsent"
	skipConsentTrue            = "true"
)

type store interface {
	Get(string) ([]byte, error)
	Put(string, []byte) error
}

type server struct {
	router *mux.Router
	hydra  client.AdminApi
	store  store
}

type config struct {
	hydraAdminURL *url.URL
	tlsConfig     *tls.Config
	store         store
}

func newServer(c *config) *server {
	hydraAdmin := client.NewAPIClient(&client.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "vcs-bdd-test",
		Debug:         true,
		Servers: client.ServerConfigurations{
			{
				URL: c.hydraAdminURL.String(),
			},
		},
		OperationServers: map[string]client.ServerConfigurations{},
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: c.tlsConfig,
			},
		},
	}).AdminApi

	router := mux.NewRouter()

	srv := &server{
		router: router,
		hydra:  hydraAdmin,
		store:  c.store,
	}

	router.HandleFunc("/login", srv.loginHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", srv.postLoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/consent", srv.consentHandler).Methods(http.MethodGet)
	router.HandleFunc("/authenticate", srv.userAuthenticationHandler).Methods(http.MethodPost)
	router.HandleFunc("/authorize", srv.userAuthorizationHandler).Methods(http.MethodPost)

	return srv
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

type testConfig struct {
	Request *AuthConfigRequest
}

type AuthConfigRequest struct {
	Sub  string `json:"sub"`
	Fail bool   `json:"fail,omitempty"`
}

type ConsentConfigRequest struct {
	UserClaims *UserClaims `json:"user_claims,omitempty"`
	Fail       bool        `json:"fail,omitempty"`
}

type UserClaims struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
}

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling request: %s", r.URL.String())

	challenge := r.URL.Query().Get("login_challenge")
	if challenge == "" {
		log.Printf("missing login_challenge")

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  loginChallengeCookieName,
		Value: challenge,
	})

	_, err := w.Write([]byte("<!DOCTYPE html>\n<html>\n<body>\n\n<p>mock login UI</p>\n\n<form action=\"/login\" method=\"post\" id=\"form1\">\n</form>\n\n<button type=\"submit\" form=\"form1\">login</button>\n\n</body>\n</html>\n"))
	if err != nil {
		log.Printf("failed to write imaginary UI: %s", err.Error())

		return
	}

	log.Printf("rendered mock login UI in response to request %s", r.URL.String())
}

func (s *server) postLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("success case %s", r.URL.String())

	http.SetCookie(w, &http.Cookie{
		Name:  skipConsentCookieName,
		Value: skipConsentTrue,
	})

	s.completeLogin(w, r, &AuthConfigRequest{
		Sub: uuid.New().String(),
	})
}

func (s *server) userAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var request *AuthConfigRequest

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Printf("failed to decode auth request: %s", err.Error())

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  skipConsentCookieName,
		Value: "",
	})

	s.completeLogin(w, r, request)
}

func (s *server) completeLogin(w http.ResponseWriter, r *http.Request, request *AuthConfigRequest) {
	cookie, err := r.Cookie(loginChallengeCookieName)
	if err != nil {
		log.Printf("failed to fetch cookie %s: %s", loginChallengeCookieName, err.Error())

		return
	}

	if request.Fail {
		var completedReq *client.CompletedRequest

		completedReq, _, err = s.hydra.RejectLoginRequest(context.Background()).LoginChallenge(cookie.Value).Execute()
		if err != nil {
			log.Printf("failed to reject login request at hydra: %s", err.Error())

			return
		}

		redirectURL := completedReq.GetRedirectTo()

		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("rejected login request; redirected to: %s", redirectURL)

		return
	}

	b, err := json.Marshal(&testConfig{
		Request: request,
	})
	if err != nil {
		log.Printf("failed to marshal test config: %s", err.Error())

		return
	}

	err = s.store.Put(request.Sub, b)
	if err != nil {
		log.Printf("failed to save test config: %s", err.Error())

		return
	}

	var completedReq *client.CompletedRequest

	completedReq, _, err = s.hydra.AcceptLoginRequest(r.Context()).
		AcceptLoginRequest(client.AcceptLoginRequest{
			Subject: request.Sub,
		}).
		LoginChallenge(cookie.Value).
		Execute()
	if err != nil {
		log.Printf("failed to accept hydra login request: %s", err.Error())

		return
	}

	redirectURL := completedReq.GetRedirectTo()

	http.Redirect(w, r, redirectURL, http.StatusFound)
	log.Printf("redirected to: %s", redirectURL)
}

func (s *server) consentHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling request: %s", r.URL.String())

	skipConsent, err := r.Cookie(skipConsentCookieName)
	if err != nil {
		log.Printf("failed to fetch cookie %s: %s", skipConsentCookieName, err.Error())

		return
	}

	log.Printf("consent skip value %s", skipConsent.Value)

	if skipConsent.Value == skipConsentTrue {
		s.completeConsent(w, r, &ConsentConfigRequest{UserClaims: &UserClaims{}}, r.URL.Query().Get("consent_challenge"))
	}

	challenge := r.URL.Query().Get("consent_challenge")
	if challenge == "" {
		log.Printf("missing consent_challenge")

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  consentChallengeCookieName,
		Value: challenge,
	})

	_, err = w.Write([]byte("mock consent UI"))
	if err != nil {
		log.Printf("failed to write imaginary UI: %s", err.Error())

		return
	}

	log.Printf("rendered mock consent UI in response to request %s", r.URL.String())
}

func (s *server) userAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(consentChallengeCookieName)
	if err != nil {
		log.Printf("failed to fetch cookie %s: %s", consentChallengeCookieName, err.Error())

		return
	}

	request := &ConsentConfigRequest{}

	err = json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		log.Printf("failed to decode user consent config request: %s", err.Error())

		return
	}

	s.completeConsent(w, r, request, cookie.Value)
}

func (s *server) completeConsent(w http.ResponseWriter, r *http.Request, request *ConsentConfigRequest, consentChallenge string) {
	consentReq, _, err := s.hydra.GetConsentRequest(r.Context()).ConsentChallenge(consentChallenge).Execute()
	if err != nil {
		log.Printf("failed to get hydra consent request: %s", err.Error())

		return
	}

	b, err := s.store.Get(consentReq.GetSubject())
	if err != nil {
		log.Printf("failed to fetch test config for sub=%s: %s", consentReq.GetSubject(), err.Error())

		return
	}

	var test testConfig

	err = json.Unmarshal(b, &test)
	if err != nil {
		log.Printf("failed to unmarshal user data %s: %s", b, err.Error())

		return
	}

	if request.Fail {
		completedReq, _, err := s.hydra.RejectConsentRequest(r.Context()).ConsentChallenge(consentChallenge).Execute()
		if err != nil {
			log.Printf("failed to reject consent request at hydra: %s", err.Error())

			return
		}

		redirectURL := completedReq.GetRedirectTo()

		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("user did not consent; redirected to %s", redirectURL)

		return
	}

	now := time.Now()

	completedReq, _, err := s.hydra.AcceptConsentRequest(r.Context()).AcceptConsentRequest(client.AcceptConsentRequest{
		GrantAccessTokenAudience: consentReq.GetRequestedAccessTokenAudience(),
		GrantScope:               consentReq.GetRequestedScope(),
		HandledAt:                &now,
		Session: &client.ConsentRequestSession{
			//AccessToken: nil,
			IdToken: request.UserClaims,
		},
	}).ConsentChallenge(consentChallenge).Execute()
	if err != nil {
		log.Printf("failed to accept hydra consent request: %s", err.Error())

		return
	}

	redirectURL := completedReq.GetRedirectTo()

	http.Redirect(w, r, redirectURL, http.StatusFound)
	log.Printf("user authorized; redirected to: %s", redirectURL)
}
