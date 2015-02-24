package replica

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Client for replica server
type Client struct {
	addr        string
	token       *Token
	unsecureSSL bool
	useSSL      bool
	httpClient  *http.Client
}

// NewClient return a new instance of Client type
func NewClient(addr string, opts ...func(*Client)) (*Client, error) {
	if addr == "" {
		addr = "localhost"
	}
	addr = strings.TrimSuffix(addr, "/")
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		addr = "http://" + addr
		u, _ = url.Parse(addr)
	}
	if u.Path == "" && !strings.Contains(u.Host, ":") &&
		!strings.HasSuffix(addr, ":7881") {
		addr += ":7881/json"
	}
	if !strings.HasSuffix(addr, "/json") {
		addr += "/json"
	}
	client := &Client{
		addr:   addr,
		useSSL: u.Scheme == "https",
		token:  new(Token),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Address returns client connection address
func (c *Client) Address() string {
	return c.addr
}

// AllowUnsignedSSL skip verifying insecure keys
func AllowUnsignedSSL(c *Client) {
	c.unsecureSSL = true
}

// AssignToken set token for new client
func AssignToken(token string) func(*Client) {
	return func(c *Client) {
		c.token = &Token{Token: token}
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.httpClient == nil {
		if c.useSSL {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.unsecureSSL},
			}
			c.httpClient = &http.Client{Transport: tr}
		} else {
			c.httpClient = &http.Client{}
		}
	}
	if c.token != nil && c.token.String() != "" {
		req.Header.Add("X-Auth-Token", c.token.String())
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	err = parsResponse(resp)
	return resp, err
}

func (c *Client) joinURL(name string) string {
	addr := c.addr
	if name == "token" {
		addr = strings.TrimSuffix(addr, "/json")
	}
	name = strings.TrimSuffix(name, "/")
	if strings.HasPrefix(name, "/") {
		return addr + name
	}

	return addr + "/" + name
}

func (c *Client) newRequest(m, p string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(m, c.joinURL(p), r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Replica Client v0.1")
	return req, nil
}

func parsResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return newHTTPError(resp.StatusCode, buf)
	}
	return nil
}
