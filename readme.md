[![Build Status](https://travis-ci.com/appsody/appsody-operator.svg?branch=master)](https://travis-ci.com/appsody/appsody-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/appsody/appsody-operator)](https://goreportcard.com/report/github.com/appsody/appsody-operator)

## Development of Appsody as a standalone project has ended, but the core technologies of Appsody have been merged with odo to create odo 2.0! See our [blog post](https://appsody.dev/blogs/DevelopmentEnded) for more details!

# Appsody Application Operator

The Appsody Application Operator can be used to deploy applications created by [Appsody Application Stacks](https://appsody.dev/) into [OKD](https://www.okd.io/) or [OpenShift](https://www.openshift.com/) clusters.

Check out our [demo](demo/README.md) page!

If there's a certain functionality you would like to see or a bug you would like to report, please use our [issues tab](https://github.com/appsody/appsody-operator/issues) to get in contact with us.

## Operator Installation

You can install the Appsody Application Operator directly via `kubectl` commands or assisted by the [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager).

Use the instructions for one of the [releases](deploy/releases) to directly install this Operator into a Kubernetes cluster.

## Overview

The architecture of the Appsody Application Operator follows the basic controller pattern:  the Operator container with the controller is deployed into a Pod and listens for incoming resources with `Kind: AppsodyApplication`.

![Operator Architecture](architecture.png)

## Documentation

For information on how to use the `AppsodyApplication` operator, see the [documentation](doc/).

## Contributing

We welcome all contributions to the Appsody Application Operator project. Please see our [Contributing guidelines](CONTRIBUTING.md)
