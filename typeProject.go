package iotmaker_capibaribe_module

import (
	"github.com/helmutkemper/seelog"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type Project struct {
	ListenAndServer   ListenAndServer `yaml:"listenAndServer"   json:"listenAndServer"`
	Sll               ssl             `yaml:"ssl"               json:"ssl"`
	Proxy             []proxy         `yaml:"proxy"             json:"proxy"`
	DebugServerEnable bool            `yaml:"debugServerEnable" json:"debugServerEnable"`
	HealthCheck       healthCheck     `yaml:"healthCheck"       json:"healthCheck"`
	AnalyticsCheck    analyticsCheck  `yaml:"analyticsCheck"    json:"analyticsCheck"`
	Listen            Listen          `yaml:"-"                 json:"-"`
	waitGroup         int             `yaml:"-"                 json:"-"`
}

func (el *Project) WaitAddDelta() {
	el.waitGroup += 1
}

func (el *Project) WaitDone() {
	el.waitGroup -= 1
}

func (el *Project) HandleFunc(w http.ResponseWriter, r *http.Request) {
	var err error
	var host = r.Host
	var path = r.URL.Path
	var re *regexp.Regexp
	var hostServer string
	var serverKey int
	var loopCounter = 0

	el.WaitAddDelta()

	defer el.WaitDone()

	if el.Proxy != nil {

		for proxyKey, proxyData := range el.Proxy {

			if proxyData.IgnorePort == true {
				if re, err = regexp.Compile(KIgnorePortRegExp); err != nil {
					HandleCriticalError(err)
					return
				}

				host = re.ReplaceAllString(host, "$1")
			}

			if proxyData.VerifyHostPathToValidateRoute(host) == true {

				if el.HealthCheck.VerifyPathToValidatePathIntoHost(path) == true {
					el.HealthCheck.WriteDataToOutputEndpoint(w, r)
					return
				}

				if el.AnalyticsCheck.VerifyPathToValidatePathIntoHost(path) == true {
					el.AnalyticsCheck.WriteDataToOutputEndpoint(w, &proxyData)
					return
				}

				// simplified true table
				// | A | B | C | S |
				// |---|---|---|---|
				// | X | X | 1 | 1 |
				// | X | 1 | X | 1 |
				// | 1 | X | X | 1 |
				// | 0 | 0 | 0 | 0 |
				A := proxyData.VerifyPathAndHeaderInformationToValidateRoute(path, w, r)
				B := proxyData.VerifyPathWithoutVerifyHeaderInformationToValidateRoute(path)
				C := proxyData.VerifyHeaderInformationWithoutVerifyPathToValidateRoute(w, r)
				if !(A || B || C) {
					continue
				}

				for {
					// Check the maximum number of interactions of the route from the proxy to prevent an infinite loop
					loopCounter += 1
					if loopCounter > el.Proxy[proxyKey].MaxAttemptToRescueLoop {
						_ = seelog.Critical("todas as rotas deram erro. final.")
						// fixme: colocar o que fazer no erro de todas as rotas
						return
					}

					hostServer, serverKey = proxyData.SelectLoadBalance()

					// Prepare the reverse proxy
					rpURL, err := url.Parse(hostServer)
					if err != nil {
						HandleCriticalError(err)
					}

					proxy := httputil.NewSingleHostReverseProxy(rpURL)
					proxy.ErrorLog = log.New(DebugLogger{}, "", 0)
					proxy.Transport = &transport{RoundTripper: http.DefaultTransport, Project: el}
					// Prepare the statistics of the TotalErrorsCounter and successes of the route in the reverse proxy
					proxy.ErrorHandler = el.Proxy[proxyKey].OnExecutionEndWithError

					//todo: implementar
					//proxy.ModifyResponse = proxyData.ModifyResponse

					// Run the route and measure execution time
					el.Proxy[proxyKey].Servers[serverKey].OnExecutionStart()
					el.Proxy[proxyKey].OnExecutionStart()
					proxy.ServeHTTP(w, r)

					// Verify error and continue to select a new route in case of error
					if el.Proxy[proxyKey].GetLastRoundError() == true {
						el.Proxy[proxyKey].Servers[serverKey].OnExecutionEndWithError()
						_ = seelog.Critical("todas as rotas deram erro. testando novamente")
						continue
					}

					// Statistics of successes of the route
					el.Proxy[proxyKey].OnExecutionEndWithSuccess()
					el.Proxy[proxyKey].Servers[serverKey].OnExecutionEndWithSuccess()
					_ = seelog.Critical("rota ok")
					return

				}
			}
		}
	}
}
