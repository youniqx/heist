package connector

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type portForwarder struct {
	ClientSet    *kubernetes.Clientset
	Config       *rest.Config
	RESTClient   *rest.RESTClient
	Object       client.Object
	Ports        []string
	ErrorChannel chan error
	StopChannel  chan struct{}
	ReadyChannel chan struct{}
}

func (p *portForwarder) Start() error {
	forwardablePod, err := p.getForwardablePod(p.Object)
	if err != nil {
		return err
	}

	ports, err := determineForwardablePorts(p.Object, forwardablePod, p.Ports)
	if err != nil {
		return err
	}

	p.StopChannel = make(chan struct{}, 1)
	p.ReadyChannel = make(chan struct{})

	if forwardablePod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", forwardablePod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		if p.StopChannel != nil {
			close(p.StopChannel)
		}
	}()

	req := p.RESTClient.Post().
		Resource("pods").
		Namespace(forwardablePod.Namespace).
		Name(forwardablePod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(p.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	address := []string{"localhost"}
	fw, err := portforward.NewOnAddresses(dialer, address, ports, p.StopChannel, p.ReadyChannel, nil, nil)
	if err != nil {
		return err
	}

	p.ErrorChannel = make(chan error)
	go func() {
		defer close(p.ErrorChannel)
		p.ErrorChannel <- fw.ForwardPorts()
	}()

	for range p.ReadyChannel {
	}

	return nil
}

func (p *portForwarder) Stop() {
	p.StopChannel <- struct{}{}
}

func (p *portForwarder) getForwardablePod(obj client.Object) (*corev1.Pod, error) {
	//nolint:gocritic
	switch t := obj.(type) {
	case *corev1.Pod:
		return t, nil
	}

	clientset, err := v1.NewForConfig(p.Config)
	if err != nil {
		return nil, err
	}

	namespace, selector, err := selectorsForObject(obj)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %w", obj, err)
	}
	pod, _, err := getFirstPod(clientset, namespace, selector.String())
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func determineForwardablePorts(obj client.Object, forwardablePod *corev1.Pod, ports []string) ([]string, error) {
	switch t := obj.(type) {
	case *corev1.Service:
		return translateServicePortToTargetPort(ports, *t, *forwardablePod)
	default:
		return convertPodNamedPortToNumber(ports, *forwardablePod)
	}
}

func translateServicePortToTargetPort(ports []string, svc corev1.Service, pod corev1.Pod) ([]string, error) {
	var translated []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		portnum, err := strconv.Atoi(remotePort)
		if err != nil {
			svcPort, err := lookupServicePortNumberByName(svc, remotePort)
			if err != nil {
				return nil, err
			}
			portnum = int(svcPort)

			if localPort == remotePort {
				localPort = strconv.Itoa(portnum)
			}
		}
		if portnum < 0 || portnum > 65535 {
			return nil, fmt.Errorf("port %s is not a valid port number", port)
		}

		// secured by out-of-bounds check in line above
		//nolint:gosec
		containerPort, err := lookupContainerPortNumberByServicePort(svc, pod, int32(portnum))
		if err != nil {
			// can't resolve a named port, or Service did not declare this port, return an error
			return nil, err
		}

		// convert the resolved target port back to a string
		remotePort = strconv.Itoa(int(containerPort))

		if localPort != remotePort {
			translated = append(translated, fmt.Sprintf("%s:%s", localPort, remotePort))
		} else {
			translated = append(translated, remotePort)
		}
	}
	return translated, nil
}

func lookupContainerPortNumberByServicePort(svc corev1.Service, pod corev1.Pod, port int32) (int32, error) {
	for _, svcportspec := range svc.Spec.Ports {
		if svcportspec.Port != port {
			continue
		}
		if svc.Spec.ClusterIP == corev1.ClusterIPNone {
			return port, nil
		}
		if svcportspec.TargetPort.Type == intstr.Int {
			if svcportspec.TargetPort.IntValue() == 0 {
				// targetPort is omitted, and the IntValue() would be zero
				return svcportspec.Port, nil
			}
			return int32(svcportspec.TargetPort.IntValue()), nil
		}
		return lookupContainerPortNumberByName(pod, svcportspec.TargetPort.String())
	}
	return port, fmt.Errorf("service %s does not have a service port %d", svc.Name, port)
}

func lookupContainerPortNumberByName(pod corev1.Pod, name string) (int32, error) {
	for _, ctr := range pod.Spec.Containers {
		for _, ctrportspec := range ctr.Ports {
			if ctrportspec.Name == name {
				return ctrportspec.ContainerPort, nil
			}
		}
	}

	return int32(-1), fmt.Errorf("pod '%s' does not have a named port '%s'", pod.Name, name)
}

func lookupServicePortNumberByName(svc corev1.Service, name string) (int32, error) {
	for _, svcportspec := range svc.Spec.Ports {
		if svcportspec.Name == name {
			return svcportspec.Port, nil
		}
	}

	return int32(-1), fmt.Errorf("service '%s' does not have a named port '%s'", svc.Name, name)
}

func convertPodNamedPortToNumber(ports []string, pod corev1.Pod) ([]string, error) {
	var converted []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		containerPortStr := remotePort
		_, err := strconv.Atoi(remotePort)
		if err != nil {
			containerPort, err := lookupContainerPortNumberByName(pod, remotePort)
			if err != nil {
				return nil, err
			}

			containerPortStr = strconv.Itoa(int(containerPort))
		}

		if localPort != remotePort {
			converted = append(converted, fmt.Sprintf("%s:%s", localPort, containerPortStr))
		} else {
			converted = append(converted, containerPortStr)
		}
	}

	return converted, nil
}

const localRemotePortPair = 2

func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == localRemotePortPair {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}
