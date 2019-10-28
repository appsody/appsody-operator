<!--
This file includes chronologically ordered list of notable changes visible to end users for each version of the Appsody Operator. Keep a summary of the change and link to the pull request.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
-->

# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]

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
- Automatically configure the `AppsodyApplication`'s Kubernetes resources to allow automatic creation of an application definition by [kAppNav](https://kappnav.io/), Kubernetes Application Navigator. ([#128](https://github.com/appsody/appsody-operator/issues/128),[#135](https://github.com/appsody/appsody-operator/issues/135))

### Changed

- Removed default values for all stacks from `appsody-operator-defaults` ConfigMap since Appsody Stacks would already have defaults. ([#104](https://github.com/appsody/appsody-operator/issues/104))
- Made the `stack` field not required to specify in `AppsodyApplication`. ([#125](https://github.com/appsody/appsody-operator/issues/125))

### Fixed

- **Breaking change:** When deploying to Knative, Knative route is only accessible if `expose` is set to `true`. This is to make `expose` behaviour consistent with non-Knative deployment so users need to explicitly declare to make their application accessible externally. ([#122](https://github.com/appsody/appsody-operator/issues/122))
- Fixed an issue for basic storage scenario (i.e. specifying `storage.mountPath` and `storage.size`) that if `storage.size` is not specified or is not valid, handle it gracefully. ([#130](https://github.com/appsody/appsody-operator/issues/130))

## [0.1.0]

The initial release of the Appsody Operator 🎉🥳

[Unreleased]: https://github.com/appsody/appsody-operator/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/appsody/appsody-operator/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/appsody/appsody-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/appsody/appsody-operator/releases/tag/v0.1.0