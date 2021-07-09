package app

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("List", func() {
	It("lists all apps", func() {
		kubeClient.GetApplicationsStub = func(ctx context.Context, namespace string) ([]wego.Application, error) {
			return []wego.Application{{
				Spec: wego.ApplicationSpec{Path: "bar"},
			}}, nil
		}

		a, err := appSrv.Get(ctx, types.NamespacedName{Name: defaultParams.Name, Namespace: defaultParams.Namespace})
		Expect(err).ShouldNot(HaveOccurred())

		Expect(a.Spec.Path).To(Equal("bar"))
	})
})
