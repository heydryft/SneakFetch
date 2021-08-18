package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"sneakfetch/cclient"
	"sneakfetch/http"
	"strings"
	"time"

	tls "sneakfetch/utls"

	"github.com/dsnet/compress/brotli"
	"github.com/google/uuid"
)

var defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4469.0 Safari/537.36"

type Response struct {
	body          string
	rawBody       []byte
	headers       http.Header
	statusCode    int
	statusMessage string
	error         bool
	errorMessage  string
}

type Header struct {
	key   string
	value string
}

type kv map[string]string

var cookieStores = map[string]map[string]kv{}

func New() Fetcher {
	id := uuid.New().String()
	cookieStores[id] = map[string]kv{}
	return Fetcher{
		id: id,
	}
}

func (fetcher Fetcher) GetCookies(domain string) kv {
	if cookieStores[fetcher.id] != nil && len(cookieStores[fetcher.id][domain]) != 0 {
		return cookieStores[fetcher.id][domain]
	}
	return kv{}
}

func (fetcher Fetcher) GetCookie(domain string, name string) string {
	if cookieStores[fetcher.id] != nil && len(cookieStores[fetcher.id][domain]) != 0 {
		return cookieStores[fetcher.id][domain][name]
	} else {
		return ""
	}
}

func (fetcher Fetcher) SetCookie(domain string, name string, value string) {
	if cookieStores[fetcher.id][domain] == nil {
		cookieStores[fetcher.id][domain] = kv{}
	}
	cookieStores[fetcher.id][domain][name] = value
}

func (fetcher Fetcher) SetCookies(domain string, cookies kv) {
	if cookieStores[fetcher.id][domain] == nil {
		cookieStores[fetcher.id][domain] = kv{}
	}
	for k, v := range cookies {
		cookieStores[fetcher.id][domain][k] = v
	}
}

func cookiesToHeader(cookies kv) string {
	arr := []string{}
	for k, v := range cookies {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(arr, ", ")
}

func headerToCookies(header []string) kv {
	cookies := kv{}
	for _, s := range header {
		arr := strings.Split(strings.Split(s, ";")[0], "=")
		cookies[arr[0]] = arr[1]
	}
	return cookies
}

func parseResponse(resp *http.Response) Response {
	var body string
	var reader io.Reader
	if resp.Header.Get("content-encoding") == "br" {
		reader, _ = brotli.NewReader(resp.Body, &brotli.ReaderConfig{})
	} else if resp.Header.Get("content-encoding") == "gzip" {
		reader, _ = gzip.NewReader(resp.Body)
	} else if resp.Header.Get("content-encoding") == "deflate" {
		reader = flate.NewReader(resp.Body)
	} else {
		reader = resp.Body
	}
	rawBody, err := ioutil.ReadAll(reader)
	if err != nil {
		body = ""
	}
	body = string(rawBody)
	return Response{
		statusCode:    resp.StatusCode,
		statusMessage: resp.Status,
		headers:       resp.Header,
		body:          body,
		rawBody:       rawBody,
		error:         false,
		errorMessage:  "",
	}
}

type RequestOptions struct {
	headers       http.Header
	body          []byte
	proxy         string
	clientHelloID tls.ClientHelloID
}

type Fetcher struct {
	id string
}

const letterBytes = "123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getDomain(u string) string {
	_url, err := url.Parse(u)
	if err != nil {
		return ""
	}
	arr := strings.Split(_url.Hostname(), ".")
	return arr[len(arr)-2] + "." + arr[len(arr)-1]
}

func setChromeHeaders(headers http.Header) http.Header {

	if len(headers["Connection"]) == 0 {
		headers["Connection"] = []string{"keep-alive"}
	}

	if len(headers["sec-ch-ua"]) == 0 {
		headers["sec-ch-ua"] = []string{"\"Chromium\";v=\"91\", \" Not A;Brand\";v=\"99\", \"Google Chrome\";v=\"91\""}
	}

	if len(headers["Upgrade-Insecure-Requests"]) == 0 {
		headers["Upgrade-Insecure-Requests"] = []string{"1"}
	}

	if len(headers["sec-ch-ua-mobile"]) == 0 {
		headers["sec-ch-ua-mobile"] = []string{"?0"}
	}

	if len(headers["User-Agent"]) == 0 {
		headers["User-Agent"] = []string{defaultUserAgent}
	}

	if len(headers["Accept"]) == 0 {
		headers["Accept"] = []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"}
	}

	if len(headers["Sec-Fetch-Site"]) == 0 {
		headers["Sec-Fetch-Site"] = []string{"none"}
	}

	if len(headers["Sec-Fetch-Mode"]) == 0 {
		headers["Sec-Fetch-Mode"] = []string{"navigate"}
	}

	if len(headers["Sec-Fetch-User"]) == 0 {
		headers["Sec-Fetch-User"] = []string{"?1"}
	}

	if len(headers["Sec-Fetch-Dest"]) == 0 {
		headers["Sec-Fetch-Dest"] = []string{"document"}
	}

	if len(headers["Accept-Encoding"]) == 0 {
		headers["Accept-Encoding"] = []string{"gzip, deflate, br"}
	}

	if len(headers["Accept-Language"]) == 0 {
		headers["Accept-Language"] = []string{"en-US,en;q=0.9,it;q=0.8"}
	}

	if len(headers["If-None-Match"]) == 0 {
		headers["If-None-Match"] = []string{fmt.Sprintf(`W/"2a6-%s/%s"`, randStr(11), randStr(15))}
	}

	return headers
}

func (fetcher Fetcher) Get(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID tls.ClientHelloID

	if len(options.clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.headers["User-Agent"] != nil && len(options.headers["User-Agent"]) > 0 {
		userAgent = options.headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	if len(options.proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}
	if options.headers == nil {
		options.headers = http.Header{}
	}

	options.headers = setChromeHeaders(options.headers)

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "GET",
		Header: options.headers,
	})

	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}

	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func (fetcher Fetcher) Post(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID tls.ClientHelloID

	if len(options.clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.headers["User-Agent"] != nil && len(options.headers["User-Agent"]) > 0 {
		userAgent = options.headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	if len(options.proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}
	if options.headers == nil {
		options.headers = http.Header{}
	}

	options.headers = setChromeHeaders(options.headers)

	reader := bytes.NewReader(options.body)

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "POST",
		Header: options.headers,
		Body:   io.NopCloser(reader),
	})

	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}

	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func (fetcher Fetcher) Put(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID tls.ClientHelloID

	if len(options.clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.headers["User-Agent"] != nil && len(options.headers["User-Agent"]) > 0 {
		userAgent = options.headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	if len(options.proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}
	if options.headers == nil {
		options.headers = http.Header{}
	}

	options.headers = setChromeHeaders(options.headers)

	reader := bytes.NewReader(options.body)

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "PUT",
		Header: options.headers,
		Body:   io.NopCloser(reader),
	})

	if err != nil {
		return Response{
			error:        true,
			errorMessage: err.Error(),
		}
	}

	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func main() {
	rand.Seed(time.Now().UnixNano())
}
