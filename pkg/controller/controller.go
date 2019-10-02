package controller

import (
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	if err := routev1.AddToScheme(m.GetScheme()); err != nil {
		return err
	}

	if err := servingv1alpha1.AddToScheme(m.GetScheme()); err != nil {
		return err
	}

	if err := prometheusv1.AddToScheme(m.GetScheme()); err != nil {
		return err
	}

	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
