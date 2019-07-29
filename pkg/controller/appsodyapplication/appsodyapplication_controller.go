package appsodyapplication

import (
	"context"
	"encoding/json"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	appsodyutils "github.com/appsody-operator/pkg/utils"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	return &ReconcileAppsodyApplication{ReconcilerBase: appsodyutils.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetRecorder("appsody-operator")),
		StackDefaults: map[string]appsodyv1alpha1.AppsodyApplicationSpec{}}
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

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{})
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
	StackDefaults map[string]appsodyv1alpha1.AppsodyApplicationSpec
}

// Reconcile reads that state of the cluster for a AppsodyApplication object and makes changes based on the state read
// and what is in the AppsodyApplication.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppsodyApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppsodyApplication")

	if request.Name == "appsody-operator" {
		configMap, err := r.GetAppsodyOpConfigMap(request.Namespace)
		if err == nil {
			for stack, values := range configMap.Data {
				var defaults appsodyv1alpha1.AppsodyApplicationSpec
				unerr := json.Unmarshal([]byte(values), &defaults)
				if unerr != nil {
					log.Error(unerr, "Failed to parse config map defaults")
				} else {
					r.StackDefaults[stack] = defaults
				}
			}
		}
		return reconcile.Result{}, nil

	}

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
	appsodyutils.InitAndValidate(instance, r.StackDefaults[instance.Spec.Stack])
	err = r.GetClient().Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "Error updating AppsodyApplication")
		return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
	}

	defaultMeta := metav1.ObjectMeta{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}

	if instance.Spec.ServiceAccountName == nil || *instance.Spec.ServiceAccountName == "" {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(serviceAccount, instance, func() error {
			appsodyutils.CustomizeServiceAccount(serviceAccount, instance)
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile ServiceAccount")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}
	} else {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.DeleteResource(serviceAccount)
		if err != nil {
			reqLogger.Error(err, "Failed to delete ServiceAccount")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}
	}

	if instance.Spec.CreateKnativeService != nil && *instance.Spec.CreateKnativeService {
		ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(ksvc, instance, func() error {
			appsodyutils.CustomizeKnativeService(ksvc, instance)
			return nil
		})

		if err != nil {
			reqLogger.Error(err, "Failed to reconcile Knative Service")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}

		// Clean up non-Knative resources
		resources := []runtime.Object{
			&corev1.Service{ObjectMeta: defaultMeta},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}},
			&appsv1.Deployment{ObjectMeta: defaultMeta},
			&appsv1.StatefulSet{ObjectMeta: defaultMeta},
			&routev1.Route{ObjectMeta: defaultMeta},
		}
		err = r.DeleteResources(resources)
		if err != nil {
			reqLogger.Error(err, "Failed to clean up non-Knative resources")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}

		return r.ManageSuccess(appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
	}

	ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
	err = r.DeleteResource(ksvc)
	if err != nil {
		reqLogger.Error(err, "Failed to delete Knative Service")
	}

	svc := &corev1.Service{ObjectMeta: defaultMeta}
	err = r.CreateOrUpdate(svc, instance, func() error {
		appsodyutils.CustomizeService(svc, instance)
		return nil
	})
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Service")
		return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
	}

	if instance.Spec.Storage != nil {
		// Delete Deployment if exists
		deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
		err = r.DeleteResource(deploy)

		if err != nil {
			reqLogger.Error(err, "Failed to delete Deployment")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		} else {
			svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
			err = r.CreateOrUpdate(svc, instance, func() error {
				appsodyutils.CustomizeService(svc, instance)
				svc.Spec.ClusterIP = corev1.ClusterIPNone
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile headless Service")
				return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
			}

			statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(statefulSet, instance, func() error {
				statefulSet.Spec.Replicas = instance.Spec.Replicas
				statefulSet.Spec.ServiceName = instance.Name + "-headless"
				statefulSet.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/name": instance.Name,
					},
				}
				appsodyutils.CustomizePersistence(statefulSet, instance)
				appsodyutils.CustomizePodSpec(&statefulSet.Spec.Template, instance)
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile StatefulSet")
				return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
			}
		}
	} else {
		// Delete StatefulSet if exists
		statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
		err = r.DeleteResource(statefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Statefulset")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}

		// Delete StatefulSet if exists
		headlesssvc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
		err = r.DeleteResource(headlesssvc)

		if err != nil {
			reqLogger.Error(err, "Failed to delete headless Service")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		} else {
			deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(deploy, instance, func() error {
				deploy.Spec.Replicas = instance.Spec.Replicas
				deploy.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/name": instance.Name,
					},
				}
				appsodyutils.CustomizePodSpec(&deploy.Spec.Template, instance)
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile Deployment")
				return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
			}
		}
	}

	if instance.Spec.Expose != nil && *instance.Spec.Expose {
		route := &routev1.Route{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(route, instance, func() error {
			appsodyutils.CustomizeRoute(route, instance)
			return nil
		})
		if err != nil {
			log.Error(err, "Failed to create Route")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}

	} else {
		route := &routev1.Route{ObjectMeta: defaultMeta}
		err = r.DeleteResource(route)
		if err != nil {
			log.Error(err, "Failed to delete Route")
			return r.ManageError(err, appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
		}
	}

	return r.ManageSuccess(appsodyv1alpha1.AppsodyApplicationStatusConditionTypeReconciled, instance)
}
