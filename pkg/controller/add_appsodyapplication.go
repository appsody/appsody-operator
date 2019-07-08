package controller

import (
	"github.com/appsody-operator/pkg/controller/appsodyapplication"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, appsodyapplication.Add)
}
