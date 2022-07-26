package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	v1 "test.bdap.com/project/api/v1"
)

func getObjectMeta(mpiJob *v1.MPIJob, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      mpiJob.Name + suffix,
		Namespace: mpiJob.Namespace,
		Labels: map[string]string{
			"app": mpiJob.Name,
		},
	}
}

// newLauncherServiceAccount creates a new launcher ServiceAccount for an MPIJob
// resource. It also sets the appropriate OwnerReferences on the resource so
// handleObject can discover the MPIJob resource that 'owns' it.
func newLauncherServiceAccount(mpiJob *v1.MPIJob) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: getObjectMeta(mpiJob, launcherSuffix),
	}
}

// newLauncherRole creates a new launcher Role for an MPIJob resource. It also
// sets the appropriate OwnerReferences on the resource so handleObject can
// discover the MPIJob resource that 'owns' it.
func newLauncherRole(mpiJob *v1.MPIJob) *rbacv1.Role {
	var podNames []string
	for i := 0; i < int(*mpiJob.Spec.NumWorkers); i++ {
		podNames = append(podNames, fmt.Sprintf("%s%s-%d", mpiJob.Name, workerSuffix, i))
	}
	return &rbacv1.Role{
		ObjectMeta: getObjectMeta(mpiJob, launcherSuffix),
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{""},
				Resources: []string{"pods"},
			},
			{
				Verbs:         []string{"create"},
				APIGroups:     []string{""},
				Resources:     []string{"pods/exec"},
				ResourceNames: podNames,
			},
		},
	}
}

// newLauncherRoleBinding creates a new launcher RoleBinding for an MPIJob
// resource. It also sets the appropriate OwnerReferences on the resource so
// handleObject can discover the MPIJob resource that 'owns' it.
func newLauncherRoleBinding(mpiJob *v1.MPIJob) *rbacv1.RoleBinding {
	launcherName := mpiJob.Name + launcherSuffix
	return &rbacv1.RoleBinding{
		ObjectMeta: getObjectMeta(mpiJob, launcherSuffix),
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      launcherName,
				Namespace: mpiJob.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     launcherName,
		},
	}
}

func (r *MPIJobReconciler) getOrCreateLauncherServiceAccount(ctx context.Context, mpiJob *v1.MPIJob) error {
	logger := log.FromContext(ctx)
	newSA := newLauncherServiceAccount(mpiJob)
	if err := ctrl.SetControllerReference(mpiJob, newSA, r.Scheme); err != nil {
		return err
	}
	var sa corev1.ServiceAccount
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + launcherSuffix}, &sa)
	if errors.IsNotFound(err) {
		logger.V(1).Info("ServiceAccount doesn't exist, creating...")
		// If the ConfigMap doesn't exist, we'll create it.
		if err := r.Create(ctx, newSA); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	// If the ConfigMap is not controlled by this MPIJob resource, we
	// should log a warning to the event recorder and return.
	if !metav1.IsControlledBy(&sa, mpiJob) {
		logger.Info("WARN:ServiceAccount is not controlled by this MPIJob resource. Skipping")
		return nil
	}
	if err := r.Update(ctx, newSA); err != nil {
		return err
	}
	return nil
}

func (r *MPIJobReconciler) getOrCreateLauncherRole(ctx context.Context, mpiJob *v1.MPIJob) error {
	logger := log.FromContext(ctx)
	newRole := newLauncherRole(mpiJob)
	if err := ctrl.SetControllerReference(mpiJob, newRole, r.Scheme); err != nil {
		return err
	}
	var role rbacv1.Role
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + launcherSuffix}, &role)
	if errors.IsNotFound(err) {
		logger.V(1).Info("Role doesn't exist, creating...")
		// If the ConfigMap doesn't exist, we'll create it.
		if err := r.Create(ctx, newRole); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	// If the ConfigMap is not controlled by this MPIJob resource, we
	// should log a warning to the event recorder and return.
	if !metav1.IsControlledBy(&role, mpiJob) {
		logger.Info("WARN:Role is not controlled by this MPIJob resource. Skipping")
		return nil
	}
	if err := r.Update(ctx, newRole); err != nil {
		return err
	}
	return nil
}

func (r *MPIJobReconciler) getOrCreateLauncherRoleBinding(ctx context.Context, mpiJob *v1.MPIJob) error {
	logger := log.FromContext(ctx)
	newRb := newLauncherRoleBinding(mpiJob)
	if err := ctrl.SetControllerReference(mpiJob, newRb, r.Scheme); err != nil {
		return err
	}
	var rb rbacv1.RoleBinding
	err := r.Get(ctx, client.ObjectKey{Namespace: mpiJob.Namespace, Name: mpiJob.Name + launcherSuffix}, &rb)
	if errors.IsNotFound(err) {
		logger.V(1).Info("RoleBinding doesn't exist, creating...")
		// If the ConfigMap doesn't exist, we'll create it.
		if err := r.Create(ctx, newRb); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	// If the ConfigMap is not controlled by this MPIJob resource, we
	// should log a warning to the event recorder and return.
	if !metav1.IsControlledBy(&rb, mpiJob) {
		logger.Info("WARN:RoleBinding is not controlled by this MPIJob resource. Skipping")
		return nil
	}
	if err := r.Update(ctx, newRb); err != nil {
		return err
	}
	return nil
}
