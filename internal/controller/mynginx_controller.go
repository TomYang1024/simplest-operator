/*
Copyright 2024 tomyang.

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

package controller

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	exampleiov1alpha1 "github.com/yanglunara/simplest-operator/api/v1alpha1"
	"github.com/yanglunara/simplest-operator/internal/resource"
)

// MyNginxReconciler reconciles a MyNginx object
type MyNginxReconciler struct {
	ResourceInter resource.ResourceInter
}

//+kubebuilder:rbac:groups=app,resources=deployment,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=service,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.io.example.io,resources=mynginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.io.example.io,resources=mynginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.io.example.io,resources=mynginxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyNginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MyNginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	logger := log.FromContext(ctx)
	logger = logger.WithValues("appService", req.NamespacedName)
	var (
		appService exampleiov1alpha1.MyNginx
	)
	if err = r.ResourceInter.GetResource(ctx, req.NamespacedName, &appService); err != nil {
		return res, client.IgnoreNotFound(err)
	}
	logger.Info("fetch appService", "appService", appService)
	if err = r.ResourceInter.CreateOrUpdateFormDeploy(ctx, &appService); err != nil {
		return
	}
	logger.Info("create or update deploy", "Deployment", appService)
	if err = r.ResourceInter.CreateOrUpdateFormService(ctx, &appService); err != nil {
		return
	}
	logger.Info("create or update service", "Service", appService)
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyNginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&exampleiov1alpha1.MyNginx{}).
		Complete(r)
}
