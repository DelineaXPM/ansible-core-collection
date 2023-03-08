# Development Guide

- [Prerequisites](#prerequisites)
- [Get the code](#get-the-code)
- [Dev environment](#dev-environment)
- [Test](#test)
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

```bash
mkdir -p ~/ansible_collections/delinea/core
```

```bash
git clone git@github.com:DelineaXPM/ansible-core-collection.git ~/ansible_collections/delinea/core
```

```bash
cd ~/ansible_collections/delinea/core
```

## Dev environment

Configure pre-commit:

```bash
python3 -m pip install pre-commit --user && pre-commit install && pre-commit
```

We use [Mage][mage] build tool to automate most of the tasks related to development.
Mage is installed by aqua manager. Run `mage init` to setup.

To list all available mage targets run `mage -l`.

## Test

Run local dockerized tests with: `mage job:setup venv:testsanity`.

## Release

Follow [this link][delinea-core-galaxy] to open the `delinea.core` collection in [Ansible Galaxy][galaxy] hub.

When creating a new release start with writing a release summary.
Run `mage ansible:changelog` to generate a new release_summary fragment interactively.

Update version in `galaxy.yml` and update installation instructions in [README.md][readme.md].

Build the collection:

```bash
ansible-galaxy collection build
```

As a result a new archive will be generated (e.g. `delinea-core-1.0.0.tar.gz`) and it should be published.

[developing-collections]: https://docs.ansible.com/ansible/latest/dev_guide/developing_collections.html
[get-python]: https://www.python.org/downloads/
[get-docker]: https://docs.docker.com/get-docker/
[get-aqua]: https://aquaproj.github.io/docs/reference/install
[get-trunk]: https://docs.trunk.io/docs/install
[mage]: https://magefile.org/
[delinea-core-galaxy]: https://galaxy.ansible.com/delinea/core
[galaxy]: https://galaxy.ansible.com/
[readme.md]: README.md
