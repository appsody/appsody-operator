<!--
This file includes chronologically ordered list of notable changes visible to end users for each version of the Appsody Operator. Keep a summary of the change and link to the pull request.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
-->

# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Support for watching multiple namespaces by setting `WATCH_NAMESPACE` to a comma-separated list of namespaces. ([#114](https://github.com/appsody/appsody-operator/issues/114))
- Allow users to add new labels or override default labels for resources created by the operator via setting labels on `AppsodyApplication` CRs. ([#118](https://github.com/appsody/appsody-operator/issues/118))
- Support for automatically creating and configuring `ServiceMonitor` resource for integration with Prometheus Operator

### Changed

- Removed default values for all stacks from `appsody-operator-defaults` ConfigMap since Appsody Stacks would already have defaults. ([#104](https://github.com/appsody/appsody-operator/issues/104))

### Fixed

- **Breaking change:** When deploying to Knative, Knative route is only accessible if `expose` is set to `true`. This is to make `expose` behaviour consistent with non-Knative deployment so users need to explicitly declare to make their application accessible externally. ([#122](https://github.com/appsody/appsody-operator/issues/122))

## [0.1.0]

The initial release of the Appsody Operator 🎉🥳

[Unreleased]: https://github.com/appsody/appsody-operator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/appsody/appsody-operator/releases/tag/v0.1.0