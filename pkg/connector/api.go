package connector

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Connector interface {
	PortForward(obj client.Object, ports []string) (PortForwarder, error)
}

type PortForwarder interface {
	Stop()
}

type connector struct {
	ClientSet  *kubernetes.Clientset
	Config     *rest.Config
	RESTClient *rest.RESTClient
}

func setKubernetesDefaults(config *rest.Config) error {
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}

	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	}
	return rest.SetKubernetesDefaults(config)
}

func New(config *rest.Config) (Connector, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	if err := setKubernetesDefaults(config); err != nil {
		return nil, err
	}

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &connector{
		ClientSet:  clientSet,
		Config:     config,
		RESTClient: restClient,
	}, nil
}

func PortForward(config *rest.Config, obj client.Object, ports []string) (PortForwarder, error) {
	instance, err := New(config)
	if err != nil {
		return nil, err
	}

	return instance.PortForward(obj, ports)
}

func (c *connector) PortForward(obj client.Object, ports []string) (PortForwarder, error) {
	forwarder := &portForwarder{
		ClientSet:  c.ClientSet,
		Config:     c.Config,
		RESTClient: c.RESTClient,
		Object:     obj,
		Ports:      ports,
	}

	if err := forwarder.Start(); err != nil {
		return nil, err
	}

	return forwarder, nil
}
