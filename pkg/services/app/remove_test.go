package app

import (
	//	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"

	"sigs.k8s.io/yaml"
)

var application wego.Application

const helmSource = `
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  name: loki
  namespace: wego-system
spec:
  interval: 30s
  url: https://charts.kube-ops.io
`

const helmGoatForHelmSource = `
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: loki
  namespace: wego-system
spec:
  chart:
    spec:
      chart: loki
      sourceRef:
        kind: HelmRepository
        name: loki
  install: {}
  interval: 5m0s
`

const helmGoatForGitSource = `
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: my-helm-app
  namespace: wego-system
spec:
  chart:
    spec:
      chart: ./hello-world
      sourceRef:
        kind: GitRepository
        name: my-helm-app
  install: {}
  interval: 5m0s
`

const gitSourceForHelmAppRepo = `
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: my-helm-app
  namespace: wego-system
spec:
  interval: 30s
  ref:
    branch: main
  secretRef:
    name: weave-gitops-kind-kind-wego-fork-test
  url: ssh://git@github.com/user/wego-fork-test.git
`

const gitSourceForHelmConfigRepo = `
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: external
  namespace: wego-system
spec:
  interval: 30s
  ref:
    branch: main
  secretRef:
    name: weave-gitops-kind-kind-external
  url: ssh://git@github.com/user/external.git
`

const kustForHelmAppsDir = `
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: my-helm-app-apps-dir
  namespace: wego-system
spec:
  interval: 1m0s
  path: ./apps/my-helm-app
  prune: true
  sourceRef:
    kind: GitRepository
    name: external
  validation: client
`

const kustForHelmTargetDir = `
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: kind-kind-my-helm-app
  namespace: wego-system
spec:
  interval: 1m0s
  path: ./targets/kind-kind/my-helm-app
  prune: true
  sourceRef:
    kind: GitRepository
    name: external
  validation: client
`

func populateAppRepo() (string, error) {
	dir, err := ioutil.TempDir("", "an-app-dir")
	if err != nil {
		return "", err
	}

	workloadPath1 := filepath.Join(dir, "kustomize", "one", "path", "to", "files")
	workloadPath2 := filepath.Join(dir, "kustomize", "another", "path", "to", "more", "files")
	if err := os.MkdirAll(workloadPath1, 0777); err != nil {
		return "", err
	}
	if err := os.MkdirAll(workloadPath2, 0777); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(workloadPath1, "nginx.yaml"), []byte("file1"), 0644); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(workloadPath2, "nginx.yaml"), []byte("file2"), 0644); err != nil {
		return "", err
	}

	return dir, nil
}

func sliceRemove(item string, items []string) []string {
	location := 0

	for idx, val := range items {
		if item == val {
			location = idx
			break
		}
	}

	return append(items[:location], items[location+1:]...)
}

var createdResources map[string][]string

var fluxDir string

var _ = Describe("Remove", func() {
	Context("Finding app manifests in repo dir", func() {
		var _ = BeforeEach(func() {
			application = makeWegoApplication(AddParams{
				Url:            "https://github.com/foo/bar",
				Path:           "./kustomize",
				Branch:         "main",
				Dir:            ".",
				DeploymentType: "kustomize",
				Namespace:      "wego-system",
				AppConfigUrl:   "NONE",
				AutoMerge:      true,
			})
		})

		It("gives a correct error message when app path not found", func() {
			application.Spec.Path = "./badpath"
			appRepoDir, err := populateAppRepo()
			Expect(err).ShouldNot(HaveOccurred())
			defer os.RemoveAll(appRepoDir)
			_, err = findAppManifests(application, appRepoDir)
			Expect(err).Should(MatchError("application path './badpath' not found"))
		})

		It("locates application manifests", func() {
			appRepoDir, err := populateAppRepo()
			Expect(err).ShouldNot(HaveOccurred())
			defer os.RemoveAll(appRepoDir)
			manifests, err := findAppManifests(application, appRepoDir)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(manifests)).To(Equal(2))
			for _, manifest := range manifests {
				Expect(manifest).To(Or(Equal([]byte("file1")), Equal([]byte("file2"))))
			}
		})
	})

	Context("Collecting resources deployed to cluster", func() {
		var _ = BeforeEach(func() {
			addParams = AddParams{
				Url:            "https://charts.kube-ops.io",
				Branch:         "main",
				Chart:          "loki",
				DeploymentType: "helm",
				Namespace:      "wego-system",
				AppConfigUrl:   "NONE",
				AutoMerge:      true,
			}
			dir, err := ioutil.TempDir("", "a-home-dir")
			Expect(err).ShouldNot(HaveOccurred())

			fluxDir = dir
			cliRunner := &runner.CLIRunner{}
			osysClient := &osysfakes.FakeOsys{}
			appSrv.(*App).flux = flux.New(osysClient, cliRunner)
			osysClient.UserHomeDirStub = func() (string, error) {
				return dir, nil
			}
			appSrv.(*App).flux.SetupFluxBin()
			createdResources = map[string][]string{}

			kubeClient.ApplyStub = func(manifest []byte, namespace string) ([]byte, error) {
				manifestMap := map[string]interface{}{}

				if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
					return nil, err
				}

				metamap := manifestMap["metadata"].(map[string]interface{})
				kind := manifestMap["kind"].(string)

				if createdResources[kind] == nil {
					createdResources[kind] = []string{}
				}

				createdResources[kind] = append(createdResources[kind], metamap["name"].(string))
				return []byte(""), nil
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		It("collects cluster resources for helm with configURL = NONE", func() {
			fluxClient.CreateSourceHelmStub = func(string, string, string) ([]byte, error) {
				return []byte(helmSource), nil
			}

			fluxClient.CreateHelmReleaseHelmRepositoryStub = func(string, string, string) ([]byte, error) {
				return []byte(helmGoatForHelmSource), nil
			}

			addParams, err := appSrv.(*App).updateParametersIfNecessary(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			err = appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			info := getAppResourceInfo(makeWegoApplication(addParams), "test-cluster")
			appResources := info.clusterResources()

			for _, res := range appResources {
				resources := createdResources[res.kind]
				Expect(resources).To(Not(BeEmpty()))
				createdResources[res.kind] = sliceRemove(res.name, resources)
			}

			for _, leftovers := range createdResources {
				Expect(leftovers).To(BeEmpty())
			}
		})

		It("collects cluster resources for helm with configURL = <url>", func() {
			addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
			addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

			fluxClient.CreateSourceGitStub = func(repoName string, _ string, _ string, _ string, _ string) ([]byte, error) {
				if repoName == "wego-fork-test" {
					return []byte(gitSourceForHelmAppRepo), nil
				} else {
					return []byte(gitSourceForHelmConfigRepo), nil
				}
			}

			fluxClient.CreateKustomizationStub = func(kustName string, _ string, _ string, _ string) ([]byte, error) {
				if kustName == "my-helm-app-apps-dir" {
					return []byte(kustForHelmAppsDir), nil
				} else {
					return []byte(kustForHelmTargetDir), nil
				}
			}

			fluxClient.CreateHelmReleaseGitRepositoryStub = func(string, string, string, string) ([]byte, error) {
				return []byte(helmGoatForGitSource), nil
			}

			addParams, err := appSrv.(*App).updateParametersIfNecessary(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			err = appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			info := getAppResourceInfo(makeWegoApplication(addParams), "test-cluster")
			appResources := info.clusterResources()

			for _, res := range appResources {
				resources := createdResources[res.kind]
				Expect(resources).To(Not(BeEmpty()))
				createdResources[res.kind] = sliceRemove(res.name, resources)
			}

			for _, leftovers := range createdResources {
				Expect(leftovers).To(BeEmpty())
			}
		})
	})
})
