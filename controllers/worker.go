package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "test.bdap.com/project/api/v1"
)

func newWorker(mpiJob *v1.MPIJob, name string) *corev1.Pod {
	podSpec := mpiJob.Spec.WorkerTemplate.DeepCopy()
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   mpiJob.Namespace,
			Labels:      podSpec.Labels,
			Annotations: podSpec.Annotations,
		},
		Spec: podSpec.Spec,
	}
}

func (r *MPIJobReconciler) getOrCreateWorker(ctx context.Context, mpiJob *v1.MPIJob) ([]*corev1.Pod, error) {
	var workerPods []*corev1.Pod
	for i := 0; i < *mpiJob.Spec.NumWorkers; i++ {
		name := fmt.Sprintf("%s-%d", mpiJob.Name+workerSuffix, i)
		var worker corev1.Pod
		err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: name}, &worker)
		// If the worker Pod doesn't exist, we'll create it.
		if errors.IsNotFound(err) {
			newWorker := newWorker(mpiJob, name)
			if err := ctrl.SetControllerReference(mpiJob, newWorker, r.Scheme); err != nil {
				return nil, err
			}
			if err := r.Create(ctx, newWorker); err != nil {
				return nil, err
			}
			workerPods = append(workerPods, newWorker)
			continue
		}
		if err != nil {
			return nil, err
		}
		// If the worker is not controlled by this MPIJob resource, we should log
		// a warning to the event recorder and return.
		if !metav1.IsControlledBy(&worker, mpiJob) {
			err := fmt.Errorf("worker pod is not controlled by this MPIJob resource")
			return nil, err
		}
		// [sw]: Because if the API server does not allow update for some reason, it will always try to update
		//if err := r.Update(ctx, newWorker); err != nil {
		//	return nil, err
		//}
		workerPods = append(workerPods, &worker)
	}
	return workerPods, nil
}
