package cclient

import (
	"github.com/Jishrocks/SneakFetch/http"

	"golang.org/x/net/proxy"

	utls "github.com/Jishrocks/SneakFetch/utls"
)

func NewClient(clientHello utls.ClientHelloID, userAgent string, proxyUrl ...string) (http.Client, error) {
	if len(proxyUrl) > 0 && len(proxyUrl) > 0 {
		dialer, err := newConnectDialer(proxyUrl[0], userAgent)
		if err != nil {
			return http.Client{}, err
		}
		return http.Client{
			Transport: newRoundTripper(clientHello, dialer),
		}, nil
	} else {
		return http.Client{
			Transport: newRoundTripper(clientHello, proxy.Direct),
		}, nil
	}
}
