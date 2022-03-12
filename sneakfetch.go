package sneakfetch

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
	"strings"
	"sync"

	"github.com/Jishrocks/SneakFetch/cclient"
	"github.com/Jishrocks/SneakFetch/config"
	"github.com/Jishrocks/SneakFetch/http"
	tls "github.com/Jishrocks/SneakFetch/utls"

	"github.com/dsnet/compress/brotli"
)

var defaultUserAgent = config.DefaultUserAgent

type Response struct {
	Body          string
	RawBody       []byte
	Headers       http.Header
	StatusCode    int
	StatusMessage string
	Error         bool
	ErrorMessage  string
}

type RequestOptions struct {
	Headers       http.Header
	Body          []byte
	Proxy         string
	ClientHelloID tls.ClientHelloID
}

type Header = http.Header

type Headers map[string]string

type kv map[string]string

var cookieStores = map[string]map[string]kv{}

// /************************************************************
//  *    _____   ______                _   _                   *
//  *   / ____| |  ____|              | | (_)                  *
//  *  | |      | |__ _   _ _ __   ___| |_ _  ___  _ __  ___   *
//  *  | |      |  __| | | | '_ \ / __| __| |/ _ \| '_ \/ __|  *
//  *  | |____  | |  | |_| | | | | (__| |_| | (_) | | | \__ \  *
//  *   \_____| |_|   \__,_|_| |_|\___|\__|_|\___/|_| |_|___/  *
//  *                                                          *
//  ************************************************************/
// // #include <stdio.h>
// // #include <stdlib.h>

// //export c_New
// func c_New() *C.char {
// 	fetcher := New()
// 	return C.CString(fetcher.id)
// }

// //export c_GetCookies
// func c_GetCookies(C_fetcherId *C.char, C_domain *C.char) *C.char {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	cookies := fetcher.GetCookies(domain)
// 	data, _ := json.Marshal(cookies)
// 	return C.CString(string(data))
// }

// //export c_GetCookie
// func c_GetCookie(C_fetcherId *C.char, C_domain *C.char, C_name *C.char) *C.char {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 		name      = C.GoString(C_name)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	cookie := fetcher.GetCookie(domain, name)
// 	return C.CString(cookie)
// }

// //export c_SetCookies
// func c_SetCookies(C_fetcherId *C.char, C_domain *C.char, cookies *C.char) {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	var data kv
// 	json.Unmarshal([]byte(C.GoString(cookies)), &data)
// 	fetcher.SetCookies(domain, data)
// }

// //export c_SetCookie
// func c_SetCookie(C_fetcherId *C.char, C_domain *C.char, C_name *C.char, C_value *C.char) {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 		name      = C.GoString(C_name)
// 		value     = C.GoString(C_value)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	fetcher.SetCookie(domain, name, value)
// }

// //export c_DeleteCookies
// func c_DeleteCookies(C_fetcherId *C.char, C_domain *C.char) {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	fetcher.DeleteCookies(domain)
// }

// //export c_DeleteCookie
// func c_DeleteCookie(C_fetcherId *C.char, C_domain *C.char, C_name *C.char) {
// 	var (
// 		fetcherId = C.GoString(C_fetcherId)
// 		domain    = C.GoString(C_domain)
// 		name      = C.GoString(C_name)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	fetcher.DeleteCookie(domain, name)
// }

// type c_RequestOptions struct {
// 	Headers Headers
// 	Body    string
// 	Proxy   string
// 	Client  string
// }

// func c_RespToJson(resp Response) *C.char {
// 	rawBody := base64.StdEncoding.EncodeToString(resp.RawBody)
// 	headers := Headers{}
// 	for k, v := range resp.Headers {
// 		headers[k] = v[0]
// 	}
// 	data, _ := json.Marshal(map[string]interface{}{
// 		"body":          resp.Body,
// 		"rawBody":       rawBody,
// 		"error":         resp.Error,
// 		"errorMessage":  resp.ErrorMessage,
// 		"headers":       headers,
// 		"statusCode":    resp.StatusCode,
// 		"statusMessage": resp.StatusMessage,
// 	})
// 	return C.CString(string(data))
// }

// //export c_Get
// func c_Get(C_fetcherId *C.char, C_requestURL *C.char, C_requestOptions *C.char) *C.char {
// 	var (
// 		fetcherId  = C.GoString(C_fetcherId)
// 		requestURL = C.GoString(C_requestURL)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	var _C_requestOptions c_RequestOptions
// 	requestOptions := RequestOptions{
// 		Headers: http.Header{},
// 	}

// 	err := json.Unmarshal([]byte(string(C.GoString(C_requestOptions))), &_C_requestOptions)

// 	if err == nil {
// 		if len(_C_requestOptions.Proxy) > 0 {
// 			requestOptions.Proxy = _C_requestOptions.Proxy
// 		}
// 		if len(getTLSParrot(_C_requestOptions.Client).Client) > 0 {
// 			requestOptions.ClientHelloID = getTLSParrot(_C_requestOptions.Client)
// 		}
// 		if len(_C_requestOptions.Headers) > 0 {
// 			for headerKey, headerVal := range _C_requestOptions.Headers {
// 				requestOptions.Headers[headerKey] = []string{headerVal}
// 			}
// 		}
// 	}

// 	resp := fetcher.Get(requestURL, requestOptions)
// 	return c_RespToJson(resp)
// }

// //export c_Post
// func c_Post(C_fetcherId *C.char, C_requestURL *C.char, C_requestOptions *C.char) *C.char {
// 	var (
// 		fetcherId  = C.GoString(C_fetcherId)
// 		requestURL = C.GoString(C_requestURL)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	var _C_requestOptions c_RequestOptions
// 	requestOptions := RequestOptions{
// 		Headers: http.Header{},
// 	}

// 	err := json.Unmarshal([]byte(string(C.GoString(C_requestOptions))), &_C_requestOptions)

// 	if err == nil {
// 		if len(_C_requestOptions.Proxy) > 0 {
// 			requestOptions.Proxy = _C_requestOptions.Proxy
// 		}
// 		if len(_C_requestOptions.Body) > 0 {
// 			body, _ := base64.StdEncoding.DecodeString(_C_requestOptions.Body)
// 			requestOptions.Body = body
// 		}
// 		if len(getTLSParrot(_C_requestOptions.Client).Client) > 0 {
// 			requestOptions.ClientHelloID = getTLSParrot(_C_requestOptions.Client)
// 		}
// 		if len(_C_requestOptions.Headers) > 0 {
// 			for headerKey, headerVal := range _C_requestOptions.Headers {
// 				requestOptions.Headers[headerKey] = []string{headerVal}
// 			}
// 		}
// 	}

// 	resp := fetcher.Post(requestURL, requestOptions)
// 	return c_RespToJson(resp)
// }

// //export c_Put
// func c_Put(C_fetcherId *C.char, C_requestURL *C.char, C_requestOptions *C.char) *C.char {
// 	var (
// 		fetcherId  = C.GoString(C_fetcherId)
// 		requestURL = C.GoString(C_requestURL)
// 	)
// 	fetcher := Fetcher{
// 		id: fetcherId,
// 	}
// 	var _C_requestOptions c_RequestOptions
// 	requestOptions := RequestOptions{
// 		Headers: http.Header{},
// 	}

// 	err := json.Unmarshal([]byte(string(C.GoString(C_requestOptions))), &_C_requestOptions)

// 	if err == nil {
// 		if len(_C_requestOptions.Proxy) > 0 {
// 			requestOptions.Proxy = _C_requestOptions.Proxy
// 		}
// 		if len(_C_requestOptions.Body) > 0 {
// 			body, _ := base64.StdEncoding.DecodeString(_C_requestOptions.Body)
// 			requestOptions.Body = body
// 		}
// 		if len(getTLSParrot(_C_requestOptions.Client).Client) > 0 {
// 			requestOptions.ClientHelloID = getTLSParrot(_C_requestOptions.Client)
// 		}
// 		if len(_C_requestOptions.Headers) > 0 {
// 			for headerKey, headerVal := range _C_requestOptions.Headers {
// 				requestOptions.Headers[headerKey] = []string{headerVal}
// 			}
// 		}
// 	}

// 	resp := fetcher.Put(requestURL, requestOptions)
// 	return c_RespToJson(resp)
// }

/******************************************************************
 *    _____         ______                _   _                   *
 *   / ____|       |  ____|              | | (_)                  *
 *  | |  __  ___   | |__ _   _ _ __   ___| |_ _  ___  _ __  ___   *
 *  | | |_ |/ _ \  |  __| | | | '_ \ / __| __| |/ _ \| '_ \/ __|  *
 *  | |__| | (_) | | |  | |_| | | | | (__| |_| | (_) | | | \__ \  *
 *   \_____|\___/  |_|   \__,_|_| |_|\___|\__|_|\___/|_| |_|___/  *
 *                                                                *
 ******************************************************************/

var mutex = sync.Mutex{}

func New() Fetcher {
	mutex.Lock()
	id := randStr(32)
	cookieStores[id] = map[string]kv{}
	mutex.Unlock()
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
	mutex.Lock()
	if cookieStores[fetcher.id][domain] == nil {
		cookieStores[fetcher.id][domain] = kv{}
	}
	cookieStores[fetcher.id][domain][name] = value
	mutex.Unlock()
}

func (fetcher Fetcher) SetCookies(domain string, cookies kv) {
	mutex.Lock()
	if cookieStores[fetcher.id][domain] == nil {
		cookieStores[fetcher.id][domain] = kv{}
	}
	for k, v := range cookies {
		cookieStores[fetcher.id][domain][k] = v
	}
	mutex.Unlock()
}

func (fetcher Fetcher) DeleteCookies(domain string) {
	mutex.Lock()
	if cookieStores[fetcher.id][domain] != nil {
		cookieStores[fetcher.id][domain] = kv{}
	}
	mutex.Unlock()
}

func (fetcher Fetcher) DeleteCookie(domain string, name string) {
	mutex.Lock()
	if cookieStores[fetcher.id][domain] != nil && len(cookieStores[fetcher.id][domain][name]) > 0 {
		delete(cookieStores[fetcher.id][domain], name)
	}
	mutex.Unlock()
}

func cookiesToHeader(cookies kv) string {
	arr := []string{}
	for k, v := range cookies {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(arr, "; ")
}

func headerToCookies(header []string) kv {
	cookies := kv{}
	for _, s := range header {
		arr := strings.SplitN(strings.Split(s, ";")[0], "=", 2)
		mutex.Lock()
		cookies[arr[0]] = arr[1]
		mutex.Unlock()
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
		StatusCode:    resp.StatusCode,
		StatusMessage: resp.Status,
		Headers:       resp.Header,
		Body:          body,
		RawBody:       rawBody,
		Error:         false,
		ErrorMessage:  "",
	}
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
	if strings.Contains(u, "localhost") {
		return "localhost"
	}
	return arr[len(arr)-2] + "." + arr[len(arr)-1]
}

func setChromeHeaders(headers http.Header) http.Header {

	if len(headers["Connection"]) == 0 {
		headers["Connection"] = []string{"keep-alive"}
	}

	if len(headers["sec-ch-ua"]) == 0 {
		headers["sec-ch-ua"] = []string{config.Sec_CH_UA}
	}

	if len(headers["sec-ch-ua-mobile"]) == 0 {
		headers["sec-ch-ua-mobile"] = []string{"?0"}
	}

	if len(headers["User-Agent"]) == 0 {
		headers["User-Agent"] = []string{defaultUserAgent}
	}

	if len(headers["Accept"]) == 0 {
		headers["Accept"] = []string{config.Accept}
	}

	if len(headers["Sec-Fetch-Site"]) == 0 {
		headers["Sec-Fetch-Site"] = []string{config.Sec_Fetch_Site}
	}

	if len(headers["Sec-Fetch-Mode"]) == 0 {
		headers["Sec-Fetch-Mode"] = []string{config.Sec_Fetch_Mode}
	}

	if len(headers["Sec-Fetch-User"]) == 0 {
		headers["Sec-Fetch-User"] = []string{"?1"}
	}

	if len(headers["Sec-Fetch-Dest"]) == 0 {
		headers["Sec-Fetch-Dest"] = []string{config.Sec_Fetch_Dest}
	}

	if len(headers["Accept-Encoding"]) == 0 {
		headers["Accept-Encoding"] = []string{config.AcceptEncoding}
	}

	if len(headers["Accept-Language"]) == 0 {
		headers["Accept-Language"] = []string{config.AcceptLanguage}
	}

	// if len(headers["If-None-Match"]) == 0 {
	// 	headers["If-None-Match"] = []string{fmt.Sprintf(`W/"2a6-%s/%s"`, randStr(11), randStr(15))}
	// }

	return headers
}

func (fetcher Fetcher) Get(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID = options.ClientHelloID

	if len(clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.Headers["User-Agent"] != nil && len(options.Headers["User-Agent"]) > 0 {
		userAgent = options.Headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	mutex.Lock()
	if len(options.Proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.Proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.Headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}
	mutex.Unlock()

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}
	if options.Headers == nil {
		options.Headers = http.Header{}
	}

	mutex.Lock()
	options.Headers = setChromeHeaders(options.Headers)
	mutex.Unlock()

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "GET",
		Header: options.Headers,
	})

	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}
	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func (fetcher Fetcher) Post(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID = options.ClientHelloID

	if len(clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.Headers["User-Agent"] != nil && len(options.Headers["User-Agent"]) > 0 {
		userAgent = options.Headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	if len(options.Proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.Proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	mutex.Lock()
	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.Headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}
	mutex.Unlock()

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}
	if options.Headers == nil {
		options.Headers = http.Header{}
	}

	options.Headers = setChromeHeaders(options.Headers)

	reader := bytes.NewReader(options.Body)

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "POST",
		Header: options.Headers,
		Body:   io.NopCloser(reader),
	})

	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}

	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func (fetcher Fetcher) Put(requestURL string, options RequestOptions) Response {
	var client http.Client
	var err error
	var clientHelloID = options.ClientHelloID

	if len(clientHelloID.Client) == 0 {
		clientHelloID = tls.HelloChrome_90
	}

	var userAgent string
	if options.Headers["User-Agent"] != nil && len(options.Headers["User-Agent"]) > 0 {
		userAgent = options.Headers["User-Agent"][0]
	} else {
		userAgent = defaultUserAgent
	}

	if len(options.Proxy) != 0 {
		client, err = cclient.NewClient(clientHelloID, userAgent, options.Proxy)
	} else {
		client, err = cclient.NewClient(clientHelloID, userAgent)
	}

	mutex.Lock()
	if len(fetcher.GetCookies(getDomain(requestURL))) != 0 {
		options.Headers["Cookie"] = []string{cookiesToHeader(fetcher.GetCookies(getDomain(requestURL)))}
	}
	mutex.Unlock()

	if err != nil {
		log.Fatal(err)
	}

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}
	if options.Headers == nil {
		options.Headers = http.Header{}
	}

	options.Headers = setChromeHeaders(options.Headers)

	reader := bytes.NewReader(options.Body)

	resp, err := client.Do(&http.Request{
		URL:    reqURL,
		Method: "PUT",
		Header: options.Headers,
		Body:   io.NopCloser(reader),
	})

	if err != nil {
		return Response{
			Error:        true,
			ErrorMessage: err.Error(),
		}
	}

	fetcher.SetCookies(getDomain(requestURL), headerToCookies(resp.Header["Set-Cookie"]))

	return parseResponse(resp)
}

func GetTLSParrot(parrotName string) tls.ClientHelloID {
	switch parrotName {
	case "Golang":
		return tls.HelloGolang
	case "Chrome_90":
		return tls.HelloChrome_90
	case "Chrome_83":
		return tls.HelloChrome_83
	case "Chrome_72":
		return tls.HelloChrome_72
	case "Chrome_70":
		return tls.HelloChrome_70
	case "Chrome_62":
		return tls.HelloChrome_62
	case "Chrome_58":
		return tls.HelloChrome_58
	case "Firefox_65":
		return tls.HelloFirefox_65
	case "Firefox_63":
		return tls.HelloFirefox_63
	case "Firefox_56":
		return tls.HelloFirefox_56
	case "Firefox_55":
		return tls.HelloFirefox_55
	default:
		return tls.HelloChrome_90
	}
}

// func main() {
// 	rand.Seed(time.Now().UnixNano())
// 	fetcher := New()
// 	res := fetcher.Post("https://api.opensea.io/graphql/", RequestOptions{
// 		Proxy:         "http://localhost:8866",
// 		ClientHelloID: getTLSParrot("Chrome_90"),
// 		Body:          []byte("{\"id\":\"EventHistoryPollQuery\",\"query\":\"query EventHistoryPollQuery(\\n  $archetype: ArchetypeInputType\\n  $categories: [CollectionSlug!]\\n  $chains: [ChainScalar!]\\n  $collections: [CollectionSlug!]\\n  $count: Int = 10\\n  $cursor: String\\n  $eventTimestamp_Gt: DateTime\\n  $eventTypes: [EventType!]\\n  $identity: IdentityInputType\\n  $showAll: Boolean = false\\n) {\\n  assetEvents(after: $cursor, archetype: $archetype, categories: $categories, chains: $chains, collections: $collections, eventTimestamp_Gt: $eventTimestamp_Gt, eventTypes: $eventTypes, first: $count, identity: $identity, includeHidden: true) {\\n    edges {\\n      node {\\n        assetBundle @include(if: $showAll) {\\n          relayId\\n          ...AssetCell_assetBundle\\n          ...bundle_url\\n          id\\n        }\\n        assetQuantity {\\n          asset @include(if: $showAll) {\\n            relayId\\n            assetContract {\\n              ...CollectionLink_assetContract\\n              id\\n            }\\n            ...AssetCell_asset\\n            ...asset_url\\n            collection {\\n              ...CollectionLink_collection\\n              id\\n            }\\n            id\\n          }\\n          ...quantity_data\\n          id\\n        }\\n        relayId\\n        eventTimestamp\\n        eventType\\n        customEventName\\n        offerExpired\\n        ...utilsAssetEventLabel\\n        devFee {\\n          asset {\\n            assetContract {\\n              chain\\n              id\\n            }\\n            id\\n          }\\n          quantity\\n          ...AssetQuantity_data\\n          id\\n        }\\n        devFeePaymentEvent {\\n          ...EventTimestamp_data\\n          id\\n        }\\n        fromAccount {\\n          address\\n          ...AccountLink_data\\n          id\\n        }\\n        price {\\n          quantity\\n          quantityInEth\\n          ...AssetQuantity_data\\n          id\\n        }\\n        endingPrice {\\n          quantity\\n          ...AssetQuantity_data\\n          id\\n        }\\n        seller {\\n          ...AccountLink_data\\n          id\\n        }\\n        toAccount {\\n          ...AccountLink_data\\n          id\\n        }\\n        winnerAccount {\\n          ...AccountLink_data\\n          id\\n        }\\n        ...EventTimestamp_data\\n        id\\n      }\\n    }\\n  }\\n}\\n\\nfragment AccountLink_data on AccountType {\\n  address\\n  config\\n  isCompromised\\n  user {\\n    publicUsername\\n    id\\n  }\\n  ...ProfileImage_data\\n  ...wallet_accountKey\\n  ...accounts_url\\n}\\n\\nfragment AssetCell_asset on AssetType {\\n  collection {\\n    name\\n    id\\n  }\\n  name\\n  ...AssetMedia_asset\\n  ...asset_url\\n}\\n\\nfragment AssetCell_assetBundle on AssetBundleType {\\n  assetQuantities(first: 2) {\\n    edges {\\n      node {\\n        asset {\\n          collection {\\n            name\\n            id\\n          }\\n          name\\n          ...AssetMedia_asset\\n          ...asset_url\\n          id\\n        }\\n        relayId\\n        id\\n      }\\n    }\\n  }\\n  name\\n  slug\\n}\\n\\nfragment AssetMedia_asset on AssetType {\\n  animationUrl\\n  backgroundColor\\n  collection {\\n    displayData {\\n      cardDisplayStyle\\n    }\\n    id\\n  }\\n  isDelisted\\n  imageUrl\\n  displayImageUrl\\n}\\n\\nfragment AssetQuantity_data on AssetQuantityType {\\n  asset {\\n    ...Price_data\\n    id\\n  }\\n  quantity\\n}\\n\\nfragment CollectionLink_assetContract on AssetContractType {\\n  address\\n  blockExplorerLink\\n}\\n\\nfragment CollectionLink_collection on CollectionType {\\n  name\\n  ...collection_url\\n  ...verification_data\\n}\\n\\nfragment EventTimestamp_data on AssetEventType {\\n  eventTimestamp\\n  transaction {\\n    blockExplorerLink\\n    id\\n  }\\n}\\n\\nfragment Price_data on AssetType {\\n  decimals\\n  imageUrl\\n  symbol\\n  usdSpotPrice\\n  assetContract {\\n    blockExplorerLink\\n    chain\\n    id\\n  }\\n}\\n\\nfragment ProfileImage_data on AccountType {\\n  imageUrl\\n  address\\n}\\n\\nfragment accounts_url on AccountType {\\n  address\\n  user {\\n    publicUsername\\n    id\\n  }\\n}\\n\\nfragment asset_url on AssetType {\\n  assetContract {\\n    address\\n    chain\\n    id\\n  }\\n  tokenId\\n}\\n\\nfragment bundle_url on AssetBundleType {\\n  slug\\n}\\n\\nfragment collection_url on CollectionType {\\n  slug\\n}\\n\\nfragment quantity_data on AssetQuantityType {\\n  asset {\\n    decimals\\n    id\\n  }\\n  quantity\\n}\\n\\nfragment utilsAssetEventLabel on AssetEventType {\\n  isMint\\n  eventType\\n}\\n\\nfragment verification_data on CollectionType {\\n  isMintable\\n  isSafelisted\\n  isVerified\\n}\\n\\nfragment wallet_accountKey on AccountType {\\n  address\\n}\\n\",\"variables\":{\"archetype\":null,\"categories\":null,\"chains\":null,\"collections\":[\"azuki\"],\"count\":100,\"cursor\":null,\"eventTimestamp_Gt\":\"2022-01-27T15:47:06.252223\",\"eventTypes\":[\"AUCTION_SUCCESSFUL\",\"AUCTION_CREATED\"],\"identity\":null,\"showAll\":true}}"),
// 		Headers: http.Header{
// 			"referer":          []string{"https://opensea.io"},
// 			"Content-Type":     []string{"application/json"},
// 			"X-API-KEY":        []string{"2f6f419a083c46de9d83ce3dbe7db601"},
// 			"X-VIEWER-ADDRESS": []string{"0x5e4867981dca0a4e41fd142454574dcad1c7bc99"},
// 			"x-signed-query":   []string{"34a59e7bd61fe3767d5f38a184f5ed424541ec63c6bdc0d7c996742460430f71"},
// 			"X-BUILD-ID":       []string{"FsBm9LH_S5rHfUOYXTuI5"},
// 		},
// 	})

// 	fmt.Println(res.Body)
// }
