package notbearclient

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dsnet/compress/brotli"

	"golang.org/x/net/publicsuffix"

	iconv "gopkg.in/iconv.v1"
)

var TimeoutError = errors.New("Timeout error")
var InterruptError = errors.New("Catch interrupt signal")
var CharsetRe = regexp.MustCompile(`charset=([\w-_]*)`)

func NewRequest(method, URL, contentType, headerName string, body map[string][]string) (*http.Request, error) {
	req, err := http.NewRequest(method, URL, strings.NewReader(url.Values(body).Encode()))
	if err != nil {
		return nil, err
	}
	if method == "POST" {
		if contentType == "" {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Add("Content-Type", contentType)
		}
	}
	if headerName != "" {
		if headers, ok := HeadersMap[headerName]; ok {
			for k, v := range headers {
				req.Header.Add(k, v)
			}
		} else {
			return nil, fmt.Errorf("%s does not exist", headerName)
		}
	}
	return req, nil
}

type Client struct {
	HttpClient *http.Client
	RetryTimes int
	Input      chan *http.Request
	Output     chan string
	Done       chan struct{}
	Error      chan error
	Context    context.Context
}

func NewClient(retryTimes, timeout int, ctx context.Context, err chan error) *Client {
	jar, jarErr := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if jarErr != nil {
		panic(jarErr)
	}
	proxyURL, _ := url.Parse("http://127.0.0.1:44625")
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		Proxy:               http.ProxyURL(proxyURL),
	}
	httpClient := &http.Client{
		Jar:       jar,
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	input := make(chan *http.Request, 16)
	output := make(chan string, 16)
	done := make(chan struct{}, 1)
	client := &Client{
		HttpClient: httpClient,
		RetryTimes: retryTimes,
		Input:      input,
		Output:     output,
		Done:       done,
		Error:      err,
		Context:    ctx,
	}
	return client
}

func (c *Client) Do(req *http.Request) *http.Response {
	var resp *http.Response
	var err error
OUTER:
	for i := 0; i < c.RetryTimes; i++ {
		select {
		case <-c.Context.Done():
			return &http.Response{}
		default:
			resp, err = c.HttpClient.Do(req)
			if err != nil {
				// if err, ok := err.(net.Error); ok && err.Timeout() {
				// 	if i < c.RetryTimes-1 {
				// 		errTimeout := NewErrTimeout(req.URL, i+1, err)
				// 		c.Error <- errTimeout
				// 		continue OUTER
				// 	} else {
				// 		errFailed := NewErrFailed(req.URL, err)
				// 		c.Error <- errFailed
				// 		return nil
				// 	}
				// } else {
				// 	errNetwork := NewErrNetwork(req.URL, err)
				// 	c.Error <- errNetwork
				// 	return nil
				// }
				continue OUTER
			}
			break OUTER
		}
	}
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Timeout() {
				errTimeout := NewErrTimeout(req.URL, err)
				c.Error <- errTimeout
			} else {
				errNetwork := NewErrNetwork(req.URL, err)
				c.Error <- errNetwork
			}
		default:
			errOther := NewErrOther(req.URL, err)
			c.Error <- errOther
		}
	}
	return resp
}

func (c *Client) ReadResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	defer resp.Body.Close()
	var reader io.Reader
	var charset string
	contentEncoding := resp.Header.Get("content-encoding")
	contentType := resp.Header.Get("content-type")
	charsetList := CharsetRe.FindStringSubmatch(contentType)
	if len(charsetList) > 0 {
		charset = charsetList[1]
	}
	switch contentEncoding {
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		reader = gzipReader
	case "br":
		brReader, err := brotli.NewReader(resp.Body, &brotli.ReaderConfig{})
		if err != nil {
			panic(err)
		}
		reader = brReader
	default:
		reader = resp.Body
	}

	if charset != "" && strings.ToLower(charset) != "utf-8" {
		c, err := iconv.Open("utf-8", charset)
		if err != nil {
			panic(err)
		}
		reader = iconv.NewReader(c, reader, 0)
	}
	var byteContent []byte
	var err error
OUTER:
	for i := 0; i < c.RetryTimes; i++ {
		select {
		case <-c.Context.Done():
			return ""
		default:
			byteContent, err = ioutil.ReadAll(reader)
			if err != nil {
				continue OUTER
			}
			break OUTER
		}
	}
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Timeout() {
				errTimeout := NewErrTimeout(resp.Request.URL, err)
				c.Error <- errTimeout
			} else {
				errNetwork := NewErrNetwork(resp.Request.URL, err)
				c.Error <- errNetwork
			}
		default:
			errOther := NewErrOther(resp.Request.URL, err)
			c.Error <- errOther
		}
	}
	// byteContent, err := ioutil.ReadAll(reader)
	// if err != nil {
	// 	c.Error <- err
	// 	return ""
	// }
	content := string(byteContent[:])
	return content
}

func (c *Client) Close() {
	close(c.Output)
	c.Done <- struct{}{}
	close(c.Done)
}

func (c *Client) Run() {
	defer c.Close()
	for {
		select {
		case <-c.Context.Done():
			return
		case request, ok := <-c.Input:
			if ok {
				resp := c.Do(request)
				s := c.ReadResponse(resp)
				c.Output <- s
			} else {
				return
			}
		}
	}
}
