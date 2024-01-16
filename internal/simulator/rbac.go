package simulator

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateRBAC(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := createServiceAccount(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	err = createRole(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}

	err = createRoleBinding(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error creating role binding: %w", err)
	}

	return nil
}

func createServiceAccount(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-simulator",
		},
	}

	_, err := clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createRole(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-simulator-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
		},
	}

	_, err := clientset.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createRoleBinding(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-simulator-binding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "batch-simulator",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "batch-simulator-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := clientset.RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func DeleteRBAC(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := deleteRoleBinding(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error deleting role binding: %w", err)
	}

	err = deleteRole(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	err = deleteServiceAccount(ctx, clientset, namespace)
	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}

func deleteServiceAccount(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := clientset.CoreV1().ServiceAccounts(namespace).Delete(ctx, "batch-simulator", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func deleteRole(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := clientset.RbacV1().Roles(namespace).Delete(ctx, "batch-simulator-role", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func deleteRoleBinding(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	err := clientset.RbacV1().RoleBindings(namespace).Delete(ctx, "batch-simulator-binding", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
