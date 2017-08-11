package httpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	Content_Type_From = "application/x-www-form-urlencoded"
	Content_Type_Json = "application/json;charset=UTF-8"
	Content_Type_Xml  = "text/xml"
)

type HttpClient struct {
	Request *http.Request
	Client  *http.Client
	Query   url.Values //QueryString
	Param   url.Values //PostFromParams
}
type Response struct {
	*http.Response
}

var DefaultTransport *http.Transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

//New return a HttpClient instance
func New() *HttpClient {
	return &HttpClient{
		Client: http.DefaultClient,
		Query:  url.Values{},
		Param:  url.Values{},
	}
}

//Get
func Get(Url string) (*Response, error) {
	c := New()
	c.Request = newRequest(http.MethodGet, Url)
	return c.Do()
}

//Head
func Head(Url string) (*Response, error) {
	c := New()
	c.Request = newRequest(http.MethodHead, Url)
	return c.Do()
}

//Post
func Post(Url, bodyType string, body io.Reader) (*Response, error) {
	c := New()
	c.Request = newRequest(http.MethodPost, Url)
	c.Request.Header.Set("Content-Type", bodyType)
	c.Body(body)
	return c.Do()
}

//Request
func newRequest(method, Url string) *http.Request {
	if !strings.HasPrefix(Url, "//") {
		if !strings.HasPrefix(Url, "http://") && !strings.HasPrefix(Url, "https://") {
			Url = "http://" + Url
		}
	}
	u, err := url.Parse(Url)
	if err != nil {
		panic(err.Error())
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	req := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	return req
}

//flush old query and old param
func (h *HttpClient) New() *HttpClient {
	h.Query = url.Values{}
	h.Param = url.Values{}
	return h
}

// set MethodGet and Url
func (h *HttpClient) Get(Url string) {
	h.Request = newRequest(http.MethodGet, Url)
	// h.Query = h.Request.URL.Query()
}

// set MedthodPost  and Url,Content-Type header,body
func (h *HttpClient) Post(Url, bodyType string, body io.Reader) {
	r := newRequest(http.MethodPost, Url)
	r.Header.Set("Content-Type", bodyType)
	h.Request = r
	// h.Query = r.URL.Query()
	h.Body(body)
}

// set Method=Head and Url
func (h *HttpClient) Head(Url string) {
	h.Request = newRequest(http.MethodHead, Url)
	// h.Query = h.Request.URL.Query()
}

//set Method=Put and Url
func (h *HttpClient) Put(Url string) {
	h.Request = newRequest(http.MethodPut, Url)
	// h.Query = h.Request.URL.Query()
}

//set Method=Patch and Url
func (h *HttpClient) Patch(Url string) {
	h.Request = newRequest(http.MethodPatch, Url)
	// h.Query = h.Request.URL.Query()
}

//set Method=Delete and Url
func (h *HttpClient) Delete(Url string) {
	h.Request = newRequest(http.MethodDelete, Url)
	// h.Query = h.Request.URL.Query()
}

//set Method=Options and Url
func (h *HttpClient) Options(Url string) {
	h.Request = newRequest(http.MethodOptions, Url)
	// h.Query = h.Request.URL.Query()
}

//set Method=Post,Content-Type="application/x-www-form-urlencoded",body=h.Param
func (h *HttpClient) PostForm(Url string) {
	h.Request = newRequest(http.MethodPost, Url)
	// h.Query = h.Request.URL.Query()
	h.Request.Header.Set("Content-Type", Content_Type_From)
	if len(h.Param) != 0 {
		h.Body(strings.NewReader(h.Param.Encode()))
	}
}

//set Method=Post,Content-Type="application/json",body=json.Marshal(o)
func (h *HttpClient) PostJson(Url string, o interface{}) error {
	h.Request = newRequest(http.MethodPost, Url)
	// h.Query = h.Request.URL.Query()
	body, err := json.Marshal(o)
	if err != nil {
		return err
	}
	h.Body(bytes.NewReader(body))
	h.Request.Header.Set("Content-Type", Content_Type_Json)
	return nil
}

//set Method=Post,Content-Type="text/xml",body=json.Marshal(o)
func (h *HttpClient) PostXml(Url string, o interface{}) error {
	h.Request = newRequest(http.MethodPost, Url)
	// h.Query = h.Request.URL.Query()
	body, err := xml.Marshal(o)
	if err != nil {
		return err
	}
	h.Body(bytes.NewBuffer(body))
	h.Request.Header.Set("Content-Type", Content_Type_Xml)
	return nil
}

//set Request Body
func (h *HttpClient) Body(body io.Reader) {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			h.Request.ContentLength = int64(v.Len())
		case *bytes.Reader:
			h.Request.ContentLength = int64(v.Len())
		case *strings.Reader:
			h.Request.ContentLength = int64(v.Len())
		}
	}
	h.Request.Body = rc
}

//Add Cookie
func (h *HttpClient) AddCookie(key, value string) {
	h.Request.AddCookie(&http.Cookie{Name: key, Value: value})
}

//Set User-Agent
func (h *HttpClient) UserAgent(UA string) {
	h.Request.Header.Set("User-Agent", UA)
}

//set Host header
func (h *HttpClient) Host(hostname string) {
	h.Request.Host = hostname
}

//return Header
func (h *HttpClient) Header() http.Header {
	return h.Request.Header
}

//set BasicAuth
func (h *HttpClient) BasicAuth(username, password string) {
	h.Request.Header.Set("Authorization", "Basic "+basicAuth(username, password))
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//when request Redirect will execute f()
func (h *HttpClient) SetCheckRedirect(f func(req *http.Request, via []*http.Request) error) {
	h.Client.CheckRedirect = f
}
func defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if len(via) == 0 {
		return nil
	}
	// Redirect requests with the first Header
	for key, val := range via[0].Header {
		// Don't copy Referer Header
		if key != "Referer" {
			req.Header[key] = val
		}
	}
	return nil
}

//enable Cookie
func (h *HttpClient) UseCookiejar() {
	jar, _ := cookiejar.New(nil)
	h.Client.Jar = jar
}

//set request timeout
func (h *HttpClient) SetTimeout(t time.Duration) {
	h.Client.Timeout = t
}

//Do  return Response and err
func (h *HttpClient) Do() (*Response, error) {
	rawquery := h.Query.Encode()
	if rawquery != "" && h.Request.URL.RawQuery != "" {
		rawquery = "&" + rawquery
	}
	h.Request.URL.RawQuery += rawquery
	if h.Client.CheckRedirect == nil {
		h.Client.CheckRedirect = defaultCheckRedirect
	}
	if len(h.Param) > 0 {
		h.Body(strings.NewReader(h.Param.Encode()))
	}
	res, err := h.Client.Do(h.Request)
	if err != nil {
		return nil, err
	}
	return &Response{res}, nil
}

//Use Proxy
func (h *HttpClient) UseProxy(host string) error {
	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	Transport := DefaultTransport
	switch u.Scheme {
	case "http", "https":
		Transport.Proxy = http.ProxyURL(u)
		h.Client.Transport = Transport
	case "socks5":
		dialer, err := proxy.FromURL(u, proxy.Direct)
		if err != nil {
			return err
		}
		Transport.Proxy = http.ProxyFromEnvironment
		Transport.Dial = dialer.Dial
		h.Client.Transport = Transport
	}
	return nil
}

//Response body to []byte
func (r *Response) Byte() []byte {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil
	}
	return b
}

//Response body to string
func (r *Response) String() string {
	return string(r.Byte())
}

//Response body save as a file
func (r *Response) DownLoadFile(filepath string) error {
	dir, _ := path.Split(filepath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0666); err != nil {
			return err
		}
	}
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, bytes.NewReader(r.Byte()))
	return nil
}

//Json.Unmarshal ResponseBody
func (r *Response) JsonUnmarshal(v interface{}) error {
	return json.Unmarshal(r.Byte(), v)
}
