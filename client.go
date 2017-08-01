package rabbithole

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	// URI of a RabbitMQ node to use, not including the path, e.g. http://127.0.0.1:15672.
	Endpoint string
	// Username to use. This RabbitMQ user must have the "management" tag.
	Username string
	// Password to use.
	Password string
	transport *http.Transport
	host     string
	timeout  time.Duration
}

func NewClient(uri string, username string, password string) (me *Client, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	me = &Client{
		Endpoint: uri,
		host:     u.Host,
		Username: username,
		Password: password,
	}
	return me, nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Client) SetTransport(tr *http.Transport)  {
	c.transport = tr
}

func newGETRequest(client *Client, path string) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest("GET", s, nil)
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.host + "/api/" + path

	return req, err
}

func newRequestWithBody(client *Client, method string, path string, body []byte) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest(method, s, bytes.NewReader(body))
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.host + "/api/" + path

	req.Header.Add("Content-Type", "application/json")

	return req, err
}

func executeRequest(client *Client, req *http.Request) (res *http.Response, err error) {
	httpc := &http.Client{Timeout: client.timeout}
	if client.transport != nil {
		httpc.Transport = client.transport
	}
	return httpc.Do(req)
}

func executeAndParseRequest(client *Client, req *http.Request, rec interface{}) (err error) {
	res, err := executeRequest(client, req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(res.Status)
	}

	err = json.Unmarshal(body, &rec)
	if err != nil {
		return err
	}
	return nil
}
