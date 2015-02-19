package replica

import (
	"encoding/json"
	"errors"
	"time"
)

// Token struct for token information
type Token struct {
	Token   string `json:"auth_token"`
	Expires int64  `json:"expires"`
}

func (t *Token) String() string { return t.Token }

// Valid return true if token is valid
func (t *Token) Valid() bool {
	return time.Now().Unix() < t.Expires
}

// GetToken returns a token or error if auth failes
func (c *Client) GetToken(user, passwd string) (*Token, error) {
	headers := map[string]string{"X-Auth-User": user, "X-Auth-Password": passwd}
	return c.getToken(headers)
}

// Token returns clients token
func (c *Client) Token() (*Token, error) {
	if c.token == nil || c.token.String() == "" {
		return nil, errors.New("token not set")
	}
	headers := map[string]string{"X-Auth-Token": c.token.String()}
	return c.getToken(headers)
}

func (c *Client) getToken(headers map[string]string) (*Token, error) {
	if c.token != nil && c.token.Valid() {
		return c.token, nil
	}
	req, err := c.newRequest("GET", "token", nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(c.token)
	if err != nil {
		return nil, err
	}
	return c.token, nil
}
