# Development Guide

- [Development Guide](#development-guide)
  - [Prerequisites](#prerequisites)
  - [Get the code](#get-the-code)
  - [Dev environment](#dev-environment)
  - [Development \& Testing](#development--testing)
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

```shell
mkdir -p ~/ansible_collections/delinea/core
```

```shell
git clone git@github.com:DelineaXPM/ansible-core-collection.git ~/ansible_collections/delinea/core
```

```shell
cd ~/ansible_collections/delinea/core
```

## Dev environment

Configure pre-commit:

```shell
python3 -m pip install pre-commit --user && pre-commit install && pre-commit
```

We use [Mage][mage] build tool to automate most of the tasks related to development.
Mage is installed by aqua manager.
Run `mage init` to setup.

> This might take up to 8 mins the first time as it sets up virtual environments for every version listed.

To list all available mage targets run `mage -l`.

## Development & Testing

For local development, you should activate one of the target environments that were automatically installed when running `mage job:setup` (or `mage init`).
The terminal will have printing the source command for you.

For example: `source .cache/stable-2.13/bin/activate` and you should have the venv activated now to develop with.

## Test

Run local dockerized tests with: `mage job:setup venv:testsanity`.

## Release

Follow [this link][delinea-core-galaxy] to open the `delinea.core` collection in [Ansible Galaxy][galaxy] hub.

When creating a new release start with writing a release summary.
Run `mage ansible:changelog` to generate a new release_summary fragment interactively.

To update the version in `galaxy.yml` run `mage ansible:bump "patch"` and update installation instructions in [README.md][readme.md].

To run the entire release lifecycle:

```shell
mage job:release
```

Run `mage ansible:doctor` to validate all the requirements for publishing are installed (you'll want to activate the virtual env first).

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
