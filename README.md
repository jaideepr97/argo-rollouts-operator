# Argo Rollouts Operator

**Project Status: BETA**

Not all planned features are completed. The API, spec, status and other user facing objects may change.

## Summary

A Kubernetes operator for managing [Argo Rollouts](https://github.com/argoproj/argo-rollouts/). This operator provides an easy way to install, upgrade and manage the lifecycle of Argo Rollouts.

This operator is built using `operator-sdk`, version - `v1.28.0`.

## Documentation

The complete documentation for the operator can be found [here](https://argo-rollouts-operator.readthedocs.io/en/latest/).

## What exactly the operator does ?

When Installed, this operator creates a Custom Resource Definition called ArgoRollout. 

Operator will then wait for the users to deploy the corresponding Custom Resource to create the rollout controller and other resources according to the provided spec.

Read more about the Argo Rollout CRD specification here.

## Where to start ?

We have a getting started [guide](docs/usage/getting_started.md) which provides information on how to start using the operator.

### Development

Instructions to run the operator locally or create your own version of the operator Image are provided in the development [section](docs/developer-guide/developer_guide.md) of the docs.

### Contributions

[WIP]






