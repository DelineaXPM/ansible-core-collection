# ansible-core-collection

![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/delineaxpm/ansible-core-collection?style=for-the-badge)

> ***Warning***
>
> UNTIL THIS NOTICE IS REMOVED THIS IS BEING ACTIVELY DEVELOPED AND IS NOT A FINAL PRODUCTION RELEASE.

Ansible core collection for Delinea DevOps Secret Vault.

## Development

- Run `python3 -m pip install pre-commit --user && pre-commit install && pre-commit`.
- For task automation, install Mage: [Install Mage](https://magefile.org/)
    - For a quick install if you have Go 1.17+ installed run `go install github.com/magefile/mage@latest`.
    - Run `mage init` to setup.
    - Run `mage -l` to list tasks.

## Included content

### Lookup plugins

| Name             | Description                                        |
| ---------------- | -------------------------------------------------- |
| delinea.core.dsv | Look up secrets from Delinea DevOps Secrets Vault. |

## Using this collection

### Installing the Collection from Ansible Galaxy

Before using this collection, you need to install it with the Ansible Galaxy command-line tool:

```bash
ansible-galaxy collection install delinea.core
```

You can also include it in a `requirements.yml` file and install it with `ansible-galaxy collection install -r requirements.yml`, using the format:

```yaml
---
collections:
  - name: delinea.core
```

Note that if you install the collection from Ansible Galaxy, it will not be upgraded automatically when you upgrade the `ansible` package. To upgrade the collection to the latest available version, run the following command:

```bash
ansible-galaxy collection install delinea.core --upgrade
```

You can also install a specific version of the collection, for example, if you need to downgrade when something is broken in the latest version (please report an issue in this repository). Use the following syntax to install version `1.0.0`:

```bash
ansible-galaxy collection install delinea.core:==1.0.0
```

See [Ansible Using collections](https://docs.ansible.com/ansible/devel/user_guide/collections_using.html) for more details.
