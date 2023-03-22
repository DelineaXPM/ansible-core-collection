# Development Guide

Table of contents:

- [Prerequisites](#prerequisites)
- [Get the code](#get-the-code)
- [Development](#development)
- [Release](#release)

Make sure to checkout [the official developer guide for developing collections][developing-collections].

## Prerequisites

- [Python][get-python] version 3.7 or higher
- [Docker][get-docker]
- [aqua][get-aqua]
- [Trunk][get-trunk]

## Get the code

Ansible requires that collections are stored in `{...}/ansible_collections/NAMESPACE/COLLECTION_NAME` path.
Therefore to be able to run tests, clone this repository to `{arbitrary path}/ansible_collections/delinea/core`.

Example with using home directory:

```shell
mkdir -p ~/ansible_collections/delinea/core
```

```shell
git clone git@github.com:DelineaXPM/ansible-core-collection.git ~/ansible_collections/delinea/core
```

```shell
cd ~/ansible_collections/delinea/core
```

## Development

Configure pre-commit:

```shell
python3 -m pip install pre-commit --user && pre-commit install && pre-commit
```

We use [Mage][mage] build tool to automate most of the tasks related to development.
Mage is installed by aqua manager.

Create a new virtual environment and install ansible:

```shell
mage init
```

Run unit and sanity tests:

```shell
mage test
```

To list all available mage targets run `mage -l`.

## Release

Follow [this link][delinea-core-galaxy] to open the `delinea.core` collection in [Ansible Galaxy][galaxy] hub.

1. Bump a new version. Available arguments are `"patch"`, `"minor"` and `"major"`:

   ```shell
   mage bump "patch"
   ```

2. Update installation instructions in [README.md][readme.md].

3. Write a release summary:

   ```shell
   mage changelog
   ```

4. Build the collection:

   ```shell
   mage build
   ```

   As a result a new archive will be generated (e.g. `delinea-core-1.0.0.tar.gz`) in the artifacts directory (`.artifacts/`).

5. Publish the collection:

   ```shell
   mage publish
   ```

Run `mage doctor` to validate all the requirements for publishing are installed.

[developing-collections]: https://docs.ansible.com/ansible/latest/dev_guide/developing_collections.html
[get-python]: https://www.python.org/downloads/
[get-docker]: https://docs.docker.com/get-docker/
[get-aqua]: https://aquaproj.github.io/docs/reference/install
[get-trunk]: https://docs.trunk.io/docs/install
[mage]: https://magefile.org/
[delinea-core-galaxy]: https://galaxy.ansible.com/delinea/core
[galaxy]: https://galaxy.ansible.com/
[readme.md]: README.md
