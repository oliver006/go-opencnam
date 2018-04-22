package opencnam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	ocontext "golang.org/x/net/context"
)

type Client struct {
	l kitlog.Logger

	number endpoint.Endpoint

	sid   string
	token string
	hc    *http.Client
}

type RequestData struct {
	number     string
	sid, token string
}

type ResponseData struct {
	Name   string
	Number string
	Price  float64
	Uri    string
}

func mustParseURL(host, path string) *url.URL {
	r, err := url.Parse(host + path)
	if err != nil {
		panic("invalid url: " + err.Error())
	}
	return r
}

func retryEndpoint(e endpoint.Endpoint, l kitlog.Logger) endpoint.Endpoint {
	bl := sd.NewEndpointer(
		sd.FixedInstancer{"1"},
		sd.Factory(func(_ string) (endpoint.Endpoint, io.Closer, error) {
			return e, nil, nil
		}),
		l,
	)
	defer bl.Close()
	return lb.Retry(3, 10*time.Second, lb.NewRoundRobin(bl))
}

func NewClient(sid, token string, host string, l kitlog.Logger, opts ...httptransport.ClientOption) *Client {
	if host == "" {
		host = "https://api.opencnam.com"
	}
	return &Client{
		number: retryEndpoint(httptransport.NewClient(
			http.MethodGet,
			mustParseURL(host, "/v3/phone/"),
			encodeRequest,
			decodeResponse,
			opts...,
		).Endpoint(), l),
		sid:   sid,
		token: token,
	}
}

func encodeRequest(ctx context.Context, r *http.Request, req interface{}) error {
	reqData := req.(*RequestData)
	r.URL.Path += reqData.number
	r.URL.RawQuery = fmt.Sprintf(
		"format=json&account_sid=%s&auth_token=%s",
		reqData.sid,
		reqData.token,
	)
	return nil
}

func decodeResponse(ctx context.Context, r *http.Response) (interface{}, error) {
	switch r.StatusCode {
	case http.StatusOK:
		var t ResponseData
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read response")
		}
		err = json.Unmarshal(b, &t)
		return &t, err
	default:
		fmt.Printf("%d\n\n", r.StatusCode)
		b, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(b))
		return nil, fmt.Errorf("unable to read response")
	}
}

func (c *Client) NumberInfo(ctx ocontext.Context, number string) (*ResponseData, error) {
	req := RequestData{
		number: number,
		sid:    c.sid,
		token:  c.token,
	}
	out, err := c.number(ctx, &req)
	if err != nil {
		return nil, err
	}
	return out.(*ResponseData), nil
}

