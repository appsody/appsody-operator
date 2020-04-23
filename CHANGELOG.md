<!--
This file includes chronologically ordered list of notable changes visible to end users for each version of the Appsody Operator. Keep a summary of the change and link to the pull request.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
-->

# Changelog

All notable changes to this project will be documented in this file.

## [0.5.0]

### Added

- Added Ingress (vanilla) support ([#259](https://github.com/appsody/appsody-operator/pull/259), [#79](https://github.com/application-stacks/runtime-component-operator/pull/79))
- Added support for external service bindings ([#259](https://github.com/appsody/appsody-operator/pull/259), [#76](https://github.com/application-stacks/runtime-component-operator/pull/76))
- Added additional service ports support ([#259](https://github.com/appsody/appsody-operator/pull/259), [#80](https://github.com/application-stacks/runtime-component-operator/pull/80))
- Added support to specify NodePort on service ([#259](https://github.com/appsody/appsody-operator/pull/259), [#60](https://github.com/application-stacks/runtime-component-operator/pull/60))

## [0.4.1]

### Fixed

- Auto-scaling (HPA) not working as expected ([#72](https://github.com/application-stacks/runtime-component-operator/pull/72))
- Operator crashes on some cluster due to optional CRDs (Knative Service, ServiceMonitor) not being present ([#254](https://github.com/appsody/appsody-operator/pull/254))
- Update the predicates for watching StatefulSet and Deployment sub-resource to check for generation to minimize number of reconciles ([#254](https://github.com/appsody/appsody-operator/pull/254))

## [0.4.0]

### Added

- Added support for integration with OpenShift's Certificate Management. ([#214](https://github.com/appsody/appsody-operator/pull/214))
- Added support for referencing images in image streams. ([#218](https://github.com/appsody/appsody-operator/pull/218))
- Added support to specify application name to group related resources. ([#237](https://github.com/appsody/appsody-operator/pull/237))
- Added support for sidecar containers. ([#237](https://github.com/appsody/appsody-operator/pull/237))
- Added optional targetPort field to CRD ([#181](https://github.com/appsody/appsody-operator/issues/181))
- Added support for naming service port. ([#237](https://github.com/appsody/appsody-operator/pull/237))
- Added must-gather scripts for troubleshooting. ([#209](https://github.com/appsody/appsody-operator/pull/209))
- Added the shortnames `app` and `apps` to the AppsodyApplication CRD. ([#198](https://github.com/appsody/appsody-operator/issues/198))
- Added OpenShift specific annotations ([#54](https://github.com/application-stacks/runtime-component-operator/pull/54))
- Set port name for Knative service if specified ([#55](https://github.com/application-stacks/runtime-component-operator/pull/55))

### Changed

- Changed the match label of the ServiceMonitor created by operator from `app.appsody.dev/monitor` to `monitor.appsody.dev/enabled`
- **Breaking change:** When `service.consumes[].namespace` is not specified, injected name of environment variable follows `<SERVICE-NAME>_<KEY>` format and binding information are mounted at `<mountPath>/<service_name>`. ([#27](https://github.com/application-stacks/runtime-component-operator/pull/27) and [#46](https://github.com/application-stacks/runtime-component-operator/pull/46))

## [0.3.0]

### Added

- Added support for basic service binding. ([#187](https://github.com/appsody/appsody-operator/issues/187))
- Added support for init containers array. ([#188](https://github.com/appsody/appsody-operator/issues/188))

### Changed

- Changed the default labels and annotations to more closely align with the
  generic operator and OpenShift guidelines ([#233](https://github.com/appsody/appsody-operator/issues/233))
- Changed the label corresponding to `metadata.name` to `app.kubernetes.io/instance` and made former label `app.kubernetes.io/name` user configurable. ([#179](https://github.com/appsody/appsody-operator/issues/179))

## [0.2.2]

### Added

- Allow users to add new annotations or override default annotations for resources created by the operator via setting annotations on `AppsodyApplication` CRs. ([#177](https://github.com/appsody/appsody-operator/issues/177))

## [0.2.1]

### Added

- Added documentation on how to do canary testing with the standard `Route` resource. ([#143](https://github.com/appsody/appsody-operator/issues/143))

### Changed

- Changed the label corresponding to the Appsody Stack information from `app.appsody.dev/stack` to `stack.appsody.dev/id`. ([#169](https://github.com/appsody/appsody-operator/issues/169))

## [0.2.0]

### Added

- Support for watching multiple namespaces by setting `WATCH_NAMESPACE` to a comma-separated list of namespaces. ([#114](https://github.com/appsody/appsody-operator/issues/114))
- Allow users to add new labels or override default labels for resources created by the operator via setting labels on `AppsodyApplication` CRs. ([#118](https://github.com/appsody/appsody-operator/issues/118))
- Support for automatically creating and configuring `ServiceMonitor` resource for integration with Prometheus Operator. ([#125](https://github.com/appsody/appsody-operator/issues/125))
- Automatically configure the `AppsodyApplication`'s Kubernetes resources to allow automatic creation of an application definition by [kAppNav](https://kappnav.io/), Kubernetes Application Navigator. ([#128](https://github.com/appsody/appsody-operator/issues/128) and [#135](https://github.com/appsody/appsody-operator/issues/135))

### Changed

- Removed default values for all stacks from `appsody-operator-defaults` ConfigMap since Appsody Stacks would already have defaults. ([#104](https://github.com/appsody/appsody-operator/issues/104))
- Made the `stack` field not required to specify in `AppsodyApplication`. ([#125](https://github.com/appsody/appsody-operator/issues/125))

### Fixed

- **Breaking change:** When deploying to Knative, Knative route is only accessible if `expose` is set to `true`. This is to make `expose` behaviour consistent with non-Knative deployment so users need to explicitly declare to make their application accessible externally. ([#122](https://github.com/appsody/appsody-operator/issues/122))
- Fixed an issue for basic storage scenario (i.e. specifying `storage.mountPath` and `storage.size`) that if `storage.size` is not specified or is not valid, handle it gracefully. ([#130](https://github.com/appsody/appsody-operator/issues/130))

## [0.1.0]

The initial release of the Appsody Operator 🎉🥳

[Unreleased]: https://github.com/appsody/appsody-operator/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/appsody/appsody-operator/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/appsody/appsody-operator/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/appsody/appsody-operator/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/appsody/appsody-operator/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/appsody/appsody-operator/compare/0.2.1...v0.2.2
[0.2.1]: https://github.com/appsody/appsody-operator/compare/v0.2.0...0.2.1
[0.2.0]: https://github.com/appsody/appsody-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/appsody/appsody-operator/releases/tag/v0.1.0
