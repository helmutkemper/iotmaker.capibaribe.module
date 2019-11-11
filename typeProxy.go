package iotmaker_capibaribe_module

import (
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

/*
pt_br: Recebe a lista de containers e endpoints para redirecionar cada endpoint de entrada
*/
type proxy struct {
	// pt_br: Quantidade máxima de testes antes de de uma falha ser aceita
	MaxAttemptToRescueLoop int `yaml:"maxAttemptToRescueLoop" json:"maxAttemptToRescueLoop"`

	// pt_br: ignora a porta de entrada de dados //todo: é isto mesmo?
	IgnorePort bool `yaml:"ignorePort" json:"ignorePort"`

	// pt_br: host do servidor para servidores com vários domínios
	Host string `yaml:"host" json:"host"`

	// escolha do tipo de load balancing
	LoadBalancing string `yaml:"loadBalancing" json:"loadBalancing"`

	// pt_br: path dentro do domínio.
	// quando definido, redireciona o path para um endereço específico
	Path string `yaml:"path" json:"path"`

	Header []headerMonitor `yaml:"header" json:"header"`

	// pt_br: lista de servidores secundários
	Servers []servers `yaml:"servers" json:"servers"`

	Analytics
}

func (el *proxy) VerifyHostPathToValidateRoute(host string) bool {
	return el.Host == host || el.Host == ""
}

func (el *proxy) SelectLoadBalance() (string, int) {
	if el.LoadBalancing == KLoadBalanceRandom {
		return el.random()
	} else if el.LoadBalancing == KLoadBalanceExecutionTime {
		return el.executionTime()
	} else if el.LoadBalancing == KLoadBalanceExecutionTimeAverage {
		return el.executionTimeAverage()
	}

	//if el.LoadBalancing == KLoadBalanceRoundRobin || el.LoadBalancing == ""
	return el.roundRobin()

}

func (el *proxy) VerifyPathAndHeaderInformationToValidateRoute(path string, w http.ResponseWriter, r *http.Request) bool {
	// simplified true table
	// | A | B | C | D | S |
	// |---|---|---|---|---|
	// | 1 | 1 | 1 | 1 | 1 |
	// | X | X | X | X | 0 |
	A := el.Path != ""
	B := len(el.Header) != 0
	C := el.Path == path
	D := el.VerifyHeaderMatchValueToRoute(w, r)
	return A && B && C && D

}

func (el *proxy) VerifyPathWithoutVerifyHeaderInformationToValidateRoute(path string) bool {
	A := el.Path == ""
	B := el.Path == path
	return A || B
}

func (el *proxy) VerifyHeaderInformationWithoutVerifyPathToValidateRoute(w http.ResponseWriter, r *http.Request) bool {
	A := len(el.Header) == 0
	B := el.VerifyHeaderMatchValueToRoute(w, r)
	return A || B
}

func (el *proxy) VerifyHeaderMatchValueToRoute(w http.ResponseWriter, r *http.Request) bool {
	for _, headerData := range el.Header {

		if headerData.Type == KHeaderTypeString && r.Header.Get(headerData.Key) == headerData.Value {
			return true
		} else if headerData.Type == KHeaderTypeRegExp {
			re := regexp.MustCompile(headerData.Value)
			if re.MatchString(r.Header.Get(headerData.Key)) == true {
				return true
			}
		}

	}

	return false
}

func (el *proxy) OnExecutionEndWithError(w http.ResponseWriter, r *http.Request, err error) {
	el.Analytics.OnExecutionEndWithError()
}

func (el *proxy) ModifyResponse(resp *http.Response) error {
	//b, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (el *proxy) executionTimeAverage() (string, int) {

	minTime := time.Duration(math.MaxInt64)
	keyToReturn := 0

	for serverKey, serverData := range el.Servers {
		if minTime > serverData.ExecutionDurationSuccessAverage {

			keyToReturn = serverKey
			minTime = serverData.ExecutionDurationSuccessAverage

		}
	}

	return el.Servers[keyToReturn].Host, keyToReturn
}

func (el *proxy) executionTime() (string, int) {

	minTime := time.Duration(math.MaxInt64)
	keyToReturn := 0

	for serverKey, serverData := range el.Servers {
		if minTime > serverData.ExecutionSuccessDurationMin {

			keyToReturn = serverKey
			minTime = serverData.ExecutionSuccessDurationMin

		}
	}

	return el.Servers[keyToReturn].Host, keyToReturn
}

func (el *proxy) roundRobin() (string, int) {

	for {

		randNumber := rand.Float64()
		for serverKey, serverData := range el.Servers {
			if randNumber <= serverData.Weight {
				return serverData.Host, serverKey
			}
		}

	}

}

func (el *proxy) random() (string, int) {
	randNumber := rand.Intn(len(el.Servers))
	return el.Servers[randNumber].Host, randNumber
}
