package appsodyapplication

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/application-stacks/runtime-component-operator/pkg/common"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	oputils "github.com/application-stacks/runtime-component-operator/pkg/utils"
	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	certmngrv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imageutil "github.com/openshift/library-go/pkg/image/imageutil"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	applicationsv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	setup(mgr)
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	reconciler := &ReconcileAppsodyApplication{ReconcilerBase: oputils.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("appsody-operator")),
		StackDefaults: map[string]appsodyv1beta1.AppsodyApplicationSpec{}, StackConstants: map[string]*appsodyv1beta1.AppsodyApplicationSpec{}}

	watchNamespaces, err := oputils.GetWatchNamespaces()
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

	configMap := &corev1.ConfigMap{}
	configMap.Namespace = ns
	configMap.Name = "appsody-operator"
	configMap.Data = common.DefaultOpConfig()
	err = reconciler.GetClient().Create(context.TODO(), configMap)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		log.Error(err, "Failed to create config map for the operator")
		os.Exit(1)
	}

	fData, err := ioutil.ReadFile("deploy/stack_defaults.yaml")
	if err != nil {
		log.Error(err, "Failed to read defaults config map from file")
		os.Exit(1)
	}

	configMap = &corev1.ConfigMap{}
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

func setup(mgr manager.Manager) {
	mgr.GetFieldIndexer().IndexField(&appsodyv1beta1.AppsodyApplication{}, indexFieldImageStreamName, func(obj runtime.Object) []string {
		instance := obj.(*appsodyv1beta1.AppsodyApplication)
		image, err := imageutil.ParseDockerImageReference(instance.Spec.ApplicationImage)
		if err == nil {
			imageNamespace := image.Namespace
			if imageNamespace == "" {
				imageNamespace = instance.Namespace
			}
			fullName := fmt.Sprintf("%s/%s", imageNamespace, image.Name)
			return []string{fullName}
		}
		return nil
	})
	mgr.GetFieldIndexer().IndexField(&appsodyv1beta1.AppsodyApplication{}, indexFieldBindingsResourceRef, func(obj runtime.Object) []string {
		instance := obj.(*appsodyv1beta1.AppsodyApplication)

		if instance.Spec.Bindings != nil && instance.Spec.Bindings.ResourceRef != "" {
			return []string{instance.Spec.Bindings.ResourceRef}
		}
		return nil
	})
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller

	reconciler := r.(*ReconcileAppsodyApplication)

	c, err := controller.New("appsodyapplication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	reconciler.SetController(c)

	watchNamespaces, err := oputils.GetWatchNamespaces()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	watchNamespacesMap := make(map[string]bool)
	for _, ns := range watchNamespaces {
		watchNamespacesMap[ns] = true
	}
	isClusterWide := oputils.IsClusterWide(watchNamespaces)

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
			return (isClusterWide || watchNamespacesMap[e.MetaOld.GetNamespace()])
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

	predSubResWithGenCheck := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return (isClusterWide || watchNamespacesMap[e.MetaOld.GetNamespace()]) && e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
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
	}, predSubResWithGenCheck)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsodyv1beta1.AppsodyApplication{},
	}, predSubResWithGenCheck)
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
		&oputils.EnqueueRequestsForServiceBinding{
			Client:          mgr.GetClient(),
			GroupName:       "appsody.dev",
			WatchNamespaces: watchNamespaces,
		})
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&EnqueueRequestsForCustomIndexField{
			Matcher: &BindingSecretMatcher{
				klient: mgr.GetClient(),
			},
		})
	if err != nil {
		return err
	}

	ok, _ := reconciler.IsGroupVersionSupported(imagev1.SchemeGroupVersion.String(), "ImageStream")
	if ok {
		c.Watch(
			&source.Kind{Type: &imagev1.ImageStream{}},
			&EnqueueRequestsForCustomIndexField{
				Matcher: &ImageStreamMatcher{
					Klient:          mgr.GetClient(),
					WatchNamespaces: watchNamespaces,
				},
			})
	}

	ok, _ = reconciler.IsGroupVersionSupported(routev1.SchemeGroupVersion.String(), "Route")
	if ok {
		c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appsodyv1beta1.AppsodyApplication{},
		}, predSubResource)
	}

	ok, _ = reconciler.IsGroupVersionSupported(servingv1alpha1.SchemeGroupVersion.String(), "Service")
	if ok {
		c.Watch(&source.Kind{Type: &servingv1alpha1.Service{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appsodyv1beta1.AppsodyApplication{},
		}, predSubResource)
	}

	ok, _ = reconciler.IsGroupVersionSupported(certmngrv1alpha2.SchemeGroupVersion.String(), "Certificate")
	if ok {
		c.Watch(&source.Kind{Type: &certmngrv1alpha2.Certificate{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appsodyv1beta1.AppsodyApplication{},
		}, predSubResource)
	}

	ok, _ = reconciler.IsGroupVersionSupported(prometheusv1.SchemeGroupVersion.String(), "ServiceMonitor")
	if ok {
		c.Watch(&source.Kind{Type: &prometheusv1.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appsodyv1beta1.AppsodyApplication{},
		}, predSubResource)
	}
	return nil
}

// blank assignment to verify that ReconcileAppsodyApplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppsodyApplication{}

// ReconcileAppsodyApplication reconciles a AppsodyApplication object
type ReconcileAppsodyApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	oputils.ReconcilerBase
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
			watchNamespaces, err = oputils.GetWatchNamespaces()
			if err != nil {
				reqLogger.Error(err, "Error getting watch namespace")
				return reconcile.Result{}, err
			}
		}
		// If the operator is running locally, use the first namespace in the `watchNamespaces`
		// `watchNamespaces` must have at least one item
		ns = watchNamespaces[0]
	}

	configMap, err := r.GetOpConfigMap("appsody-operator-defaults", ns)
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

	configMap, err = r.GetOpConfigMap("appsody-operator-constants", ns)
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

	configMap, err = r.GetOpConfigMap("appsody-operator", ns)
	if err != nil {
		log.Info("Failed to find appsody-operator config map")
		common.Config = common.DefaultOpConfig()
		configMap = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "appsody-operator", Namespace: ns}}
		configMap.Data = common.Config
	} else {
		common.Config.LoadFromConfigMap(configMap)
	}

	_, err = controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		configMap.Data = common.Config
		return nil
	})

	if err != nil {
		log.Info("Failed to update appsody-operator config map")
	}

	// Fetch the AppsodyApplication instance
	instance := &appsodyv1beta1.AppsodyApplication{}
	var ba common.BaseComponent
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
	_, err = oputils.Validate(instance)
	// If there's any validation error, don't bother with requeuing
	if err != nil {
		reqLogger.Error(err, "Error validating AppsodyApplication")
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		return reconcile.Result{}, nil
	}

	if r.IsApplicationSupported() {
		// Get labels from Application CRs selector and merge with instance labels
		existingAppLabels, err := r.GetSelectorLabelsFromApplications(instance)
		if err != nil {
			r.ManageError(errors.Wrapf(err, "unable to get %q Application CR selector's labels ", instance.Spec.ApplicationName), common.StatusConditionTypeReconciled, instance)
		}
		instance.Labels = oputils.MergeMaps(existingAppLabels, instance.Labels)
	} else {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported on the cluster", applicationsv1beta1.SchemeGroupVersion.String()))
	}

	if r.IsOpenShift() {
		// The order of items passed to the MergeMaps matters here! Annotations from GetOpenShiftAnnotations have higher importance. Otherwise,
		// it is not possible to override converted annotations.
		instance.Annotations = oputils.MergeMaps(instance.Annotations, oputils.GetOpenShiftAnnotations(instance))
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

	imageReferenceOld := instance.Status.ImageReference
	instance.Status.ImageReference = instance.Spec.ApplicationImage
	if r.IsOpenShift() {
		image, err := imageutil.ParseDockerImageReference(instance.Spec.ApplicationImage)
		if err == nil {
			imageStream := &imagev1.ImageStream{}
			imageNamespace := image.Namespace
			if imageNamespace == "" {
				imageNamespace = instance.Namespace
			}
			err = r.GetClient().Get(context.Background(), types.NamespacedName{Name: image.Name, Namespace: imageNamespace}, imageStream)
			if err == nil {
				image := imageutil.LatestTaggedImage(imageStream, image.Tag)
				if image != nil {
					instance.Status.ImageReference = image.DockerImageReference
				}
			} else if err != nil && !kerrors.IsNotFound(err) {
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
	}
	if imageReferenceOld != instance.Status.ImageReference {
		reqLogger.Info("Updating status.imageReference", "status.imageReference", instance.Status.ImageReference)
		err = r.UpdateStatus(instance)
		if err != nil {
			reqLogger.Error(err, "Error updating AppsodyApplication status")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
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

	if r.IsServiceBindingSupported() {
		result, err = r.ReconcileBindings(instance)
		if err != nil || result != (reconcile.Result{}) {
			return result, err
		}
	} else if instance.Spec.Bindings != nil {
		return r.ManageError(errors.New("failed to reconcile as the operator failed to find Service Binding CRDs"), common.StatusConditionTypeReconciled, instance)
	}
	resolvedBindingSecret, err := r.GetResolvedBindingSecret(ba)
	if err != nil {
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	if instance.Spec.ServiceAccountName == nil || *instance.Spec.ServiceAccountName == "" {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(serviceAccount, instance, func() error {
			oputils.CustomizeServiceAccount(serviceAccount, instance)
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

	isKnativeSupported, err := r.IsGroupVersionSupported(servingv1alpha1.SchemeGroupVersion.String(), "Service")
	if err != nil {
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if !isKnativeSupported {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported on the cluster", servingv1alpha1.SchemeGroupVersion.String()))
	}

	if instance.Spec.CreateKnativeService != nil && *instance.Spec.CreateKnativeService {
		// Clean up non-Knative resources
		resources := []runtime.Object{
			&corev1.Service{ObjectMeta: defaultMeta},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}},
			&appsv1.Deployment{ObjectMeta: defaultMeta},
			&appsv1.StatefulSet{ObjectMeta: defaultMeta},
			&autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta},
		}
		err = r.DeleteResources(resources)
		if err != nil {
			reqLogger.Error(err, "Failed to clean up non-Knative resources")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		if ok, _ := r.IsGroupVersionSupported(networkingv1beta1.SchemeGroupVersion.String(), "Ingress"); ok {
			r.DeleteResource(&networkingv1beta1.Ingress{ObjectMeta: defaultMeta})
		}
		if r.IsOpenShift() {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.DeleteResource(route)
			if err != nil {
				reqLogger.Error(err, "Failed to clean up non-Knative resource Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}

		if isKnativeSupported {
			ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(ksvc, instance, func() error {
				oputils.CustomizeKnativeService(ksvc, instance)
				oputils.CustomizeServiceBinding(resolvedBindingSecret, &ksvc.Spec.Template.Spec.PodSpec, instance)
				return nil
			})

			if err != nil {
				reqLogger.Error(err, "Failed to reconcile Knative Service")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
			return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
		}
		return r.ManageError(errors.New("failed to reconcile Knative service as operator could not find Knative CRDs"), common.StatusConditionTypeReconciled, instance)
	}

	if isKnativeSupported {
		ksvc := &servingv1alpha1.Service{ObjectMeta: defaultMeta}
		err = r.DeleteResource(ksvc)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Knative Service")
			r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	svc := &corev1.Service{ObjectMeta: defaultMeta}
	err = r.CreateOrUpdate(svc, instance, func() error {
		oputils.CustomizeService(svc, ba)
		svc.Annotations = oputils.MergeMaps(svc.Annotations, instance.Spec.Service.Annotations)
		monitoringEnabledLabelName := getMonitoringEnabledLabelName(ba)
		if instance.Spec.Monitoring != nil {
			svc.Labels[monitoringEnabledLabelName] = "true"
		} else {
			if _, ok := svc.Labels[monitoringEnabledLabelName]; ok {
				delete(svc.Labels, monitoringEnabledLabelName)
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
			oputils.CustomizeService(svc, instance)
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
			oputils.CustomizeStatefulSet(statefulSet, instance)
			oputils.CustomizePodSpec(&statefulSet.Spec.Template, instance)
			oputils.CustomizePersistence(statefulSet, instance)
			oputils.CustomizeServiceBinding(resolvedBindingSecret, &statefulSet.Spec.Template.Spec, instance)
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
			oputils.CustomizeDeployment(deploy, instance)
			oputils.CustomizePodSpec(&deploy.Spec.Template, instance)
			oputils.CustomizeServiceBinding(resolvedBindingSecret, &deploy.Spec.Template.Spec, instance)
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
			oputils.CustomizeHPA(hpa, instance)
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

	if ok, err := r.IsGroupVersionSupported(routev1.SchemeGroupVersion.String(), "Route"); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", routev1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Expose != nil && *instance.Spec.Expose {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(route, instance, func() error {
				key, cert, caCert, destCACert, err := r.GetRouteTLSValues(ba)
				if err != nil {
					return err
				}
				oputils.CustomizeRoute(route, ba, key, cert, caCert, destCACert)
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

		if ok, err := r.IsGroupVersionSupported(networkingv1beta1.SchemeGroupVersion.String(), "Ingress"); err != nil {
			reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", networkingv1beta1.SchemeGroupVersion.String()))
			r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		} else if ok {
			if instance.Spec.Expose != nil && *instance.Spec.Expose {
				ing := &networkingv1beta1.Ingress{ObjectMeta: defaultMeta}
				err = r.CreateOrUpdate(ing, instance, func() error {
					oputils.CustomizeIngress(ing, instance)
					return nil
				})
				if err != nil {
					reqLogger.Error(err, "Failed to reconcile Ingress")
					return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
				}
			} else {
				ing := &networkingv1beta1.Ingress{ObjectMeta: defaultMeta}
				err = r.DeleteResource(ing)
				if err != nil {
					reqLogger.Error(err, "Failed to delete Ingress")
					return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
				}
			}
		}
	}

	if ok, err = r.IsGroupVersionSupported(prometheusv1.SchemeGroupVersion.String(), "ServiceMonitor"); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", prometheusv1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Monitoring != nil && (instance.Spec.CreateKnativeService == nil || !*instance.Spec.CreateKnativeService) {
			sm := &prometheusv1.ServiceMonitor{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(sm, instance, func() error {
				oputils.CustomizeServiceMonitor(sm, instance)
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
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported", prometheusv1.SchemeGroupVersion.String()))
	}

	return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
}

func getMonitoringEnabledLabelName(ba common.BaseComponent) string {
	return "monitor." + ba.GetGroupName() + "/enabled"
}
