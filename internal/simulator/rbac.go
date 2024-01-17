package simulator

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	serviceAccountName     = "batch-simulator"
	clusterRoleName        = "batch-simulator-role"
	clusterRoleBindingName = "batch-simulator-binding"
)

func CreateRBAC(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := createServiceAccount(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	err = createClusterRole(ctx, clientset)
	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}

	err = createClusterRoleBinding(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error creating role binding: %w", err)
	}

	return nil
}

func createServiceAccount(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
		},
	}

	_, err := clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createClusterRole(ctx context.Context, clientset kubernetes.Interface) error {
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
		},
	}

	_, err := clientset.RbacV1().ClusterRoles().Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createClusterRoleBinding(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func DeleteRBAC(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := deleteClusterRole(ctx, clientset)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	err = deleteClusterRoleBinding(ctx, clientset)
	if err != nil {
		return fmt.Errorf("error deleting role binding: %w", err)
	}

	err = deleteServiceAccount(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}

func deleteServiceAccount(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := clientset.CoreV1().ServiceAccounts(namespace).Delete(ctx, serviceAccountName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func deleteClusterRole(ctx context.Context, clientset kubernetes.Interface) error {
	err := clientset.RbacV1().ClusterRoles().Delete(ctx, clusterRoleName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func deleteClusterRoleBinding(ctx context.Context, clientset kubernetes.Interface) error {
	err := clientset.RbacV1().ClusterRoleBindings().Delete(ctx, clusterRoleBindingName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
