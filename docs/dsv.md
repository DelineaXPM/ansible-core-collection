# dsv -- Get secrets from Delinea DevOps Secrets Vault

## Synopsis

Uses the Delinea DevOps Secrets Vault Python SDK to get Secrets from a DSV _tenant_ using a _client_id_ and _client_secret_.

## Requirements

The below requirements are needed on the host that executes this module.

- python-dsv-sdk - https://pypi.org/project/python-dsv-sdk/

## Parameters

\_terms (True, any, None)
The path to the secret, e.g. `/staging/servers/web1`.

data_key (False, any, None)
Specific field in secret data to return. If empty then entire secret object is returned. If defined, but not found then an error is returned.

tenant (True, any, None)
The first format parameter in the default _url_template_.

tld (False, any, com)
The top-level domain of the tenant; the second format parameter in the default _url_template_.

client_id (True, any, None)
The client_id with which to request the Access Grant.

client*secret (True, any, None)
The client secret associated with the specific \_client_id*.

url_template (False, any, https://{}.secretsvaultcloud.{}/v1)
The path to prepend to the base URL to form a valid REST API request.

## Examples

```yaml
- hosts: localhost
  vars:
    secret: "{{ lookup('delinea.core.dsv', '/test/secret') }}"
  tasks:
    - ansible.builtin.debug:
        msg: 'the password is {{ secret["data"]["password"] }}'
```

## Return Values

\_list (, list, )
One or more JSON responses to `GET /secrets/{path}`.
See https://dsv.secretsvaultcloud.com/api/index.html#operation/getSecret.

## Status

## Authors

- Delinea (!UNKNOWN) (https://delinea.com/)
