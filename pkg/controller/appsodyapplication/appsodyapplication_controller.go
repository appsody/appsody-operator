package appsodyapplication

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/appsody/appsody-operator/pkg/common"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	appsodyutils "github.com/appsody/appsody-operator/pkg/utils"
	certmngrv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"
)

var log = logf.Log.WithName("controller_appsodyapplication")

// Holds a list of namespaces the operator will be watching
var watchNamespaces []string

// Add creates a new AppsodyApplication Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	reconciler := &ReconcileAppsodyApplication{ReconcilerBase: appsodyutils.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("appsody-operator")),
		StackDefaults: map[string]appsodyv1beta1.AppsodyApplicationSpec{}, StackConstants: map[string]*appsodyv1beta1.AppsodyApplicationSpec{}}

	watchNamespaces, err := appsodyutils.GetWatchNamespaces()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}
	log.Info("newReconciler", "watchNamespaces", watchNamespaces)

	ns, err := k8sutil.GetOperatorNamespace()
	// When running the operator locally, `ns` will be empty string
	if ns == "" {
		// If the operator is running locally, use the first namespace in the `watchNamespaces`
		// `watchNamespaces` must have at least one item
		ns = watchNamespaces[0]
	}

	fData, err := ioutil.ReadFile("deploy/stack_defaults.yaml")
	if err != nil {
		log.Error(err, "Failed to read defaults config map from file")
		os.Exit(1)
	}

	configMap := &corev1.ConfigMap{}
	err = yaml.Unmarshal(fData, configMap)
	if err != nil {
		log.Error(err, "Failed to parse defaults config map from file")
		os.Exit(1)
	}
	configMap.Namespace = ns
	err = reconciler.GetClient().Create(context.TODO(), configMap)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		log.Error(err, "Failed to create defaults config map in the cluster")
		os.Exit(1)
	}

	fData, err = ioutil.ReadFile("deploy/stack_constants.yaml")
	if err != nil {
		log.Error(err, "Failed to read constants config map from file")
		os.Exit(1)
	}

	configMap = &corev1.ConfigMap{}
	err = yaml.Unmarshal(fData, configMap)
	if err != nil {
		log.Error(err, "Failed to parse constants config map from file")
		os.Exit(1)
	}
	configMap.Namespace = ns
	err = reconciler.GetClient().Create(context.TODO(), configMap)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		log.Error(err, "Failed to create constants config map in the cluster")
		os.Exit(1)
	}

	return reconciler
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appsodyapplication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	watchNamespaces, err := appsodyutils.GetWatchNamespaces()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	watchNamespacesMap := make(map[string]bool)
	for _, ns := range watchNamespaces {
		watchNamespacesMap[ns] = true
	}
	isClusterWide := appsodyutils.IsClusterWide(watchNamespaces)

	log.V(1).Info("Adding a new controller", "watchNamespaces", watchNamespaces, "isClusterWide", isClusterWide)

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration() && (isClusterWide || watchNamespacesMap[e.MetaOld.GetNamespace()])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Meta.GetNamespace()]
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Meta.GetNamespace()]
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Meta.GetNamespace()]
		},
	}

	// Watch for changes to primary resource AppsodyApplication
	err = c.Watch(&source.Kind{Type: &appsodyv1beta1.AppsodyApplication{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	predSubResource := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration() && (isClusterWide || watchNamespacesMap[e.MetaOld.GetNamespace()])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Meta.GetNamespace()]
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &autoscalingv1.HorizontalPodAutoscaler{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		OwnerType: &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&appsodyutils.EnqueueRequestsForServiceBinding{
			Client:          mgr.GetClient(),
			GroupName:       "appsody.dev",
			WatchNamespaces: watchNamespaces,
		})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)

	err = c.Watch(&source.Kind{Type: &servingv1alpha1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)

	err = c.Watch(&source.Kind{Type: &certmngrv1alpha2.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResource)

	return nil
}

// blank assignment to verify that ReconcileAppsodyApplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppsodyApplication{}

// ReconcileAppsodyApplication reconciles a AppsodyApplication object
type ReconcileAppsodyApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	appsodyutils.ReconcilerBase
	StackDefaults   map[string]appsodyv1beta1.AppsodyApplicationSpec
	StackConstants  map[string]*appsodyv1beta1.AppsodyApplicationSpec
	lastDefautsRV   string
	lastConstantsRV string
}

// Reconcile reads that state of the cluster for a AppsodyApplication object and makes changes based on the state read
// and what is in the AppsodyApplication.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppsodyApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppsodyApplication")

	ns, err := k8sutil.GetOperatorNamespace()
	// When running the operator locally, `ns` will be empty string
	if ns == "" {
		// Since this method can be called directly from unit test, populate `watchNamespaces`.
		if watchNamespaces == nil {
			watchNamespaces, err = appsodyutils.GetWatchNamespaces()
			if err != nil {
				reqLogger.Error(err, "Error getting watch namespace")
				return reconcile.Result{}, err
			}
		}
		// If the operator is running locally, use the first namespace in the `watchNamespaces`
		// `watchNamespaces` must have at least one item
		ns = watchNamespaces[0]
	}

	configMap, err := r.GetAppsodyOpConfigMap("appsody-operator-defaults", ns)
	if err != nil {
		log.Info("Failed to find config map defaults in namespace " + ns)
	} else {
		if r.lastDefautsRV != configMap.ResourceVersion {
			for k := range r.StackDefaults {
				delete(r.StackDefaults, k)
			}
			for stack, values := range configMap.Data {
				var defaults appsodyv1beta1.AppsodyApplicationSpec
				unerr := yaml.Unmarshal([]byte(values), &defaults)
				if unerr != nil {
					reqLogger.Error(unerr, "Failed to parse config map defaults")
				} else {
					r.StackDefaults[stack] = defaults
				}
			}
		}
		r.lastDefautsRV = configMap.ResourceVersion
	}

	configMap, err = r.GetAppsodyOpConfigMap("appsody-operator-constants", ns)
	if err != nil {
		log.Info("Failed to find config map constants")
	} else {
		if r.lastConstantsRV != configMap.ResourceVersion {
			for k := range r.StackConstants {
				delete(r.StackConstants, k)
			}
			for stack, values := range configMap.Data {
				var constants appsodyv1beta1.AppsodyApplicationSpec
				unerr := yaml.Unmarshal([]byte(values), &constants)
				if unerr != nil {
					reqLogger.Error(unerr, "Failed to parse config map constants")
				} else {
					r.StackConstants[stack] = &constants
				}
			}
		}
		r.lastConstantsRV = configMap.ResourceVersion
	}

	// Fetch the AppsodyApplication instance
	instance := &appsodyv1beta1.AppsodyApplication{}
	var ba common.BaseApplication
	ba = instance
	err = r.GetClient().Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	stackDefaults, ok := r.StackDefaults[instance.Spec.Stack]
	if ok {
		_, ok = r.StackConstants[instance.Spec.Stack]
		if ok {
			instance.Initialize(stackDefaults, r.StackConstants[instance.Spec.Stack])
		} else {
			instance.Initialize(stackDefaults, r.StackConstants["generic"])
		}
	} else {
		stackDefaults, ok = r.StackDefaults["generic"]
		if !ok {
			err = fmt.Errorf("Failed to find stack neither `%v` nor `generic` in the ConfigMap holding default values", instance.Spec.Stack)
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
		_, ok = r.StackConstants[instance.Spec.Stack]
		if ok {
			instance.Initialize(stackDefaults, r.StackConstants[instance.Spec.Stack])
		} else {
			instance.Initialize(stackDefaults, r.StackConstants["generic"])
		}
	}

	_, err = appsodyutils.Validate(instance)
	// If there's any validation error, don't bother with requeuing
	if err != nil {
		reqLogger.Error(err, "Error validating AppsodyApplication")
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		return reconcile.Result{}, nil
	}

	currentGen := instance.Generation
	err = r.GetClient().Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Error updating AppsodyApplication")
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	if currentGen == 1 {
		return reconcile.Result{}, nil
	}

	defaultMeta := metav1.ObjectMeta{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}

	result, err := r.ReconcileProvides(instance)
	if err != nil || result != (reconcile.Result{}) {
		return result, err
	}

	result, err = r.ReconcileConsumes(instance)
	if err != nil || result != (reconcile.Result{}) {
		return result, err
	}
	result, err = r.ReconcileCertificate(instance)
	if err != nil || result != (reconcile.Result{}) {
		return result, nil
	}

	if instance.Spec.ServiceAccountName == nil || *instance.Spec.ServiceAccountName == "" {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(serviceAccount, instance, func() error {
			appsodyutils.CustomizeServiceAccount(serviceAccount, instance)
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile ServiceAccount")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.DeleteResource(serviceAccount)
		if err != nil {
			reqLogger.Error(err, "Failed to delete ServiceAccount")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	if instance.Spec.CreateKnativeService != nil && *instance.Spec.CreateKnativeService {
		ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(ksvc, instance, func() error {
			appsodyutils.CustomizeKnativeService(ksvc, instance)
			if r.IsOpenShift() {
				ksvc.Spec.Template.ObjectMeta.Annotations = appsodyutils.MergeMaps(appsodyutils.GetConnectToAnnotation(instance), ksvc.Spec.Template.ObjectMeta.Annotations)
			}
			return nil
		})

		if err != nil {
			reqLogger.Error(err, "Failed to reconcile Knative Service")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		// Clean up non-Knative resources
		resources := []runtime.Object{
			&corev1.Service{ObjectMeta: defaultMeta},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}},
			&appsv1.Deployment{ObjectMeta: defaultMeta},
			&appsv1.StatefulSet{ObjectMeta: defaultMeta},
			&routev1.Route{ObjectMeta: defaultMeta},
			&autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta},
		}
		err = r.DeleteResources(resources)
		if err != nil {
			reqLogger.Error(err, "Failed to clean up non-Knative resources")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
	}

	// Check if Knative is supported and delete Knative service if supported
	if ok, err = r.IsGroupVersionSupported(servingv1alpha1.SchemeGroupVersion.String()); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", servingv1alpha1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
		err = r.DeleteResource(ksvc)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Knative Service")
			r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported. Skip deleting the resource", servingv1alpha1.SchemeGroupVersion.String()))
	}

	svc := &corev1.Service{ObjectMeta: defaultMeta}
	err = r.CreateOrUpdate(svc, instance, func() error {
		appsodyutils.CustomizeService(svc, ba)
		svc.Annotations = appsodyutils.MergeMaps(svc.Annotations, instance.Spec.Service.Annotations)
		if instance.Spec.Monitoring != nil {
			svc.Labels["app."+ba.GetGroupName()+"/monitor"] = "true"
		} else {
			if _, ok := svc.Labels["app."+ba.GetGroupName()+"/monitor"]; ok {
				delete(svc.Labels, "app."+ba.GetGroupName()+"/monitor")
			}
		}
		return nil
	})
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Service")
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	if instance.Spec.Storage != nil {
		// Delete Deployment if exists
		deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
		err = r.DeleteResource(deploy)

		if err != nil {
			reqLogger.Error(err, "Failed to delete Deployment")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
		err = r.CreateOrUpdate(svc, instance, func() error {
			appsodyutils.CustomizeService(svc, instance)
			svc.Spec.ClusterIP = corev1.ClusterIPNone
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile headless Service")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(statefulSet, instance, func() error {
			appsodyutils.CustomizeStatefulSet(statefulSet, instance)
			appsodyutils.CustomizePodSpec(&statefulSet.Spec.Template, instance)
			appsodyutils.CustomizePersistence(statefulSet, instance)
			if r.IsOpenShift() {
				statefulSet.Annotations = appsodyutils.MergeMaps(appsodyutils.GetConnectToAnnotation(instance), statefulSet.Annotations)
			}
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile StatefulSet")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

	} else {
		// Delete StatefulSet if exists
		statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
		err = r.DeleteResource(statefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Statefulset")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		// Delete StatefulSet if exists
		headlesssvc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
		err = r.DeleteResource(headlesssvc)

		if err != nil {
			reqLogger.Error(err, "Failed to delete headless Service")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
		deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(deploy, instance, func() error {
			appsodyutils.CustomizeDeployment(deploy, instance)
			appsodyutils.CustomizePodSpec(&deploy.Spec.Template, instance)
			if r.IsOpenShift() {
				deploy.Annotations = appsodyutils.MergeMaps(appsodyutils.GetConnectToAnnotation(instance), deploy.Annotations)
			}
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile Deployment")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

	}

	if instance.Spec.Autoscaling != nil {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(hpa, instance, func() error {
			appsodyutils.CustomizeHPA(hpa, instance)
			return nil
		})

		if err != nil {
			reqLogger.Error(err, "Failed to reconcile HorizontalPodAutoscaler")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta}
		err = r.DeleteResource(hpa)
		if err != nil {
			reqLogger.Error(err, "Failed to delete HorizontalPodAutoscaler")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	if ok, err := r.IsGroupVersionSupported(routev1.SchemeGroupVersion.String()); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", routev1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Expose != nil && *instance.Spec.Expose {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(route, instance, func() error {
				//Check if the CA available in a secret
				destCACert := ""
				caCert := ""
				cert := ""
				key := ""
				if instance.Spec.Service != nil && instance.Spec.Service.Certificate != nil {
					tlsSecret := &corev1.Secret{}
					r.GetClient().Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-svc-tls", Namespace: instance.Namespace}, tlsSecret)
					caCrt, ok := tlsSecret.Data["ca.crt"]
					if ok {
						destCACert = string(caCrt)
					}
				}
				if instance.Spec.Route != nil && instance.Spec.Route.Certificate != nil {
					tlsSecret := &corev1.Secret{}
					r.GetClient().Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-route-tls", Namespace: instance.Namespace}, tlsSecret)
					v, ok := tlsSecret.Data["ca.crt"]
					if ok {
						caCert = string(v)
					}
					v, ok = tlsSecret.Data["tls.crt"]
					if ok {
						cert = string(v)
					}
					v, ok = tlsSecret.Data["tls.key"]
					if ok {
						key = string(v)
					}
				}
				appsodyutils.CustomizeRoute(route, ba, key, cert, caCert, destCACert)

				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		} else {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.DeleteResource(route)
			if err != nil {
				reqLogger.Error(err, "Failed to delete Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
	} else {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported", routev1.SchemeGroupVersion.String()))
	}

	if ok, err = r.IsGroupVersionSupported(prometheusv1.SchemeGroupVersion.String()); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", routev1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Monitoring != nil && (instance.Spec.CreateKnativeService == nil || !*instance.Spec.CreateKnativeService) {
			sm := &prometheusv1.ServiceMonitor{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(sm, instance, func() error {
				appsodyutils.CustomizeServiceMonitor(sm, instance)
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile ServiceMonitor")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		} else {
			sm := &prometheusv1.ServiceMonitor{ObjectMeta: defaultMeta}
			err = r.DeleteResource(sm)
			if err != nil {
				reqLogger.Error(err, "Failed to delete ServiceMonitor")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}

	} else {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported", routev1.SchemeGroupVersion.String()))
	}

	return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
}
