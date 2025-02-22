package rollouts

import (
	"context"
	"fmt"
	"reflect"

	rolloutsApi "github.com/iam-veeramalla/argo-rollouts-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconciles rollouts serviceaccount.
func (r *ArgoRolloutsReconciler) reconcileRolloutsServiceAccount(cr *rolloutsApi.ArgoRollout) (*corev1.ServiceAccount, error) {

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: cr.Namespace,
		},
	}
	setRolloutsLabels(&sa.ObjectMeta)

	if err := fetchObject(r.Client, cr.Namespace, sa.Name, sa); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get the serviceAccount associated with %s : %s", sa.Name, err)
		}

		if err := controllerutil.SetControllerReference(cr, sa, r.Scheme); err != nil {
			return nil, err
		}

		log.Info(fmt.Sprintf("Creating serviceaccount %s", sa.Name))
		err := r.Client.Create(context.TODO(), sa)
		if err != nil {
			return nil, err
		}

	}
	return sa, nil
}

// Reconciles rollouts role.
func (r *ArgoRolloutsReconciler) reconcileRolloutsRole(cr *rolloutsApi.ArgoRollout) (*v1.Role, error) {

	expectedPolicyRules := getPolicyRules()

	role := &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: cr.Namespace,
		},
	}
	setRolloutsLabels(&role.ObjectMeta)

	if err := fetchObject(r.Client, cr.Namespace, role.Name, role); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to reconcile the role for the service account associated with %s : %s", role.Name, err)
		}

		if err = controllerutil.SetControllerReference(cr, role, r.Scheme); err != nil {
			return nil, err
		}

		log.Info(fmt.Sprintf("Creating role %s", role.Name))
		role.Rules = expectedPolicyRules
		return role, r.Client.Create(context.TODO(), role)
	}

	// Reconcile if the role already exists and modified.
	if !reflect.DeepEqual(role.Rules, expectedPolicyRules) {
		role.Rules = expectedPolicyRules
		return role, r.Client.Update(context.TODO(), role)
	}

	return role, nil
}

// Reconcile rollouts rolebinding.
func (r *ArgoRolloutsReconciler) reconcileRolloutsRoleBinding(cr *rolloutsApi.ArgoRollout, role *v1.Role, sa *corev1.ServiceAccount) error {

	expectedRoleBinding := &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: cr.Namespace,
		},
	}
	setRolloutsLabels(&expectedRoleBinding.ObjectMeta)

	expectedRoleBinding.RoleRef = v1.RoleRef{
		APIGroup: v1.GroupName,
		Kind:     "Role",
		Name:     role.Name,
	}

	expectedRoleBinding.Subjects = []v1.Subject{
		{
			Kind:      v1.ServiceAccountKind,
			Name:      sa.Name,
			Namespace: sa.Namespace,
		},
	}

	actualRoleBinding := &v1.RoleBinding{}

	// Fetch the rolebinding if exists and store that in actualRoleBinding.
	if err := fetchObject(r.Client, cr.Namespace, expectedRoleBinding.Name, actualRoleBinding); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get the rolebinding associated with %s : %s", expectedRoleBinding.Name, err)
		}

		if err := controllerutil.SetControllerReference(cr, expectedRoleBinding, r.Scheme); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Creating rolebinding %s", expectedRoleBinding.Name))
		return r.Client.Create(context.TODO(), expectedRoleBinding)
	}

	// Reconcile if the role already exists and modified.
	if !reflect.DeepEqual(expectedRoleBinding.Subjects, actualRoleBinding.Subjects) {
		actualRoleBinding.Subjects = expectedRoleBinding.Subjects
		r.Client.Update(context.TODO(), actualRoleBinding)
	}

	return nil
}

// Reconcile rollouts service.
func (r *ArgoRolloutsReconciler) reconcileRolloutsService(cr *rolloutsApi.ArgoRollout) error {

	expectedSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: cr.Namespace,
		},
	}
	setRolloutsLabels(&expectedSvc.ObjectMeta)

	expectedSvc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "metrics",
			Port:       8090,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(8090),
		},
	}

	expectedSvc.Spec.Selector = map[string]string{
		DefaultRolloutsSelectorKey: DefaultArgoRolloutsResourceName,
	}

	actualSvc := &corev1.Service{}

	// Fetch the service if exists and store that in actualSvc.
	if err := fetchObject(r.Client, cr.Namespace, expectedSvc.Name, actualSvc); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get the service %s : %s", expectedSvc.Name, err)
		}

		if err := controllerutil.SetControllerReference(cr, expectedSvc, r.Scheme); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Creating service %s", expectedSvc.Name))
		return r.Client.Create(context.TODO(), expectedSvc)
	}

	if !reflect.DeepEqual(actualSvc.Spec.Ports, expectedSvc.Spec.Ports) {
		actualSvc.Spec.Ports = expectedSvc.Spec.Ports
		return r.Client.Create(context.TODO(), actualSvc)
	}

	return nil
}

// Reconciles secrets for rollouts controller
func (r *ArgoRolloutsReconciler) reconcileRolloutsSecrets(cr *rolloutsApi.ArgoRollout) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultRolloutsNotificationSecretName,
			Namespace: cr.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	if err := fetchObject(r.Client, cr.Namespace, secret.Name, secret); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get the secret %s : %s", secret.Name, err)
		}

		if err := controllerutil.SetControllerReference(cr, secret, r.Scheme); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Creating secret %s", secret.Name))
		return r.Client.Create(context.TODO(), secret)
	}

	// secret found, do nothing
	return nil
}

// Deletes rollout resources when the corresponding rollout CR is deleted.
func (r *ArgoRolloutsReconciler) deleteRolloutResources(cr *rolloutsApi.ArgoRollout) error {
	if cr.DeletionTimestamp != nil {
		log.Info(fmt.Sprintf("Argo Rollout resource in %s namespace is deleted, Deleting the Argo Rollout workloads",
			cr.Namespace))

		serviceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsResourceName,
				Namespace: cr.Namespace,
			},
		}

		if err := r.Client.Delete(context.TODO(), serviceAccount); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting the service account %s in %s",
				DefaultArgoRolloutsResourceName, cr.Namespace))
		}

		role := &v1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsResourceName,
				Namespace: cr.Namespace,
			},
		}
		if err := r.Client.Delete(context.TODO(), role); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting role %s in %s",
				DefaultArgoRolloutsResourceName, cr.Namespace))
		}

		rolebinding := &v1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsResourceName,
				Namespace: cr.Namespace,
			},
		}
		if err := r.Client.Delete(context.TODO(), rolebinding); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting the rolebinding %s in %s",
				DefaultArgoRolloutsResourceName, cr.Namespace))
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultRolloutsNotificationSecretName,
				Namespace: cr.Namespace,
			},
		}
		if err := r.Client.Delete(context.TODO(), secret); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting the secret %s in %s",
				DefaultRolloutsNotificationSecretName, cr.Namespace))
		}

		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsResourceName,
				Namespace: cr.Namespace,
			},
		}
		if err := r.Client.Delete(context.TODO(), deploy); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting the deployment %s in %s",
				DefaultArgoRolloutsResourceName, cr.Namespace))
		}

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsResourceName,
				Namespace: cr.Namespace,
			},
		}
		if err := r.Client.Delete(context.TODO(), svc); err != nil {
			log.Error(err, fmt.Sprintf("Error deleting the service %s in %s",
				DefaultArgoRolloutsResourceName, cr.Namespace))
		}
	}

	return nil
}
