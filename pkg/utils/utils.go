package utils

import (
	"github.com/application-stacks/runtime-component-operator/pkg/common"

	oputils "github.com/application-stacks/runtime-component-operator/pkg/utils"
)

// GetOpenShiftAnnotations add any additional annotations that are provided by the CLI tool
func GetOpenShiftAnnotations(ba common.BaseComponent) map[string]string {
	annos := map[string]string{}

	for key, val := range ba.GetLabels() {
		if key == "image.opencontainers.org/source" {
			annos["app.openshift.io/vcs-uri"] = val
		}
	}

	return oputils.MergeMaps(annos, oputils.GetConnectToAnnotation(ba))
}
