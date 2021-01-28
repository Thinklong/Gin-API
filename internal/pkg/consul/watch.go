package consul

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

func (s *Manager) WatchKey(key string, callback func(key string, value []byte)) error {
	var data = `{"type":"key", "key":"` + key + `"}`
	param, err := parseParam(data)
	if err != nil {
		return err
	}
	plan, _ := watch.Parse(param)
	plan.HybridHandler = func(blockParamVal watch.BlockingParamVal, val interface{}) {
		var v []byte
		if val != nil {
			v = (val).(*api.KVPair).Value
		}
		callback(key, v)
	}
	return plan.Run(getConsulConfig().Address)
}

func (s *Manager) WatchService(service string, callback func(service string, serviceList *List)) error {
	var data = `{"type":"service", "service":"` + service + `"}`
	param, err := parseParam(data)
	if err != nil {
		return err
	}
	plan, _ := watch.Parse(param)
	plan.HybridHandler = func(blockParamVal watch.BlockingParamVal, val interface{}) {
		var v []*api.ServiceEntry
		if val != nil {
			v = (val).([]*api.ServiceEntry)
		}
		ss, _ := parseServices(service, v)
		callback(service, ss)
	}
	return plan.Run(getConsulConfig().Address)
}

func parseParam(data string) (map[string]interface{}, error) {
	var param map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader([]byte(data)))
	if err := dec.Decode(&param); err != nil {
		return nil, err
	}
	return param, nil
}
