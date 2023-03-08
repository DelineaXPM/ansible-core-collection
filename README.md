# Ansible Collection - delinea.core

![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/delineaxpm/ansible-core-collection?style=for-the-badge)

Ansible core collection for Delinea DevOps Secrets Vault.

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

You can also include it in a `requirements.yml` file and install it with
`ansible-galaxy collection install -r requirements.yml`, using the format:

```yaml
---
collections:
  - name: delinea.core
```

Note that if you install the collection from Ansible Galaxy, it will not be upgraded
automatically when you upgrade the `ansible` package. To upgrade the collection to
the latest available version, run the following command:

```bash
ansible-galaxy collection install delinea.core --upgrade
```

You can also install a specific version of the collection, for example, if you need
to downgrade when something is broken in the latest version (please report an issue
in this repository). Use the following syntax to install version `1.0.0`:

```bash
ansible-galaxy collection install delinea.core:==1.0.0
```

See [Ansible Using collections](https://docs.ansible.com/ansible/devel/user_guide/collections_using.html) for more details.

## Contributing

Read the [Development Guide][developer.md] to learn about our build, test, and release processes.

[developer.md]: https://github.com/DelineaXPM/ansible-core-collection/blob/main/DEVELOPER.md
