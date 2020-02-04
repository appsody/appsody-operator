package appsodyapplication

import (
	"context"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/pkg/common"
	appsodyutils "github.com/appsody/appsody-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ handler.EventHandler = &EnqueueRequestsForImageStream{}

const (
	indexFieldImageStreamName      = "spec.applicationImageStream.name"
	indexFieldImageStreamNamespace = "spec.applicationImageStream.namespace"
)

// EnqueueRequestsForImageStream enqueues reconcile Requests for applications affected by the secrets that
// EventHandler is called for
type EnqueueRequestsForImageStream struct {
	handler.Funcs
	WatchNamespaces    []string
	GroupName          string
	Client             client.Client
	ApplicationFactory func() common.BaseApplication
}

// Update implements EventHandler
func (e *EnqueueRequestsForImageStream) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.MetaNew, q)
}

// Delete implements EventHandler
func (e *EnqueueRequestsForImageStream) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Meta, q)
}

// Generic implements EventHandler
func (e *EnqueueRequestsForImageStream) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Meta, q)
}

// handle common implementation to enqueue reconcile Requests for applications affected by the secrets that
// EventHandler is called for
func (e *EnqueueRequestsForImageStream) handle(evtMeta metav1.Object, q workqueue.RateLimitingInterface) {
	apps, _ := e.matchApplication(evtMeta)
	for _, app := range apps {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: app.Namespace,
				Name:      app.Name,
			}})
	}
}

// matchApplication returns the NamespacedName of all applications using the input ImageStreamTag
func (e *EnqueueRequestsForImageStream) matchApplication(imageStreamTag metav1.Object) ([]appsodyv1beta1.AppsodyApplication, error) {
	apps := []appsodyv1beta1.AppsodyApplication{}
	var namespaces []string

	if appsodyutils.IsClusterWide(e.WatchNamespaces) {
		nsList := &corev1.NamespaceList{}
		err := e.Client.List(context.Background(), nsList, client.InNamespace(""))
		if err != nil {
			return nil, err
		}
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		namespaces = e.WatchNamespaces
	}

	for _, ns := range namespaces {
		appList := &appsodyv1beta1.AppsodyApplicationList{}
		err := e.Client.List(context.Background(),
			appList,
			client.InNamespace(ns),
			client.MatchingField(indexFieldImageStreamName, imageStreamTag.GetName()),
			client.MatchingField(indexFieldImageStreamNamespace, imageStreamTag.GetNamespace()))
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		apps = append(apps, appList.Items...)
	}
	return apps, nil
}
