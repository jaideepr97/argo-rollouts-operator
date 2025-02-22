
# Development

## Prerequisites

 * Go 1.19+
 * Operator SDK `v1.28.0`
 * Bash or equivalent
 * Docker

### Run and test operator locally

To run the operator locally on your machine (outside a container), invoke the following make target:

``` bash
make install run
```

This will install the CRDs into your cluster, then run the operator on your machine.

To run the unit tests, invoke the following make target:

``` bash
make test
```

### Build operator

Use the following make target to build the operator. A container image wil be created locally.

The name of the image is specified by the `IMG` variable defined in the `Makefile`.

``` bash
make docker-build
```

### Push the development container image.

Override the name of the image to push by specifying the `IMG` variable.

``` bash
make docker-push IMG=quay.io/my-org/argorollout-operator:latest
```

### Generate the bundle artifacts.

Override the name of the development image by specifying the `IMG` variable.

``` bash
rm -fr bundle/
make bundle IMG=quay.io/my-org/argorollout-operator:latest
```

### Build and push the development bundle image.

Override the name of the bundle image by specifying the `BUNDLE_IMG` variable.

``` bash
make bundle-build BUNDLE_IMG=quay.io/my-org/argorollout-operator-bundle:latest
make bundle-push BUNDLE_IMG=quay.io/my-org/argorollout-operator-bundle:latest
```

### Build and push the development catalog image.

Override the name of the catalog image by specifying the `CATALOG_IMG` variable.
Specify the bundle image to include using the `BUNDLE_IMG` variable

``` bash
make catalog-build BUNDLE_IMG=quay.io/my-org/argorollout-operator-bundle:latest CATALOG_IMG=quay.io/my-org/argorollout-operator-index:latest
make catalog-push CATALOG_IMG=quay.io/my-org/argorollout-operator-index:latest
```

### Build and Verify ArgoRollout Operator Docs

#### Prerequisites

- `Python3`

Create a Python Virtual Environment. This is not mandatory, you can continue without creating a Virtual Environment as well.

```bash
python3 -m venv doc
```

Get into the virtual environment, if you have created one using the above step.

```bash
source doc/bin/activate
```

Install the required Python libraries

```bash
pip3 install mkdocs
pip3 install mkdocs-material
```

Start the `mkdocs` server locally to verify the UI changes.

```bash
mkdocs serve
```

### Run the e2e tests.

[WIP]