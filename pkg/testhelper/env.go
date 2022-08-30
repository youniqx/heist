package testhelper

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	//nolint:golint,stylecheck
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeOutAfter       = 1 * time.Minute
	pollInterval       = 200 * time.Millisecond
	selfLinkSplitCount = 6
)

var ErrUnknownResource error = fmt.Errorf("unknown resource")

type KubernetesTestHelper interface {
	Create(objects ...client.Object)
	WaitUntilExists(objects ...client.Object)
	CleanupCreatedObject()
	DeleteIfPresent(obj ...client.Object)
	Object(obj client.Object) AsyncAssertion
	Heist() *heist.Clientset
	VaultConfigSpec(namespace string, name string) AsyncAssertion
}

type testHelper struct {
	Dynamic   dynamic.Interface
	Client    client.Client
	ClientSet *heist.Clientset

	CreatedObjects []client.Object
}

func (t *testHelper) Create(objects ...client.Object) {
	for _, obj := range objects {
		Expect(t.Client.Create(context.TODO(), obj)).To(Succeed())
		t.WaitUntilExists(obj)
		t.CreatedObjects = append(t.CreatedObjects, obj)
	}
}

func (t *testHelper) WaitUntilExists(objects ...client.Object) {
	for _, obj := range objects {
		Eventually(func() error {
			return t.Client.Get(context.TODO(), client.ObjectKeyFromObject(obj), obj)
		}, timeOutAfter, pollInterval).ShouldNot(HaveOccurred())
	}
}

func (t *testHelper) CleanupCreatedObject() {
	t.DeleteIfPresent(t.CreatedObjects...)
	t.CreatedObjects = nil
}

func (t *testHelper) VaultConfigSpec(namespace string, name string) AsyncAssertion {
	return Eventually(func() *heistv1alpha1.VaultClientConfigSpec {
		clientConfig, err := t.ClientSet.HeistV1alpha1().
			VaultClientConfigs(namespace).
			Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		return &clientConfig.Spec
	}, timeOutAfter, pollInterval)
}

func (t *testHelper) Heist() *heist.Clientset {
	return t.ClientSet
}

func (t *testHelper) DeleteIfPresent(objects ...client.Object) {
	for _, obj := range objects {
		err := t.Client.Get(context.TODO(), client.ObjectKeyFromObject(obj), obj)
		if err != nil {
			continue
		}

		Expect(t.Client.Delete(context.TODO(), obj)).To(Succeed())
		Eventually(func() error {
			return t.Client.Get(context.TODO(), client.ObjectKeyFromObject(obj), obj)
		}, timeOutAfter, pollInterval).Should(HaveOccurred())
	}
}

var selfLinkRegex = regexp.MustCompile("/apis/([^/]+)/([^/]+)/namespaces/([^/]+)/([^/]+)/([^/]+)")

func (t *testHelper) Object(obj client.Object) AsyncAssertion {
	return Eventually(func() *unstructured.Unstructured {
		result, err := t.fetchObject(obj)
		if err != nil {
			fmt.Printf("wantErr while fetching obj %s: %s\n", obj.GetName(), err)
		}
		return result
	}, timeOutAfter, pollInterval)
}

func (t *testHelper) fetchObject(obj client.Object) (*unstructured.Unstructured, error) {
	gvr, err := parseGroupVersionResource(obj)
	if err != nil {
		return nil, err
	}
	return t.Dynamic.
		Resource(gvr).
		Namespace(obj.GetNamespace()).
		Get(context.TODO(), obj.GetName(), metav1.GetOptions{}, "status")
}

func (t *testHelper) deleteObject(obj client.Object) error {
	gvr, err := parseGroupVersionResource(obj)
	if err != nil {
		return err
	}
	return t.Dynamic.
		Resource(gvr).
		Namespace(obj.GetNamespace()).
		Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{}, "status")
}

func parseGroupVersionResource(obj client.Object) (schema.GroupVersionResource, error) {
	if link := obj.GetSelfLink(); link != "" {
		matches := selfLinkRegex.FindStringSubmatch(link)
		if len(matches) != selfLinkSplitCount {
			return schema.GroupVersionResource{}, fmt.Errorf("referenced object does not have a self link")
		}
		gvr := schema.GroupVersionResource{
			Group:    matches[1],
			Version:  matches[2],
			Resource: matches[4],
		}
		return gvr, nil
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	resourceKind, err := getPluralNaming(gvk.Kind)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("couln't get plural name: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resourceKind,
	}

	return gvr, nil
}

func getPluralNaming(singularName string) (pluralName string, err error) {
	switch strings.ToLower(singularName) {
	case "vaultbinding":
		fallthrough
	case "vaultcertificaterole":
		fallthrough
	case "vaultclientconfig":
		fallthrough
	case "vaultkvsecretengine":
		fallthrough
	case "vaultkvsecret":
		fallthrough
	case "vaultsyncsecret":
		fallthrough
	case "vaulttransitengine":
		fallthrough
	case "vaulttransitkey":
		return fmt.Sprintf("%ss", strings.ToLower(singularName)), nil
	case "vaultcertificateauthority":
		regex := regexp.MustCompile("y$")
		return regex.ReplaceAllString(strings.ToLower(singularName), "ies"), nil
	}

	return "", fmt.Errorf("plural name of resource %s cannot be resolved: %w", singularName, ErrUnknownResource)
}

func New(cfg *rest.Config, k8sClient client.Client) KubernetesTestHelper {
	return &testHelper{
		ClientSet: heist.NewForConfigOrDie(cfg),
		Dynamic:   dynamic.NewForConfigOrDie(cfg),
		Client:    k8sClient,
	}
}
