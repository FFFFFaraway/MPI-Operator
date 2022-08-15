package controllers

import (
	"context"
	v1 "github.com/FFFFFaraway/MPI-Operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func newWorker(mpiJob *v1.MPIJob) *appsv1.StatefulSet {
	template := *mpiJob.Spec.WorkerTemplate.DeepCopy()
	if template.Labels == nil {
		template.Labels = map[string]string{}
	}
	template.Labels["app"] = mpiJob.Name + workerSuffix
	template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpiJob.Name + workerSuffix,
			Namespace: mpiJob.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: mpiJob.Spec.NumWorkers,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": mpiJob.Name + workerSuffix,
				},
			},
			Template:            template,
			ServiceName:         mpiJob.Name,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

// return pods, err, and success
func (r *MPIJobReconciler) getOrCreateWorker(ctx context.Context, mpiJob *v1.MPIJob) (*appsv1.StatefulSet, error) {
	logger := log.FromContext(ctx)
	if mpiJob.Spec.WorkerTemplate.Spec.RestartPolicy != corev1.RestartPolicyAlways {
		logger.Info("WARN:Overwrite RestartPolicy in WorkerTemplate to Always.")
	}
	newWorker := newWorker(mpiJob)
	if err := ctrl.SetControllerReference(mpiJob, newWorker, r.Scheme); err != nil {
		return nil, err
	}
	var worker appsv1.StatefulSet
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + workerSuffix}, &worker)
	// If the worker StatefulSet doesn't exist, we'll create it.
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, newWorker); err != nil {
			return nil, err
		}
		return newWorker, nil
	}
	if err != nil {
		return nil, err
	}
	// If the worker is not controlled by this MPIJob resource, we should log
	// a warning to the event recorder and return.
	if !metav1.IsControlledBy(&worker, mpiJob) {
		logger.Info("WARN:worker statefulset is not controlled by this MPIJob resource. Skipping",
			"Pod Name", worker.Name)
		// we don't control this worker pod
		return nil, nil
	}
	if err := r.Update(ctx, newWorker); err != nil {
		return nil, err
	}
	return newWorker, nil
}
