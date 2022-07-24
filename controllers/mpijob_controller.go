/*
Copyright 2022.

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

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "test.bdap.com/project/api/v1"
)

// MPIJobReconciler reconciles a MPIJob object
type MPIJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	configSuffix            = "-config"
	configVolumeName        = "mpi-job-config"
	configMountPath         = "/etc/mpi"
	kubexecScriptName       = "kubexec.sh"
	hostfileName            = "hostfile"
	kubectlDeliveryName     = "kubectl-delivery"
	kubectlTargetDirEnv     = "TARGET_DIR"
	kubectlVolumeName       = "mpi-job-kubectl"
	kubectlMountPath        = "/opt/kube"
	launcherSuffix          = "-launcher"
	workerSuffix            = "-worker"
	initContainerCpu        = "100m"
	initContainerEphStorage = "5Gi"
	initContainerMem        = "512Mi"
)

//+kubebuilder:rbac:groups=batch.test.bdap.com,resources=mpijobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.test.bdap.com,resources=mpijobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.test.bdap.com,resources=mpijobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MPIJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *MPIJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var mpiJob batchv1.MPIJob
	if err := r.Get(ctx, req.NamespacedName, &mpiJob); err != nil {
		logger.Error(err, "unable to fetch MPIJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Discover one MPIJob")

	if err := r.getOrCreateConfigMap(ctx, &mpiJob); err != nil {
		logger.Error(err, "can't getOrCreateConfigMap")
		return ctrl.Result{}, err
	}
	if err := r.getOrCreateLauncherServiceAccount(ctx, &mpiJob); err != nil {
		logger.Error(err, "can't getOrCreateLauncherServiceAccount")
		return ctrl.Result{}, err
	}
	if err := r.getOrCreateLauncherRole(ctx, &mpiJob); err != nil {
		logger.Error(err, "can't getOrCreateLauncherRole")
		return ctrl.Result{}, err
	}
	if err := r.getOrCreateLauncherRoleBinding(ctx, &mpiJob); err != nil {
		logger.Error(err, "can't getOrCreateLauncherRoleBinding")
		return ctrl.Result{}, err
	}
	worker, err := r.getOrCreateWorker(ctx, &mpiJob)
	if err != nil {
		logger.Error(err, "can't getOrCreateWorker")
		return ctrl.Result{}, err
	}
	launcher, err := r.getOrCreateLauncher(ctx, &mpiJob)
	if err != nil {
		logger.Error(err, "can't getOrCreateLauncher")
		return ctrl.Result{}, err
	}
	logger.Info("pod info", "worker", worker, "launcher", launcher)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MPIJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.MPIJob{}).
		Complete(r)
}
