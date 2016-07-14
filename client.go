package httpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
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
	Content_Type_Json = "application/json"
	Content_Type_Xml  = "text/xml"
)

type HttpClient struct {
	R      *http.Request
	Client *http.Client
	Params url.Values
}
type Response struct {
	*http.Response
}

var DefaultTransport *http.Transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func New() *HttpClient {
	return &HttpClient{
		Client: http.DefaultClient,
	}
}

func Get(Url string) (*Response, error) {
	c := New()
	c.R = newRequest("GET", Url)
	return c.Do()
}

func Head(Url string) (*Response, error) {
	c := New()
	c.R = newRequest("HEAD", Url)
	return c.Do()
}

func Post(Url, bodyType string, body io.Reader) (*Response, error) {
	c := New()
	c.R = newRequest("POST", Url)
	c.R.Header.Set("Content-Type", bodyType)
	c.SetBody(body)
	return c.Do()
}

//Request
func newRequest(method, Url string) *http.Request {
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

func (h *HttpClient) Get(Url string) {
	h.R = newRequest("GET", Url)
}

func (h *HttpClient) Post(Url, bodyType string, body io.Reader) {
	r := newRequest("POST", Url)
	r.Header.Set("Content-Type", bodyType)
	h.R = r
	h.SetBody(body)
}

func (h *HttpClient) Head(Url string) {
	h.R = newRequest("HEAD", Url)
}

func (h *HttpClient) Put(Url string) {
	h.R = newRequest("PUT", Url)
}

func (h *HttpClient) Patch(Url string) {
	h.R = newRequest("PATCH", Url)
}

func (h *HttpClient) Delete(Url string) {
	h.R = newRequest("DELETE", Url)
}

func (h *HttpClient) Options(Url string) {
	h.R = newRequest("OPTIONS", Url)
}

func (h *HttpClient) PostForm(Url string, v url.Values) {
	h.R = newRequest("POST", Url)
	h.R.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if v != nil {
		h.SetBody(strings.NewReader(v.Encode()))
	}
}

func (h *HttpClient) PostJson(Url string, o interface{}) error {
	h.R = newRequest("POST", Url)
	body, err := json.Marshal(o)
	if err != nil {
		return err
	}
	h.SetBody(bytes.NewBuffer(body))
	h.R.Header.Set("Content-Type", "application/json")
	return nil
}

func (h *HttpClient) PostXml(Url string, o interface{}) error {
	h.R = newRequest("POST", Url)
	body, err := xml.Marshal(o)
	if err != nil {
		return err
	}
	h.SetBody(bytes.NewBuffer(body))
	h.R.Header.Set("Content-Type", "text/xml")
	return nil
}

func (h *HttpClient) SetBody(body io.Reader) {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			h.R.ContentLength = int64(v.Len())
		case *bytes.Reader:
			h.R.ContentLength = int64(v.Len())
		case *strings.Reader:
			h.R.ContentLength = int64(v.Len())
		}
	}
	h.R.Body = rc
}

func (h *HttpClient) AddCookie(key, value string) {
	h.R.AddCookie(&http.Cookie{Name: key, Value: value})
}

func (h *HttpClient) UserAgent(UA string) {
	h.R.Header.Set("User-Agent", UA)
}

func (h *HttpClient) Host(hostname string) {
	h.R.Host = hostname
}

func (h *HttpClient) Header() http.Header {
	if h.R.Header == nil {
		h.R.Header = make(http.Header)
	}
	return h.R.Header
}

func (h *HttpClient) QueryAdd(key, value string) {
	q := h.R.URL.Query()
	q.Add(key, value)
	h.R.URL.RawQuery = q.Encode()
}

func (h *HttpClient) QuerySet(key, value string) {
	q := h.R.URL.Query()
	q.Set(key, value)
	h.R.URL.RawQuery = q.Encode()
}

func (h *HttpClient) QueryDel(key string) {
	q := h.R.URL.Query()
	q.Del(key)
	h.R.URL.RawQuery = q.Encode()
}

func (h *HttpClient) QueryGet(key string) {
	q := h.R.URL.Query()
	q.Get(key)
	h.R.URL.RawQuery = q.Encode()
}

func (h *HttpClient) BasicAuth(username, password string) {
	h.R.Header.Set("Authorization", "Basic "+basicAuth(username, password))
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//Client
func (h *HttpClient) SetCheckRedirect(f func(req *http.Request, via []*http.Request) error) {
	h.Client.CheckRedirect = f
}

func (h *HttpClient) UseCookiejar() {
	jar, _ := cookiejar.New(nil)
	h.Client.Jar = jar
}

func (h *HttpClient) SetTimeout(t time.Duration) {
	h.Client.Timeout = t
}

func (h *HttpClient) UseProxy(proxyip string) {
	if !strings.Contains(proxyip, "http://") && !strings.Contains(proxyip, "https://") {
		proxyip = "http://" + proxyip
	}
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyip)
	}
	DefaultTransport.Proxy = proxy
	h.Client.Transport = DefaultTransport
}

func (h *HttpClient) Do() (*Response, error) {
	res, err := h.Client.Do(h.R)
	if err != nil {
		return nil, err
	}
	return &Response{res}, nil
}

//Response
func (r *Response) Byte() []byte {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil
	}
	return b
}

func (r *Response) String() string {
	return string(r.Byte())
}

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
