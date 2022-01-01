package gitops

import (
	"github.com/weaveworks/weave-gitops/core/repository"
)

const (
	webhookSourceAlertFileName    = "webhook-source-alert.yaml"
	webhookSourceProviderFileName = "webhook-source-provider.yaml"
)

const (
	webhookSourceProvider = "webhook-source-provider"
)

//type ApplicationRuntime struct {
//	serviceName string
//	namespace   string
//}
//
//func (ar ApplicationRuntime) path(fileName string) string {
//	return fmt.Sprintf("%s/gitops/%s", gitops.baseDir, fileName)
//}
//
//func (ar ApplicationRuntime) Files() ([]repository.File, error) {
//var files []File
//
//address := url.URL{
//	Host: "http",
//	Path: fmt.Sprintf("%s.%s.svc.cluster.local/gitops/source/event/", ar.serviceName, ar.namespace),
//}
//
//provider := notificationv1.Provider{
//	TypeMeta: metav1.TypeMeta{
//		Kind:       notificationv1.ProviderKind,
//		APIVersion: notificationv1.GroupVersion.String(),
//	},
//	ObjectMeta: metav1.ObjectMeta{
//		Name:      webhookSourceProvider,
//		Namespace: ar.namespace,
//	},
//	Spec: notificationv1.ProviderSpec{
//		Type:    notificationv1.GenericProvider,
//		Address: address.String(),
//	},
//}
//
//providerJson, err := json.Marshal(&provider)
//if err != nil {
//	return nil, fmt.Errorf("webhook provider marshal to json: %s", provider.Name)
//}
//
//providerYaml, err := yaml.JSONToYAML(providerJson)
//if err != nil {
//	return nil, fmt.Errorf("webhook provider marshal to yaml: %s", provider.Name)
//}
//
//files = append(files, File{Path: ar.path(webhookSourceProviderFileName), Data: providerYaml})
//
//alert := notificationv1.Alert{
//	TypeMeta: metav1.TypeMeta{
//		Kind:       notificationv1.AlertKind,
//		APIVersion: notificationv1.GroupVersion.String(),
//	},
//	ObjectMeta: metav1.ObjectMeta{},
//	Spec: notificationv1.AlertSpec{
//		EventSeverity: "info",
//	},
//	Status: notificationv1.AlertStatus{},
//}

//return nil, nil
//}

type ClusterRuntime struct {
}

func (ar ClusterRuntime) Files() ([]repository.File, error) {
	return nil, nil
}
