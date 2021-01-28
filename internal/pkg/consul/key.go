package consul

import "github.com/hashicorp/consul/api"

func (s *Manager) SetKeyValue(key string, value []byte) error {
	kv := &api.KVPair{
		Key:   key,
		Flags: 0,
		Value: value,
	}
	_, err := s.consulClient.KV().Put(kv, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Manager) GetKeyValue(key string) (string, error) {
	kv, _, err := s.consulClient.KV().Get(key, nil)
	if err != nil {
		return "", err
	}
	if kv == nil {
		return "", err
	}
	return string(kv.Value), nil
}
