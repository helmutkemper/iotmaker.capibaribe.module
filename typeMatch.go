package iotmaker_capibaribe_module

type match struct {
	Status []status `yaml:"status"  json:"status"`
	Header []header `yaml:"header"  json:"header"`
	Body   []string `yaml:"body"    json:"body"`
}
