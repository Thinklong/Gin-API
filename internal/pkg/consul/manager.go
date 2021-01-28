package consul

import (
	"github.com/hashicorp/consul/api"
	"os"
	"strconv"
	"time"
)

type Info struct {
	ServiceName string
	ServiceID   string
	IP          string
	Port        int
	Timestamp   int
}
type Manager struct {
	serviceInfo  *Info
	consulClient *api.Client
	quitChan     chan os.Signal
	checker      *api.AgentServiceCheck
}

type List []Info

func getConsulConfig() *api.Config {
	consulConf := api.DefaultConfig()
	addr := os.Getenv("APP_CONSUL_ADDR")
	if addr != "" {
		consulConf.Address = addr
	} else {
		consulConf.Address = "127.0.0.1:8500"
	}
	return consulConf
}

func (s *Manager) Init(serviceName string, ip string, port int) {
	s.serviceInfo = &Info{
		ServiceName: serviceName,
		ServiceID:   "",
		IP:          ip,
		Port:        port,
		Timestamp:   time.Now().Nanosecond(),
	}
	s.consulClient, _ = api.NewClient(getConsulConfig())
	s.quitChan = make(chan os.Signal, 1)
}

func (s *Manager) getServiceId() string {
	return s.serviceInfo.ServiceName + "-" + s.serviceInfo.IP + "-" + strconv.Itoa(s.serviceInfo.Port)
}

func (s *Manager) AddChecker(checker *api.AgentServiceCheck) {
	s.checker = checker
}

func (s *Manager) Start() error {
	var tags []string
	service := &api.AgentServiceRegistration{
		ID:      s.getServiceId(),
		Name:    s.serviceInfo.ServiceName,
		Port:    s.serviceInfo.Port,
		Address: s.serviceInfo.IP,
		Tags:    tags,
		Check:   s.checker,
	}
	err := s.consulClient.Agent().ServiceRegister(service)
	if err != nil {
		return err
	}
	go s.WaitToUnDeregisterService()
	return nil
}

func (s *Manager) Close() {
	s.quitChan <- os.Kill
}

func (s *Manager) WaitToUnDeregisterService() {
	<-s.quitChan
	if s.consulClient == nil {
		return
	}
	if err := s.consulClient.Agent().ServiceDeregister(s.getServiceId()); err != nil {
		println("deregister service failed, " + err.Error())
	}
}

func (s *Manager) Discover(serviceName string) (*List, error) {
	servicesData, _, err := s.consulClient.Health().Service(serviceName, "", true,
		&api.QueryOptions{})
	if err != nil {
		return nil, err
	}
	return parseServices(serviceName, servicesData)
}

func parseServices(serviceName string, servicesData []*api.ServiceEntry) (*List, error) {
	var serviceList List
	for _, entry := range servicesData {
		if serviceName != entry.Service.Service {
			continue
		}
		for _, health := range entry.Checks {
			if health.ServiceName != serviceName || health.Status != "passing" {
				continue
			}
			var node = Info{
				ServiceName: entry.Service.Service,
				ServiceID:   entry.Service.ID,
				IP:          entry.Service.Address,
				Port:        entry.Service.Port,
				Timestamp:   time.Now().Nanosecond(),
			}
			serviceList = append(serviceList, node)
		}
	}
	return &serviceList, nil
}
