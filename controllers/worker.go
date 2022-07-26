package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

// return pods, err, and success
func (r *MPIJobReconciler) getOrCreateWorker(ctx context.Context, mpiJob *v1.MPIJob) ([]*corev1.Pod, error) {
	logger := log.FromContext(ctx)
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
			logger.Info("WARN:Some worker pod is not controlled by this MPIJob resource. Skipping",
				"Pod Name", worker.Name)
			// we don't control this worker pod
			continue
		}
		workerPods = append(workerPods, &worker)
	}
	return workerPods, nil
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func isPodReady(status corev1.PodStatus) bool {
	if status.Conditions == nil {
		return false
	}
	for _, c := range status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func workerReady(workers []*corev1.Pod) bool {
	for _, w := range workers {
		if w.Status.Phase == corev1.PodRunning && isPodReady(w.Status) {
			continue
		}
		return false
	}
	return true
}
