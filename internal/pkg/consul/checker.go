package consul

import "github.com/hashicorp/consul/api"

func NewHTTPChecker(url, interval, timeout string) *api.AgentServiceCheck {
	return &api.AgentServiceCheck{
		HTTP:     url,
		Interval: interval,
		Timeout:  timeout,
	}
}

func NewTCPChecker(addr, interval, timeout string) *api.AgentServiceCheck {
	return &api.AgentServiceCheck{
		TCP:      addr,
		Interval: interval,
		Timeout:  timeout,
	}
}
