package rollouts

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileArgoRollouts_verifyRolloutsResources(t *testing.T) {
	a := makeTestArgoRollout()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
	}

	res, err := r.Reconcile(context.TODO(), req)
	assert.NoError(t, err)
	if res.Requeue {
		t.Fatal("reconcile requeued request")
	}

	sa := &corev1.ServiceAccount{}
	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, sa); err != nil {
		t.Fatalf("failed to find the rollouts serviceaccount: %#v\n", err)
	}

	role := &v1.Role{}
	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, role); err != nil {
		t.Fatalf("failed to find the rollouts role: %#v\n", err)
	}

	rolebinding := &v1.RoleBinding{}
	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, rolebinding); err != nil {
		t.Fatalf("failed to find the rollouts rolebinding: %#v\n", err)
	}

	service := &corev1.Service{}
	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, service); err != nil {
		t.Fatalf("failed to find the rollouts service: %#v\n", err)
	}

	secret := &corev1.Secret{}
	if err = r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      DefaultRolloutsNotificationSecretName,
		Namespace: testNamespace,
	}, secret); err != nil {
		t.Fatalf("failed to find the rollouts secret: %#v\n", err)
	}
}

func TestReconcileArgoRollouts_CleanUp(t *testing.T) {

	a := makeTestArgoRollout()

	resources := []runtime.Object{a}

	r := makeTestReconciler(t, resources...)
	assert.NoError(t, createNamespace(r, a.Namespace))

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
	}
	res, err := r.Reconcile(context.TODO(), req)
	assert.NoError(t, err)
	if res.Requeue {
		t.Fatal("reconcile requeued request")
	}

	r.Client.Delete(context.TODO(), a)

	// check if rollouts resources are deleted
	tt := []struct {
		name     string
		resource client.Object
	}{
		{
			fmt.Sprintf("ServiceAccount %s", DefaultArgoRolloutsResourceName),
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			},
		},
		{
			fmt.Sprintf("Role %s", DefaultArgoRolloutsResourceName),
			&rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			},
		},
		{
			fmt.Sprintf("RoleBinding %s", DefaultArgoRolloutsResourceName),
			&v1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			},
		},
		{
			fmt.Sprintf("Secret %s", DefaultRolloutsNotificationSecretName),
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultRolloutsNotificationSecretName,
					Namespace: a.Namespace,
				},
				Type: corev1.SecretTypeOpaque,
			},
		},
		{
			fmt.Sprintf("Service %s", DefaultArgoRolloutsResourceName),
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			if err = fetchObject(r.Client, a.Namespace, test.name, test.resource); err == nil {
				t.Errorf("Expected %s to be deleted", test.name)
			}
		})
	}
}
