/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/kustomize"
	"github.com/weaveworks/weave-gitops/core/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appNameFlag    = "app-name"
	namespaceFlag  = "namespace"
	pathFlag       = "path"
	sourceKindFlag = "source-kind"
	sourceNameFlag = "source-name"
)

type Params struct {
	Name       string
	AppName    string
	Namespace  string
	SourceName string
	SourceKind string
	Path       string
}

var (
	params Params
)

// appCmd represents the app command
var KustomizationCmd = &cobra.Command{
	Use:   "kustomization",
	Short: "Adds a kustomization to an app in the GitOps repository",
	Long: `This command will add a kustomization to the specified application.  T
This adds just basic fields and can easily be edited manually to fit more advanced
scenarios.`,
	RunE: runCmd,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("add app requires a name argument")
		}
		params.Name = args[0]
		return nil
	},
}

func init() {
	KustomizationCmd.Flags().StringVar(&params.Namespace, appNameFlag, "", "The app name the kustomization should be added to.")
	KustomizationCmd.Flags().StringVar(&params.Namespace, namespaceFlag, "", "Namespace for the app")
	KustomizationCmd.Flags().StringVar(&params.Path, pathFlag, "", "The source's file path where the files to be kustomized reside.")
	KustomizationCmd.Flags().StringVar(&params.Namespace, sourceKindFlag, "", "The kind of the source (GitRepository | Bucket | HelmRepository)")
	KustomizationCmd.Flags().StringVar(&params.Namespace, sourceNameFlag, "", "The name of the source (Git Repository, Bucket, etc) that has the files")
}

func runCmd(cmd *cobra.Command, args []string) error {
	r := bufio.NewReader(os.Stdin)
	return createKustomization(r)
}

func createKustomization(r *bufio.Reader) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine local directory: %w", err)
	}

	repo, err := internal.GitRepository(dir)
	if err != nil {
		return err
	}

	gitCommitter := repository.NewGitWriter(false)
	appRepoFetcher := app.NewRepoFetcher()
	kustSvc := kustomize.NewCreator(gitCommitter, appRepoFetcher)

	if params.Namespace == "" {
		fmt.Printf("Namespace (e.g. flux-system): ")

		params.Namespace, err = readAndFormatInput(r, "namespace")
		if err != nil {
			return err
		}
	}

	if params.SourceKind == "" {
		fmt.Printf("Source kind (GitRepository | Bucket | HelmRepository): ")

		params.SourceKind, err = readAndFormatInput(r, "source-kind")
		if err != nil {
			return err
		}
	}

	if params.SourceName == "" {
		fmt.Printf("Source name: ")

		params.SourceName, err = readAndFormatInput(r, "source-name")
		if err != nil {
			return err
		}
	}

	if params.Path == "" {
		fmt.Printf("File path in the source: ")

		params.Path, err = readAndFormatInput(r, "path")
		if err != nil {
			return err
		}
	}

	if params.AppName == "" {
		fmt.Printf("App name: ")

		params.AppName, err = readAndFormatInput(r, "app-name")
		if err != nil {
			return err
		}
	}

	_, err = kustSvc.Create(dir, repo, internal.StubAuth{}, kustomize.CreateInput{
		AppName: params.AppName,
		Kustomization: v1beta2.Kustomization{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      params.Name,
				Namespace: params.Namespace,
			},
			Spec: v1beta2.KustomizationSpec{
				Path: params.Path,
				SourceRef: v1beta2.CrossNamespaceSourceReference{
					Kind: params.SourceKind,
					Name: params.SourceName,
				},
			},
			Status: v1beta2.KustomizationStatus{},
		},
	})
	if err != nil {
		return fmt.Errorf("issue creating a kustomization in app %s: %w", params.AppName, err)
	}

	return nil
}

func readAndFormatInput(r *bufio.Reader, field string) (string, error) {
	input, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("issue reading input for %s: %w", field, err)
	}

	input = strings.Replace(input, "\n", "", -1)
	return input, nil
}
