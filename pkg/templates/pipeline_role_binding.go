package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineRoleBinding(pipeline api.Pipeline) rbacv1.RoleBinding {
	return rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: pipeline.GetResourcePrefix(),
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      pipeline.GetResourcePrefix(),
				Namespace: pipeline.GetNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     pipeline.GetResourcePrefix(),
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
