/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wego

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wegov1alpha1 "github.com/weaveworks/weave-gitops/controllers/wego-controller/apis/wego/v1alpha1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=wego.weave.works,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=wego.weave.works,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=wego.weave.works,resources=applications/finalizers,verbs=update
// Flux resources
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories/finalizers,verbs=get;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("application", req.NamespacedName)

	var app wegov1alpha1.Application
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		log.Error(err, "unable to fetch Application")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileGitRepository(ctx, app); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wegov1alpha1.Application{}).
		Complete(r)
}

func (r *ApplicationReconciler) reconcileGitRepository(ctx context.Context, app wegov1alpha1.Application) error {
	gitRepo := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{APIVersion: sourcev1.GroupVersion.String(), Kind: "GitRepository"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: app.Spec.URL,
			Reference: &sourcev1.GitRepositoryRef{
				Branch: app.Spec.Reference.Branch,
			},
		},
	}

	if err := ctrl.SetControllerReference(&app, &gitRepo, r.Scheme); err != nil {
		return err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("wego-controller")}
	if err := r.Patch(ctx, &gitRepo, client.Apply, applyOpts...); err != nil {
		return err
	}

	return nil
}
