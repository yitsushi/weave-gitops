package gitops_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fluxcd/pkg/gittestserver"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var uninstallParams gitops.UninstallParams

func checkFluxUninstallFailure() {
	fluxErrMsg := "flux uninstall failed"

	loggedMsg := ""
	logger.PrintlnStub = func(str string, args ...interface{}) {
		loggedMsg = fmt.Sprintf(str, args...)
	}

	fluxClient.UninstallStub = func(namespace string, dryRun bool) error {
		return errors.New(fluxErrMsg)
	}

	err := gitopsSrv.Uninstall(uninstallParams)

	Expect(loggedMsg).To(Equal(fmt.Sprintf("received error uninstalling flux: %q, continuing with uninstall", fluxErrMsg)))
	Expect(err).To(MatchError(gitops.UninstallError{}))
	Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
	Expect(fluxClient.UninstallCallCount()).To(Equal(1))
	namespace, dryRun := fluxClient.UninstallArgsForCall(0)
	Expect(namespace).To(Equal(wego.DefaultNamespace))
	Expect(dryRun).To(Equal(false))
}

func checkAppCRDUninstallFailure() {
	manifestsErrMsg := "gitops manifests uninstall failed"

	loggedMsg := ""
	logger.PrintfStub = func(str string, args ...interface{}) {
		loggedMsg = fmt.Sprintf(str, args...)
	}

	kubeClient.DeleteStub = func(ctx context.Context, manifest []byte) error {
		return errors.New(manifestsErrMsg)
	}

	err := gitopsSrv.Uninstall(uninstallParams)

	Expect(loggedMsg).To(ContainSubstring("error applying wego-app manifest"))
	Expect(err).To(MatchError(gitops.UninstallError{}))
	Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
	Expect(fluxClient.UninstallCallCount()).To(Equal(1))
	Expect(kubeClient.DeleteCallCount()).To(Equal(6))

	namespace, dryRun := fluxClient.UninstallArgsForCall(0)
	Expect(namespace).To(Equal(wego.DefaultNamespace))
	Expect(dryRun).To(Equal(false))
}

var _ = Describe("Uninstall", func() {
	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{
			GetWegoConfigStub: func(c context.Context, s string) (*kube.WegoConfig, error) {
				return &kube.WegoConfig{FluxNamespace: wego.DefaultNamespace, WegoNamespace: wego.DefaultNamespace}, nil
			},
		}
		logger = &loggerfakes.FakeLogger{}
		gitopsSrv = gitops.New(logger, fluxClient, kubeClient)

		uninstallParams = gitops.UninstallParams{
			Namespace: wego.DefaultNamespace,
			DryRun:    false,
		}
	})

	It("logs warning information if wego is not installed before proceeding", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
		Expect(fluxClient.UninstallCallCount()).To(Equal(1))

		loggedMsg := ""
		logger.PrintlnStub = func(str string, args ...interface{}) {
			loggedMsg = str
		}

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal("gitops is not fully installed... removing any partial installation\n"))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}
		loggedMsg = ""

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal("gitops is not fully installed... removing any partial installation\n"))
	})

	It("Does not log warning information if wego is installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		loggedMsg := ""
		logger.PrintlnStub = func(str string, args ...interface{}) {
			loggedMsg = str
		}

		Expect(gitopsSrv.Uninstall(uninstallParams)).Should(Succeed())
		Expect(loggedMsg).To(Equal(""))
	})

	It("Generates an error if flux uninstall fails with wego installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if flux uninstall fails with only flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if flux uninstall fails with partial or no flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}

		checkFluxUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with wego installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		}

		checkAppCRDUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with only flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}

		checkAppCRDUninstallFailure()
	})

	It("Generates an error if CRD uninstall fails with partial or no flux installed", func() {
		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}

		checkAppCRDUninstallFailure()
	})

	It("deletes weave gitops manifests", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		wegoAppManifests, err := manifests.GenerateManifests(manifests.Params{AppVersion: version.Version, Namespace: "default"})
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.DeleteCallCount()).To(Equal(len(wegoAppManifests)+1), "deletes all wego app manifests plus the app crd")
	})

	It("fails if we can't fetch the wego config", func() {
		kubeClient.GetWegoConfigReturns(nil, errors.New("error"))

		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err.Error()).Should(ContainSubstring("errors occurred during uninstall"))
	})

	It("avoid uninstalling flux when its namespace is different", func() {
		kubeClient.GetWegoConfigReturns(&kube.WegoConfig{FluxNamespace: "flux-namespace"}, nil)

		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.UninstallCallCount()).To(Equal(0))
	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			uninstallParams.DryRun = true
		})

		It("calls flux uninstall", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.UninstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.UninstallArgsForCall(0)
			Expect(namespace).To(Equal(wego.DefaultNamespace))
			Expect(dryRun).To(Equal(true))
		})

		It("does not call kube apply", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.DeleteCallCount()).To(Equal(0))
		})
	})
	FDescribe("Actual Uninstall tests", func() {
		var (
			gOps        gitops.GitopsService
			gitServer   *gittestserver.GitServer
			namespace   *corev1.Namespace
			localGitDir string
		)

		BeforeEach(func() {
			gitServer, err = gittestserver.NewTempGitServer()

			username := "test-user"
			password := "test-password"

			gitServer.Auth(username, password)
			gitServer = gitServer.AutoCreate()
			repoPath := "bar/test-reponame"

			Expect(gitServer.InitRepo("testdata/git/repo1", "main", repoPath)).To(Succeed())

			errc := make(chan error)
			go func() {
				errc <- gitServer.StartHTTP()
			}()

			select {
			case err := <-errc:
				Expect(err).NotTo(HaveOccurred())
				break
			case <-time.After(time.Second):
				break
			}

			addr := gitServer.HTTPAddressWithCredentials() + "/" + repoPath

			fluxClient := &fluxfakes.FakeFlux{}
			fluxClient.InstallReturns([]byte{}, nil)

			k8s, _, err := kube.NewKubeHTTPClientWithConfig(env.Rest, "test-context")
			Expect(err).NotTo(HaveOccurred())

			namespace = &corev1.Namespace{}
			namespace.Name = "kube-test-" + rand.String(5)
			Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
			Expect(k8s.Raw().Create(context.Background(), namespace))

			gOps = gitops.New(&loggerfakes.FakeLogger{}, fluxClient, k8s)

			params := gitops.InstallParams{
				Namespace:  namespace.Name,
				DryRun:     false,
				ConfigRepo: gitproviders.RepoURL{},
			}

			m, err := gOps.Install(params)
			Expect(err).NotTo(HaveOccurred())

			localGitDir, err = os.MkdirTemp("", "test-repo")
			Expect(err).NotTo(HaveOccurred())

			gg := wrapper.NewGoGit()

			_, err = gg.PlainCloneContext(context.Background(), localGitDir, false, &gogit.CloneOptions{
				URL: addr,
			})
			Expect(err).NotTo(HaveOccurred())

			gc := git.New(&http.BasicAuth{Username: username, Password: password}, gg)
			gc.Open(repoPath)

			// _, err = gc.Clone(context.Background(), localGitDir, addr, "main")
			// Expect(err).NotTo(HaveOccurred())

			gp := &gitprovidersfakes.FakeGitProvider{}
			gp.GetDefaultBranchReturns("main", nil)

			_, err = gOps.StoreManifests(gc, gp, params, m)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			gitServer.StopHTTP()
			Expect(os.RemoveAll(gitServer.Root())).To(Succeed())
			Expect(os.RemoveAll(localGitDir)).To(Succeed())
		})
		It("runs?", func() {
			entries, err := os.ReadDir(gitServer.Root())
			Expect(err).NotTo(HaveOccurred())

			for _, e := range entries {
				fmt.Println(e.Name())
			}

		})
	})
})
