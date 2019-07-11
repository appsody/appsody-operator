package utils

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ReconcilerBase base reconsiler with some common behaviour
type ReconcilerBase struct {
	client client.Client
	scheme *runtime.Scheme
}

//NewReconcilerBase creates a new ReconsilerBase
func NewReconcilerBase(client client.Client, scheme *runtime.Scheme) ReconcilerBase {
	return ReconcilerBase{
		client: client,
		scheme: scheme,
	}
}

// GetClient returns client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

var log = logf.Log.WithName("utils")

// CreateOrUpdateResource creates a kubernetes resource if it doesn't exists or updates existing one
func (r *ReconcilerBase) CreateOrUpdateResource(obj metav1.Object, owner metav1.Object) error {
	runtimeObj, ok := obj.(runtime.Object)
	if !ok {
		return fmt.Errorf("is not a %T a runtime.Object", obj)
	}

	existingObj := unstructured.Unstructured{}
	gvk, err := apiutil.GVKForObject(runtimeObj, r.scheme)
	if err != nil {

	}

	existingObj.SetGroupVersionKind(schema.GroupVersionKind{Group: gvk.Group, Kind: gvk.Kind, Version: gvk.Version})

	log.Info("Existing object", "Object", runtimeObj.GetObjectKind())

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, &existingObj)
	if err != nil && apierrors.IsNotFound(err) {
		controllerutil.SetControllerReference(owner, obj, r.scheme)
		err = r.client.Create(context.TODO(), runtimeObj)
		if err != nil {
			log.Error(err, "unable to create object", "object", runtimeObj)
		}
		return err
	} else if err == nil {
		obj.SetResourceVersion(existingObj.GetResourceVersion())
		controllerutil.SetControllerReference(owner, obj, r.scheme)
		err = r.client.Update(context.TODO(), runtimeObj)
		if err != nil {
			log.Error(err, "unable to update object", "object", runtimeObj)
		}
		return err
	}
	return err
}

// DeleteResource deletes kubernetes resource
func (r *ReconcilerBase) DeleteResource(obj runtime.Object) error {
	err := r.client.Delete(context.TODO(), obj, nil)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "unable to delete object ", "object", obj)
		return err
	}
	return nil
}
