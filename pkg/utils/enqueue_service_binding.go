package utils

import (
	"context"
	"strings"

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

var _ handler.EventHandler = &EnqueueRequestsForServiceBinding{}

// EnqueueRequestsForServiceBinding enqueues Requests for the Base Application objects that rely on the secret
// the event is called.
type EnqueueRequestsForServiceBinding struct {
	handler.Funcs
	WatchNamespaces []string
	GroupName       string
	Client          client.Client
}

var logger = log.WithName("EnqueueRequestsForServiceBinding")

// Update implements EventHandler
func (e *EnqueueRequestsForServiceBinding) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.MetaNew, q)
}

// Delete implements EventHandler
func (e *EnqueueRequestsForServiceBinding) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Meta, q)
}

// Generic implements EventHandler
func (e *EnqueueRequestsForServiceBinding) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Meta, q)
}

func (e *EnqueueRequestsForServiceBinding) handle(evtMeta metav1.Object, q workqueue.RateLimitingInterface) {
	apps, _ := e.matchApplication(evtMeta)
	for _, app := range apps {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: app.Namespace,
				Name:      app.Name,
			}})
	}
}

func (e *EnqueueRequestsForServiceBinding) matchApplication(mSecret metav1.Object) ([]types.NamespacedName, error) {
	dependents, err := e.getDependentSecrets(mSecret)
	if err != nil {
		return nil, err
	}

	matched := []types.NamespacedName{}
	tmpSecret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: mSecret.GetName(), Namespace: mSecret.GetNamespace(), Annotations: mSecret.GetAnnotations()}}
	dependents = append(dependents, tmpSecret)
	for _, depSecret := range dependents {
		if depSecret.Annotations != nil {
			if consumedBy, ok := depSecret.Annotations["service."+e.GroupName+"/consumed-by"]; ok {
				for _, app := range strings.Split(consumedBy, ",") {
					matched = append(matched, types.NamespacedName{Name: app, Namespace: depSecret.Namespace})
				}
			}
		}
	}
	return matched, nil
}

func (e *EnqueueRequestsForServiceBinding) getDependentSecrets(secret metav1.Object) ([]corev1.Secret, error) {
	dependents := []corev1.Secret{}
	var namespaces []string

	if e.isClusterWide() {
		nsList := &corev1.NamespaceList{}
		err := e.Client.List(context.Background(), nsList)
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
		depSecret := &corev1.Secret{}
		err := e.Client.Get(context.Background(), client.ObjectKey{Name: secret.GetName(), Namespace: ns}, depSecret)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		dependents = append(dependents, *depSecret)
	}

	return dependents, nil
}

func (e *EnqueueRequestsForServiceBinding) isClusterWide() bool {
	return len(e.WatchNamespaces) == 1 && e.WatchNamespaces[0] == ""
}
