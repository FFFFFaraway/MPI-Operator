package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "test.bdap.com/project/api/v1"
)

func newLauncher(mpiJob *v1.MPIJob, kubectlDeliveryImage string) (*corev1.Pod, error) {
	podSpec := mpiJob.Spec.LauncherTemplate.DeepCopy()
	podSpec.Spec.ServiceAccountName = mpiJob.Name + launcherSuffix
	podSpec.Spec.InitContainers = append(podSpec.Spec.InitContainers, corev1.Container{
		Name:            kubectlDeliveryName,
		Image:           kubectlDeliveryImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			{
				Name:  kubectlTargetDirEnv,
				Value: kubectlMountPath,
			},
			{
				Name:  "NAMESPACE",
				Value: mpiJob.Namespace,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      kubectlVolumeName,
				MountPath: kubectlMountPath,
			},
			{
				Name:      configVolumeName,
				MountPath: configMountPath,
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse(initContainerCpu),
				corev1.ResourceMemory:           resource.MustParse(initContainerMem),
				corev1.ResourceEphemeralStorage: resource.MustParse(initContainerEphStorage),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse(initContainerCpu),
				corev1.ResourceMemory:           resource.MustParse(initContainerMem),
				corev1.ResourceEphemeralStorage: resource.MustParse(initContainerEphStorage),
			},
		},
	})

	if len(podSpec.Spec.Containers) == 0 {
		err := fmt.Errorf("launcher pod does not have any containers in its spec")
		return nil, err
	}
	container := podSpec.Spec.Containers[0]
	container.Env = append(container.Env,
		corev1.EnvVar{
			Name:  "OMPI_MCA_plm_rsh_agent",
			Value: fmt.Sprintf("%s/%s", configMountPath, kubexecScriptName),
		},
		corev1.EnvVar{
			Name:  "OMPI_MCA_orte_default_hostfile",
			Value: fmt.Sprintf("%s/%s", configMountPath, hostfileName),
		},
	)

	container.VolumeMounts = append(container.VolumeMounts,
		corev1.VolumeMount{
			Name:      kubectlVolumeName,
			MountPath: kubectlMountPath,
		},
		corev1.VolumeMount{
			Name:      configVolumeName,
			MountPath: configMountPath,
		})
	podSpec.Spec.Containers[0] = container

	scriptsMode := int32(0555)
	hostfileMode := int32(0444)
	podSpec.Spec.Volumes = append(podSpec.Spec.Volumes,
		corev1.Volume{
			Name: kubectlVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: configVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					// [sw] find the configmap
					LocalObjectReference: corev1.LocalObjectReference{
						Name: mpiJob.Name + configSuffix,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  kubexecScriptName,
							Path: kubexecScriptName,
							Mode: &scriptsMode,
						},
						{
							Key:  hostfileName,
							Path: hostfileName,
							Mode: &hostfileMode,
						},
					},
				},
			},
		})
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mpiJob.Name + launcherSuffix,
			Namespace:   mpiJob.Namespace,
			Labels:      podSpec.Labels,
			Annotations: podSpec.Annotations,
		},
		Spec: podSpec.Spec,
	}, nil
}

func (r *MPIJobReconciler) getOrCreateLauncher(ctx context.Context, mpiJob *v1.MPIJob) (*corev1.Pod, error) {
	var launcher corev1.Pod
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + launcherSuffix}, &launcher)
	// If the worker Pod doesn't exist, we'll create it.
	if errors.IsNotFound(err) {
		newLauncher, err := newLauncher(mpiJob, "coreharbor.bdap.com/sw/kubectl-delivery")
		if err != nil {
			return nil, err
		}
		if err := ctrl.SetControllerReference(mpiJob, newLauncher, r.Scheme); err != nil {
			return nil, err
		}
		if err := r.Create(ctx, newLauncher); err != nil {
			return nil, err
		}
		return newLauncher, nil
	}
	if err != nil {
		return nil, err
	}
	// If the launcher is not controlled by this MPIJob resource, we should log
	// a warning to the event recorder and return.
	if !metav1.IsControlledBy(&launcher, mpiJob) {
		err := fmt.Errorf("launcher pod is not controlled by this MPIJob resource")
		return nil, err
	}

	return &launcher, nil
}
