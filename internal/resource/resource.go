package resource

import (
	"context"

	"github.com/yanglunara/simplest-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ResourceInter interface {
	CreateOrUpdateFormDeploy(ctx context.Context, as *v1alpha1.MyNginx) error
	CreateOrUpdateFormService(ctx context.Context, as *v1alpha1.MyNginx) error
	GetResource(ctx context.Context, namespace types.NamespacedName, as *v1alpha1.MyNginx) error
}

var (
	_ ResourceInter = &resources{}
)

type resources struct {
	client.Client
	Scheme *runtime.Scheme
}

func NewResources(client client.Client, sm *runtime.Scheme) ResourceInter {
	return &resources{
		Client: client,
		Scheme: sm,
	}
}

func (r *resources) newContainers(app *v1alpha1.MyNginx) []corev1.Container {
	containerPorts := make([]corev1.ContainerPort, 0, len(app.Spec.Ports))
	for _, port := range app.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.Port,
		})
	}
	return []corev1.Container{
		{
			Name:      app.Name,
			Image:     app.Spec.Image,
			Resources: app.Spec.Resources,
			Env:       app.Spec.Envs,
			Ports:     containerPorts,
		},
	}
}

func (r *resources) mutateService(app *v1alpha1.MyNginx, svc *corev1.Service) {
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: svc.Spec.ClusterIP,
		Ports:     app.Spec.Ports,
		Type:      corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			"mynginx": app.Name,
		},
	}
}

func (r *resources) mutateDeploy(as *v1alpha1.MyNginx, deploy *appsv1.Deployment) error {
	labels := map[string]string{
		"myapp": as.Name,
	}
	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: as.Spec.Size,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: r.newContainers(as),
			},
		},
		Selector: &selector,
	}
	return nil
}

func (r *resources) CreateOrUpdateFormDeploy(ctx context.Context, as *v1alpha1.MyNginx) (err error) {
	var (
		deploy appsv1.Deployment
	)
	deploy.Name = as.Name
	deploy.Namespace = as.Namespace
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		r.mutateDeploy(as, &deploy)
		return controllerutil.SetControllerReference(as, &deploy, r.Scheme)
	})
	return err
}

func (r *resources) CreateOrUpdateFormService(ctx context.Context, as *v1alpha1.MyNginx) (err error) {
	var (
		svc corev1.Service
	)
	svc.Name = as.Name
	svc.Namespace = as.Namespace
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		r.mutateService(as, &svc)
		return controllerutil.SetControllerReference(as, &svc, r.Scheme)
	})
	return err
}

func (r *resources) GetResource(ctx context.Context, namespace types.NamespacedName, as *v1alpha1.MyNginx) (err error) {
	return r.Client.Get(ctx, namespace, as)
}
