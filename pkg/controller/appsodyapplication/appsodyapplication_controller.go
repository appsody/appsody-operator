package appsodyapplication

import (
	"context"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	appsodyutils "github.com/appsody-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_appsodyapplication")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppsodyApplication Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAppsodyApplication{ReconcilerBase: appsodyutils.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme())}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appsodyapplication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppsodyApplication
	err = c.Watch(&source.Kind{Type: &appsodyv1alpha1.AppsodyApplication{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner AppsodyApplication
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1alpha1.AppsodyApplication{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAppsodyApplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppsodyApplication{}

// ReconcileAppsodyApplication reconciles a AppsodyApplication object
type ReconcileAppsodyApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	appsodyutils.ReconcilerBase
}

// Reconcile reads that state of the cluster for a AppsodyApplication object and makes changes based on the state read
// and what is in the AppsodyApplication.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppsodyApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppsodyApplication")

	// Fetch the AppsodyApplication instance
	instance := &appsodyv1alpha1.AppsodyApplication{}
	err := r.GetClient().Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Spec.ServiceAccountName == "" {
		serviceAccount := appsodyutils.GenerateSeviceAccount(instance)
		err = r.CreateOrUpdate(serviceAccount, instance, func() error {
			if instance.Spec.PullSecret != "" {
				serviceAccount.ImagePullSecrets[0].Name = instance.Spec.PullSecret
			}
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to create ServiceAccount")
		}
	}

	deploy := appsodyutils.GenerateDeployment(instance)
	r.CreateOrUpdate(deploy, instance, func() error {
		deploy.Spec.Replicas = instance.Spec.Replicas
		deploy.Spec.Template.Spec.Containers[0].Image = instance.Spec.ApplicationImage
		deploy.Spec.Template.Spec.Containers[0].Resources = instance.Spec.ResourceConstraints
		deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = instance.Spec.ReadinessProbe
		deploy.Spec.Template.Spec.Containers[0].LivenessProbe = instance.Spec.LivenessProbe
		deploy.Spec.Template.Spec.Containers[0].VolumeMounts = instance.Spec.VolumeMounts
		deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy = instance.Spec.PullPolicy
		deploy.Spec.Template.Spec.Containers[0].Env = instance.Spec.Env
		deploy.Spec.Template.Spec.Containers[0].EnvFrom = instance.Spec.EnvFrom
		deploy.Spec.Template.Spec.Volumes = instance.Spec.Volumes
		if instance.Spec.ServiceAccountName != "" {
			deploy.Spec.Template.Spec.ServiceAccountName = instance.Spec.ServiceAccountName
		}
		return nil
	})

	if err != nil {
		reqLogger.Error(err, "Failed to create Deployment")
	}

	svc := appsodyutils.GenerateService(instance)
	r.CreateOrUpdate(svc, instance, func() error {
		svc.Spec.Ports[0].Port = instance.Spec.Service.Port
		svc.Spec.Type = instance.Spec.Service.Type
		return nil
	})
	if err != nil {

		reqLogger.Error(err, "Failed to create Service")
	}
	return reconcile.Result{}, nil
}
