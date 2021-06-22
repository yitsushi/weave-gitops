package gitops_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var (
	fluxClient *fluxfakes.FakeFlux
	kubeClient *kubefakes.FakeKube

	gitopsSrv     gitops.GitopsService
	installParams gitops.InstallParams
)

var _ = BeforeEach(func() {
	fluxClient = &fluxfakes.FakeFlux{}
	kubeClient = &kubefakes.FakeKube{}
	gitopsSrv = gitops.New(fluxClient, kubeClient)

	installParams = gitops.InstallParams{
		Namespace: "wego-system",
		DryRun:    false,
	}
})

var _ = Describe("Install", func() {
	It("checks flux presence on the cluster", func() {
		kubeClient.FluxPresentStub = func() (bool, error) {
			return true, nil
		}

		_, err := gitopsSrv.Install(installParams)
		Expect(err).Should(HaveOccurred())
	})

	It("calls flux install", func() {
		_, err := gitopsSrv.Install(installParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.InstallCallCount()).To(Equal(1))

		namespace, dryRun := fluxClient.InstallArgsForCall(0)
		Expect(namespace).To(Equal("wego-system"))
		Expect(dryRun).To(Equal(false))
	})

	It("applies app crd", func() {
		_, err := gitopsSrv.Install(installParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.ApplyCallCount()).To(Equal(1))

		appCRD, namespace := kubeClient.ApplyArgsForCall(0)
		Expect(appCRD).To(ContainSubstring("kind: Application"))
		Expect(namespace).To(Equal("wego-system"))
	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			installParams.DryRun = true
			fluxClient.InstallStub = func(s string, b bool) ([]byte, error) {
				return []byte("manifests"), nil
			}
		})

		It("calls flux install", func() {
			manifests, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(manifests)).To(ContainSubstring("manifests"))

			Expect(fluxClient.InstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.InstallArgsForCall(0)
			Expect(namespace).To(Equal("wego-system"))
			Expect(dryRun).To(Equal(true))
		})

		It("appends app crd to flux install output", func() {
			manifests, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(string(manifests)).To(ContainSubstring("kind: Application"))
		})

		It("does not call kube apply", func() {
			_, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})
})
