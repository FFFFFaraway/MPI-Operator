package controllers

import (
	"bytes"
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

func newConfigMap(mpiJob *v1.MPIJob) *corev1.ConfigMap {
	kubexec := fmt.Sprintf(`#!/bin/sh
set -x
POD_NAME=$1
shift
%s/kubectl exec ${POD_NAME} -- /bin/sh -c "$*"`, kubectlMountPath)

	slots := 1
	var buffer bytes.Buffer
	for i := 0; i < *mpiJob.Spec.NumWorkers; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s-%d slots=%d\n", mpiJob.Name, workerSuffix, i, slots))
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpiJob.Name + configSuffix,
			Namespace: mpiJob.Namespace,
			Labels: map[string]string{
				"app": mpiJob.Name,
			},
		},
		Data: map[string]string{
			hostfileName:      buffer.String(),
			kubexecScriptName: kubexec,
		},
	}
}

func (r *MPIJobReconciler) getOrCreateConfigMap(ctx context.Context, mpiJob *v1.MPIJob) error {
	logger := log.FromContext(ctx)
	newCM := newConfigMap(mpiJob)
	if err := ctrl.SetControllerReference(mpiJob, newCM, r.Scheme); err != nil {
		return err
	}
	var cm corev1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + configSuffix}, &cm)
	if errors.IsNotFound(err) {
		logger.V(1).Info("ConfigMap doesn't exist, creating...")
		// If the ConfigMap doesn't exist, we'll create it.
		if err := r.Create(ctx, newCM); err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}

	// If the ConfigMap is not controlled by this MPIJob resource, we
	// should log a warning to the event recorder and return.
	if !metav1.IsControlledBy(&cm, mpiJob) {
		logger.Info("WARN:ConfigMap is not controlled by this MPIJob resource. Skipping")
		return nil
	}

	if err := r.Update(ctx, newCM); err != nil {
		return err
	}
	return nil
}
