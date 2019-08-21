[![Build Status](https://travis-ci.com/appsody/appsody-operator.svg?branch=master)](https://travis-ci.com/appsody/appsody-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/appsody/appsody-operator)](https://goreportcard.com/report/github.com/appsody/appsody-operator)

# Appsody Application Operator

The Appsody Application Operator has been designed to deploy applications created by [Appsody Application Stacks](https://appsody.dev/) into [OKD](https://www.okd.io/) clusters.  The goal of this project is to iteratively grow the operator's set of day-2 capabilities.

Check out our [demo](demo/README.md) page!

If there's a certain functionality you would like to see or a bug you would like to report, please use our [issues tab](https://github.com/appsody/appsody-operator/issues) to get in contact with us.

## Operator Installation

You can install the Appsody Application Operator via a single `kubectl` command or assisted by the [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager).

### Direct installation

Use the instructions for one of the [releases](https://github.com/appsody/appsody-operator/tree/master/deploy/releases) to directly install this Operator into a Kubernetes cluster.

### OLM-assisted installation

*Note:* OLM is labelled as a tech preview for OKD / OpenShift 3.11.  

* install OLM as described in [here](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/install/install.md#installing-olm).
* install the operator via this command:

## Overview

The architecture of the Appsody Application Operator follows the basic controller pattern:  the Operator container with the controller is deployed into a Pod and listens for incoming resources with `Kind: AppsodyApplication`.

![Operator Architecture](architecture.png)

## User guide

For more information on how to use the `AppsodyApplication` operator, see [user guide](doc/user-guide.md).
