package utils

import (
	"context"
	"fmt"
	"math"
	"time"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ReconcilerBase base reconsiler with some common behaviour
type ReconcilerBase struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//NewReconcilerBase creates a new ReconsilerBase
func NewReconcilerBase(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) ReconcilerBase {
	return ReconcilerBase{
		client:   client,
		scheme:   scheme,
		recorder: recorder,
	}
}

// GetClient returns client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

var log = logf.Log.WithName("utils")

// CreateOrUpdate ...
func (r *ReconcilerBase) CreateOrUpdate(obj metav1.Object, owner metav1.Object, reconcile func() error) error {

	mutate := func(o runtime.Object) error {
		err := reconcile()
		return err
	}

	controllerutil.SetControllerReference(owner, obj, r.scheme)
	runtimeObj, ok := obj.(runtime.Object)
	if !ok {
		return fmt.Errorf("is not a %T a runtime.Object", obj)
	}
	result, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), runtimeObj, mutate)
	if err != nil {
		return err
	}

	var gvk schema.GroupVersionKind
	gvk, err = apiutil.GVKForObject(runtimeObj, r.scheme)
	if err == nil {
		log.Info("Reconciled", "Kind", gvk.Kind, "Name", obj.GetName(), "Status", result)
	}

	return err
}

// DeleteResource deletes kubernetes resource
func (r *ReconcilerBase) DeleteResource(obj runtime.Object) error {
	err := r.client.Delete(context.TODO(), obj)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "Unable to delete object ", "object", obj)
		return err
	}
	return nil
}

// DeleteResources ...
func (r *ReconcilerBase) DeleteResources(resources []runtime.Object) error {
	for i := range resources {
		err := r.DeleteResource(resources[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAppsodyOpConfigMap ...
func (r *ReconcilerBase) GetAppsodyOpConfigMap(ns string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: "appsody-operator", Namespace: ns}, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// ManageError ...
func (r *ReconcilerBase) ManageError(issue error, conditionType appsodyv1alpha1.AppsodyApplicationStatusConditionType, cr *appsodyv1alpha1.AppsodyApplication) (reconcile.Result, error) {
	r.GetRecorder().Event(cr, "Warning", "ProcessingError", issue.Error())

	condition := GetCondition(conditionType, cr.Status.Conditions)
	lastUpdate := condition.LastUpdateTime.Time
	lastStatus := condition.Status

	statusCondition := appsodyv1alpha1.AppsodyApplicationStatusCondition{
		//LastTransitionTime: ,
		LastUpdateTime: metav1.Now(),
		Reason:         issue.Error(),
		Type:           conditionType,
		// Message: ,
		// Status: ,
	}

	SetCondition(statusCondition, cr.Status.Conditions)

	err := r.GetClient().Status().Update(context.Background(), cr)
	if err != nil {
		log.Error(err, "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	var retryInterval time.Duration
	if lastUpdate.IsZero() || lastStatus == "Success" {
		retryInterval = time.Second
	} else {
		retryInterval = statusCondition.LastUpdateTime.Sub(lastUpdate).Round(time.Second)
	}

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *ReconcilerBase) ManageSuccess(conditionType appsodyv1alpha1.AppsodyApplicationStatusConditionType, cr *appsodyv1alpha1.AppsodyApplication) (reconcile.Result, error) {
	statusCondition := appsodyv1alpha1.AppsodyApplicationStatusCondition{
		//LastTransitionTime: ,
		LastUpdateTime: metav1.Now(),
		Type:           conditionType,
		// Message: ,
		// Status: ,
	}

	SetCondition(statusCondition, cr.Status.Conditions)
	err := r.GetClient().Status().Update(context.Background(), cr)
	if err != nil {
		log.Error(err, "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}
	return reconcile.Result{}, nil
}
