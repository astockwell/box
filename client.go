package box

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var GrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer" // Via https://github.com/box/box-python-sdk/blob/1.5/boxsdk/auth/jwt_auth.py#L21
var APIBaseURL = "https://api.box.com/2.0"
var APITokenURL = "https://api.box.com/oauth2/token"

type Client struct {
	ClientID                 string
	clientSecret             string
	EnterpriseID             string
	JWTKeyID                 string
	RSAPrivateKeyPemFilePath string
	GrantType                string
	APIBaseURL               string
	lastToken                *OauthTokenResponse
	lastTokenRetrieved       *time.Time
}

type OauthTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	RestrictedTo []string `json:"restricted_to"`
	TokenType    string   `json:"token_type"`
}

func NewClient(clientID, clientsecret, enterpriseID, jWTKeyID, rSAPrivateKeyPemFilePath string) (*Client, error) {
	return &Client{
		ClientID:                 clientID,
		clientSecret:             clientsecret,
		EnterpriseID:             enterpriseID,
		JWTKeyID:                 jWTKeyID,
		RSAPrivateKeyPemFilePath: rSAPrivateKeyPemFilePath,
		GrantType:                GrantType,
		APIBaseURL:               APIBaseURL,
	}, nil
}

func (c *Client) refreshAccessToken() error {
	// log.Println("Refreshing access token")
	tokenRequested := time.Now()

	// Generate Nonce
	jwtNonce, err := GenerateRandomString(32)
	if err != nil {
		return err
	}

	// Box JWT Claims reference: https://developer.box.com/v2.0/docs/construct-jwt-claim-manually#section-6-constructing-the-claims
	// TODO: allow for sub type of 'enterprise' or 'user' and make struct generic (instead of c.EnterpriseID, should be c.Sub I guess???)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":          c.ClientID,                              // (string, required) The Client ID of the service that created the JWT assertion.
		"sub":          c.EnterpriseID,                          // (string, required) One of either: enterprise_id for a token specific to an enterprise when creating and managing app users; OR app user_id for a token specific to an individual app user
		"box_sub_type": "enterprise",                            // (string, required) "enterprise" or "user" depending on the type of token being requested in the sub claim.
		"aud":          APITokenURL,                             // (string, required) Always “https://api.box.com/oauth2/token” for OAuth2 token requests
		"jti":          jwtNonce,                                // (string, required) A universally unique identifier specified by the client for this JWT. This is a unique string that is at least 16 characters and at most 128 characters.
		"exp":          time.Now().Add(30 * time.Second).Unix(), // (NumericDate, required) The unix time as to when this JWT will expire. This can be set to a maximum value of 60 seconds beyond the issue time. Note: It is recommended to set this value to less than the maximum allowed 60 seconds.
		// "iat":          "",                                 // (NumericDate, optional) Issued at time. The token cannot be used before this time.
		// "nbf":          "",                                 // (NumericDate, optional) Not before. Specifies when the token will start being valid.
	})

	// Box JWT Header reference: https://developer.box.com/v2.0/docs/construct-jwt-claim-manually#section-5-constructing-the-header
	token.Header["kid"] = c.JWTKeyID
	// spew.Dump(token)

	privateKeyPem, err := ioutil.ReadFile(c.RSAPrivateKeyPemFilePath)
	if err != nil {
		return err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPem)
	if err != nil {
		return err
	}
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(privateKey)
	// fmt.Println(tokenString, err)

	// Remove from memory
	privateKey = nil
	privateKeyPem = []byte{}

	// Get new access token from Oauth2 API
	res, err := http.PostForm(APITokenURL, url.Values{
		"grant_type":    {c.GrantType},
		"client_id":     {c.ClientID},
		"client_secret": {c.clientSecret},
		"assertion":     {tokenString},
	})
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		io.Copy(buf, res.Body)
		res.Body.Close()
		body := strings.TrimSpace(buf.String())
		// fmt.Println(body)
		return fmt.Errorf("Unexpected status code while retrieving new Oauth2 access token: [%v]. HTTP Response body: [%s]", res.StatusCode, body)
	}
	// spew.Dump(res)

	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	// body := strings.TrimSpace(buf.String())
	// fmt.Println(body)

	var otr OauthTokenResponse
	if err := json.Unmarshal(buf.Bytes(), &otr); err != nil {
		return err
	}
	if otr.AccessToken == "" {
		return fmt.Errorf("Unexpected blank access token from Oauth2 token API: %v", buf.String())
	}
	// spew.Dump(otr)

	c.lastToken = &otr
	c.lastTokenRetrieved = &tokenRequested

	return nil
}

func (c *Client) HttpDo(req *http.Request) (*http.Response, error) {
	// check c.lastToken != nil and is not expired
	// if nil or expired, get new one
	if c.lastToken == nil || c.lastTokenRetrieved == nil {
		err := c.refreshAccessToken()
		if err != nil {
			return nil, err
		}
	} else {
		lastTokenDuration, err := time.ParseDuration(fmt.Sprintf("%ds", c.lastToken.ExpiresIn-10))
		if err != nil {
			return nil, err
		}
		if time.Now().After(c.lastTokenRetrieved.Add(lastTokenDuration)) {
			err := c.refreshAccessToken()
			if err != nil {
				return nil, err
			}
		}
	}
	// spew.Dump(c.lastToken)

	// make request with valid access token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.lastToken.AccessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// log.Printf("Recieved (%s) response, retrying with new token\n", resp.Status)
		err := c.refreshAccessToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.lastToken.AccessToken))
		return http.DefaultClient.Do(req)
	}

	return resp, nil
}
