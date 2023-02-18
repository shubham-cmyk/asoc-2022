package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ClusterRoleName is the name of the ClusterRole for Serverless Job
	ClusterRoleName = "serverless-clusterrole"
	// ServiceAccountName is the name of the ServiceAccount for Serverless Job
	ServiceAccountName = "serverless-service-account"
)

// Serverless Cluster Role
func createServerlessClusterRole(ctx context.Context, k8sClient client.Client, clusterRoleName string) error {
	var clusterRole = rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "create", "update", "delete"},
				APIGroups: []string{""},
				Resources: []string{"secrets", "jobs"},
			},
			{
				Verbs:     []string{"get", "list", "create", "update", "delete"},
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
			},
			{
				Verbs:     []string{"get", "list", "create", "update", "delete"},
				APIGroups: []string{"serverless.devs.serverless.domain"},
				Resources: []string{"serverless"},
			},
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: clusterRoleName}, &rbacv1.ClusterRole{}); err != nil {
		if kerrors.IsNotFound(err) {
			if err := k8sClient.Create(ctx, &clusterRole); err != nil {
				return errors.Wrap(err, "failed to create ClusterRole")
			}
		}
	}
	return nil
}

// Serverless Cluster Role Binding
func createServerlessClusterRoleBinding(ctx context.Context, k8sClient client.Client, namespace, clusterRoleName string, serviceAccountName string) error {
	var crbName = fmt.Sprintf("%s-clusterrole-binding", namespace)
	var clusterRoleBinding = rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crbName,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: crbName}, &rbacv1.ClusterRoleBinding{}); err != nil {
		if kerrors.IsNotFound(err) {
			if err := k8sClient.Create(ctx, &clusterRoleBinding); err != nil {
				return errors.Wrap(err, "failed to create ClusterRoleBinding")
			}
		}
	}
	return nil

}

// Serverless Service Account
func createServerlessServiceAccount(ctx context.Context, k8sClient client.Client, namespace string, serviceAccountName string) error {

	var serverlessServiceAcccount = v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: serviceAccountName, Namespace: namespace}, &serverlessServiceAcccount); err != nil {
		if kerrors.IsNotFound(err) {
			if err := k8sClient.Create(ctx, &serverlessServiceAcccount); err != nil {
				return errors.Wrap(err, "failed to create ServiceAccount")
			}
		}
	}

	return nil

}
