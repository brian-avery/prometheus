//  Copyright Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package kube

import (
	"context"
	"fmt"
	"net/http"

	istioinformer "github.com/maistra/xns-informer/pkg/generated/istio"
	kubeinformer "github.com/maistra/xns-informer/pkg/generated/kube"
	serviceapisinformer "github.com/maistra/xns-informer/pkg/generated/serviceapis"
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	"google.golang.org/grpc/credentials"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeVersion "k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/cmd/util"
	serviceapisclient "sigs.k8s.io/service-apis/pkg/client/clientset/versioned"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	memberroll "istio.io/istio/pkg/servicemesh/controller"
	"istio.io/pkg/version"
)

var _ ExtendedClient = MockClient{}

type MockPortForwarder struct {
}

func (m MockPortForwarder) Start() error {
	return nil
}

func (m MockPortForwarder) Address() string {
	return "localhost:3456"
}

func (m MockPortForwarder) Close() {
}

func (m MockPortForwarder) WaitForStop() {
}

var _ PortForwarder = MockPortForwarder{}

// MockClient for tests that rely on kube.Client.
type MockClient struct {
	kubernetes.Interface
	RestClient *rest.RESTClient
	// Results is a map of podName to the results of the expected test on the pod
	Results          map[string][]byte
	DiscoverablePods map[string]map[string]*v1.PodList
	RevisionValue    string
	ConfigValue      *rest.Config
	IstioVersions    *version.MeshInfo
}

func (c MockClient) Istio() istioclient.Interface {
	panic("not used in mock")
}

func (c MockClient) ServiceApis() serviceapisclient.Interface {
	panic("not used in mock")
}

func (c MockClient) IstioInformer() istioinformer.SharedInformerFactory {
	panic("not used in mock")
}

func (c MockClient) ServiceApisInformer() serviceapisinformer.SharedInformerFactory {
	panic("not used in mock")
}

func (c MockClient) Metadata() metadata.Interface {
	panic("not used in mock")
}

func (c MockClient) KubeInformer() kubeinformer.SharedInformerFactory {
	panic("not used in mock")
}

func (c MockClient) DynamicInformer() xnsinformers.DynamicSharedInformerFactory {
	panic("not used in mock")
}

func (c MockClient) MetadataInformer() xnsinformers.MetadataSharedInformerFactory {
	panic("not used in mock")
}

func (c MockClient) RunAndWait(stop <-chan struct{}) {
	panic("not used in mock")
}

func (c MockClient) Kube() kubernetes.Interface {
	return c.Interface
}

func (c MockClient) DynamicClient() dynamic.Interface {
	panic("not used in mock")
}

func (c MockClient) MetadataClient() metadata.Interface {
	panic("not used in mock")
}

func (c MockClient) SetNamespaces(namespaces ...string) {
	panic("not used in mock")
}

func (c MockClient) AddMemberRoll(namespace, memberRollName string) error {
	panic("not used in mock")
}

func (c MockClient) GetMemberRoll() memberroll.MemberRollController {
	panic("not used in mock")
}

func (c MockClient) AllDiscoveryDo(_ context.Context, _, _ string) (map[string][]byte, error) {
	return c.Results, nil
}

func (c MockClient) EnvoyDo(_ context.Context, podName, _, _, _ string, _ []byte) ([]byte, error) {
	results, ok := c.Results[podName]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve Pod: pods %q not found", podName)
	}
	return results, nil
}

func (c MockClient) RESTConfig() *rest.Config {
	return c.ConfigValue
}

func (c MockClient) GetIstioVersions(_ context.Context, _ string) (*version.MeshInfo, error) {
	return c.IstioVersions, nil
}

func (c MockClient) PodsForSelector(_ context.Context, namespace string, labelSelectors ...string) (*v1.PodList, error) {
	podsForNamespace, ok := c.DiscoverablePods[namespace]
	if !ok {
		return &v1.PodList{}, nil
	}
	var allPods v1.PodList
	for _, selector := range labelSelectors {
		pods, ok := podsForNamespace[selector]
		if !ok {
			return &v1.PodList{}, nil
		}
		allPods.TypeMeta = pods.TypeMeta
		allPods.ListMeta = pods.ListMeta
		allPods.Items = append(allPods.Items, pods.Items...)
	}
	return &allPods, nil
}

func (c MockClient) Revision() string {
	return c.RevisionValue
}

func (c MockClient) REST() rest.Interface {
	panic("not implemented by mock")
}

func (c MockClient) ApplyYAMLFiles(string, ...string) error {
	panic("not implemented by mock")
}

func (c MockClient) ApplyYAMLFilesDryRun(string, ...string) error {
	panic("not implemented by mock")
}

// CreatePerRPCCredentials -- when implemented -- mocks per-RPC credentials (bearer token)
func (c MockClient) CreatePerRPCCredentials(ctx context.Context, tokenNamespace, tokenServiceAccount string, audiences []string,
	expirationSeconds int64) (credentials.PerRPCCredentials, error) {
	panic("not implemented by mock")
}

func (c MockClient) DeleteYAMLFiles(string, ...string) error {
	panic("not implemented by mock")
}

func (c MockClient) DeleteYAMLFilesDryRun(string, ...string) error {
	panic("not implemented by mock")
}

func (c MockClient) Ext() clientset.Interface {
	panic("not implemented by mock")
}

func (c MockClient) Dynamic() dynamic.Interface {
	panic("not implemented by mock")
}

func (c MockClient) GetKubernetesVersion() (*kubeVersion.Info, error) {
	return &kubeVersion.Info{
		Major: "1",
		Minor: "16",
	}, nil
}

func (c MockClient) GetIstioPods(_ context.Context, _ string, _ map[string]string) ([]v1.Pod, error) {
	return nil, fmt.Errorf("TODO MockClient doesn't implement IstioPods")
}

func (c MockClient) PodExec(_, _, _ string, _ string) (string, string, error) {
	return "", "", fmt.Errorf("TODO MockClient doesn't implement exec")
}

func (c MockClient) PodLogs(_ context.Context, _ string, _ string, _ string, _ bool) (string, error) {
	return "", fmt.Errorf("TODO MockClient doesn't implement logs")
}

func (c MockClient) NewPortForwarder(_, _, _ string, _, _ int) (PortForwarder, error) {
	return MockPortForwarder{}, nil
}

// UtilFactory mock's kubectl's utility factory.  This code sets up a fake factory,
// similar to the one in https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/describe/describe_test.go
func (c MockClient) UtilFactory() util.Factory {
	tf := cmdtesting.NewTestFactory()
	_, _, codec := cmdtesting.NewExternalScheme()
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     cmdtesting.DefaultHeader(),
			Body: cmdtesting.ObjBody(codec,
				cmdtesting.NewInternalType("", "", "foo"))},
	}
	return tf
}