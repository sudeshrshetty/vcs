/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4vc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	fositeoauth "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/hmac"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/trustbloc/vcs/pkg/restapi/resterr"
	"github.com/trustbloc/vcs/pkg/restapi/v1/issuer"
	"github.com/trustbloc/vcs/pkg/restapi/v1/oidc4vc"
	oidc4vcsvc "github.com/trustbloc/vcs/pkg/service/oidc4vc"
	"github.com/trustbloc/vcs/pkg/storage/mongodb/oidc4vcstatestore"
)

const (
	clientID = "test-client"
)

var fositeStore = &storage.MemoryStore{ //nolint:gochecknoglobals
	Clients: map[string]fosite.Client{
		clientID: &fosite.DefaultClient{
			ID:            clientID,
			Secret:        []byte(`$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO`), // = "foobar"
			RedirectURIs:  []string{"/client/cb"},
			ResponseTypes: []string{"code"},
			GrantTypes:    []string{"authorization_code"},
			Scopes:        []string{"openid", "profile"},
		},
	},
	AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
	IDSessions:             make(map[string]fosite.Requester),
	AccessTokens:           map[string]fosite.Requester{},
	RefreshTokens:          map[string]storage.StoreRefreshToken{},
	PKCES:                  map[string]fosite.Requester{},
	Users:                  make(map[string]storage.MemoryUserRelation),
	AccessTokenRequestIDs:  map[string]string{},
	RefreshTokenRequestIDs: map[string]string{},
	IssuerPublicKeys:       map[string]storage.IssuerPublicKeys{},
	PARSessions:            map[string]fosite.AuthorizeRequester{},
}

func TestAuthorizeCodeGrantFlow(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = resterr.HTTPErrorHandler

	opState := "QIn85XAEHwlPyCVRhTww"

	srv := httptest.NewServer(e)
	defer srv.Close()

	// prepend client redirect URIs with test server URL
	for _, client := range fositeStore.Clients {
		c, ok := client.(*fosite.DefaultClient)
		if ok {
			c.RedirectURIs[0] = srv.URL + c.RedirectURIs[0]
		}
	}

	config := new(fosite.Config)
	config.EnforcePKCE = true

	var hmacStrategy = &fositeoauth.HMACSHAStrategy{
		Enigma: &hmac.HMACStrategy{
			Config: &fosite.Config{
				GlobalSecret: []byte("secret-for-signing-and-verifying-signatures"),
			},
		},
		Config: &fosite.Config{
			AuthorizeCodeLifespan: time.Minute,
			AccessTokenLifespan:   time.Hour,
		},
	}

	oauth2Provider := compose.Compose(config, fositeStore, hmacStrategy,
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2PKCEFactory,
		compose.PushedAuthorizeHandlerFactory,
	)

	controller := oidc4vc.NewController(&oidc4vc.Config{
		OAuth2Provider:          oauth2Provider,
		StateStore:              &memoryStateStore{kv: make(map[string]*oidc4vcstatestore.AuthorizeState)},
		IssuerInteractionClient: mockIssuerInteractionClient(t, srv.URL, opState),
		IssuerVCSPublicHost:     srv.URL,
	})

	oidc4vc.RegisterHandlers(e, controller)

	registerThirdPartyOIDCAuthorizeEndpoint(t, e)
	registerClientCallback(t, e)

	oauthClient := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: "foobar",
		RedirectURL:  srv.URL + "/client/cb",
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  srv.URL + "/oidc/token",
			AuthURL:   srv.URL + "/oidc/authorize",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	params := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", "MLSjJIlPzeRQoN9YiIsSzziqEuBSmS4kDgI3NDjbfF8"),
		oauth2.SetAuthURLParam("op_state", opState),
	}

	authCodeURL := oauthClient.AuthCodeURL(opState, params...)

	resp, err := http.DefaultClient.Get(authCodeURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	code := resp.Request.URL.Query().Get("code")
	require.NotEmpty(t, code)

	token, err := oauthClient.Exchange(context.TODO(), code,
		oauth2.SetAuthURLParam("code_verifier", "xalsLDydJtHwIQZukUyj6boam5vMUaJRWv-BnGCAzcZi3ZTs"),
	)

	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
}

func mockIssuerInteractionClient(
	t *testing.T,
	serverURL string,
	opState string,
) *MockIssuerInteractionClient {
	t.Helper()

	client := NewMockIssuerInteractionClient(gomock.NewController(t))

	client.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req issuer.PrepareAuthorizationRequestJSONRequestBody,
			reqEditors ...issuer.RequestEditorFn,
		) (*http.Response, error) {
			b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
				AuthorizationEndpoint: serverURL + "/third-party/oidc/authorize",
				AuthorizationRequest: issuer.OAuthParameters{
					ClientId:     clientID,
					ClientSecret: "foobar",
					ResponseType: req.ResponseType,
					Scope:        lo.FromPtr(req.Scope),
				},
				PushedAuthorizationRequestEndpoint: nil,
			})
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(b)),
			}, nil
		})

	client.EXPECT().StoreAuthorizationCodeRequest(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req issuer.StoreAuthorizationCodeRequestJSONRequestBody,
			reqEditors ...issuer.RequestEditorFn,
		) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(nil)),
			}, nil
		})

	client.EXPECT().ExchangeAuthorizationCodeRequest(
		gomock.Any(),
		issuer.ExchangeAuthorizationCodeRequestJSONRequestBody{
			OpState: opState,
		},
	).Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer(nil))}, nil)

	return client
}

func registerThirdPartyOIDCAuthorizeEndpoint(t *testing.T, e *echo.Echo) {
	t.Helper()

	e.GET("/third-party/oidc/authorize", func(c echo.Context) error {
		req := c.Request()

		// TODO: Validate authorize request

		q := &url.Values{}
		q.Set("code", "foo")
		q.Set("state", req.URL.Query().Get("state"))

		redirectURI := req.URL.Query().Get("redirect_uri") + "?" + q.Encode()

		return c.Redirect(http.StatusSeeOther, redirectURI)
	})
}

func registerClientCallback(t *testing.T, e *echo.Echo) {
	t.Helper()

	e.GET("/client/cb", func(c echo.Context) error {
		req := c.Request()

		code := req.URL.Query().Get("code")
		require.Contains(t, code, "ory_ac_")

		return nil
	})
}

type memoryStateStore struct {
	kv map[string]*oidc4vcstatestore.AuthorizeState
	mu sync.RWMutex
}

func (s *memoryStateStore) GetAuthorizeState(
	_ context.Context,
	opState string,
) (*oidc4vcstatestore.AuthorizeState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.kv[opState]
	if !ok {
		return nil, fmt.Errorf("key %s not found", opState)
	}

	return v, nil
}

func (s *memoryStateStore) SaveAuthorizeState(
	_ context.Context,
	opState string,
	state *oidc4vcstatestore.AuthorizeState,
	_ ...func(insertOptions *oidc4vcsvc.InsertOptions),
) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.kv[opState] = state

	return nil
}
