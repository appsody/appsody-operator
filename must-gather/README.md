# Appsody Operator must-gather

Appsody Operator must-gather is a tool built on top of OpenShift must-gather that expands its capabilities to gather information about the Appsody Operator.

Usage
oc adm must-gather --image=docker.io/appsody/application-operator:daily-must-gather
Note: must-gather flag is a new feature added to oc v4.x. If you are using an older version of oc, you can get a new version of the CLI from here. You can use oc v4.x against 3.11 cluster.

The command above will create a local directory with a dump of the Appsody Operator collection state. Note that this command will only get data related to the Appsody Operator Collection of the OpenShift cluster. 

In order to get data about other parts of the cluster (not specific to Appsody Operator) you should run just oc adm must-gather (without passing a custom image). Run oc adm must-gather -h to see more options.