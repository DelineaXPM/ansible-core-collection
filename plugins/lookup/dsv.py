# -*- coding: utf-8 -*-
# Copyright: (c) 2023, Delinea <https://delinea.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)
from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = r"""
name: dsv
author:
    - Delinea (!UNKNOWN) (https://delinea.com/)
short_description: Get secrets from Delinea DevOps Secrets Vault
version_added: 1.0.0
description:
    - Uses the Delinea DevOps Secrets Vault Python SDK to get Secrets from a
      DSV I(tenant) using a I(client_id) and I(client_secret).
requirements:
    - python-dsv-sdk - https://pypi.org/project/python-dsv-sdk/
options:
    _terms:
        description: The path to the secret, e.g. C(/staging/servers/web1).
        required: true
    data_key:
        description: Specific field in secret data to return. If empty then
            entire secret object is returned. If defined, but not found then
            an error is returned.
        env:
            - name: DSV_DATA_KEY
        ini:
            - section: dsv_lookup
              key: data_key
        required: false
    tenant:
        description: The first format parameter in the default I(url_template).
        env:
            - name: DSV_TENANT
        ini:
            - section: dsv_lookup
              key: tenant
        required: true
    tld:
        default: com
        description: The top-level domain of the tenant; the second format
            parameter in the default I(url_template).
        env:
            - name: DSV_TLD
        ini:
            - section: dsv_lookup
              key: tld
        required: false
    client_id:
        description: The client_id with which to request the Access Grant.
        env:
            - name: DSV_CLIENT_ID
        ini:
            - section: dsv_lookup
              key: client_id
        required: true
    client_secret:
        description: The client secret associated with the specific I(client_id).
        env:
            - name: DSV_CLIENT_SECRET
        ini:
            - section: dsv_lookup
              key: client_secret
        required: true
    url_template:
        default: https://{}.secretsvaultcloud.{}/v1
        description: The path to prepend to the base URL to form a valid REST
            API request.
        env:
            - name: DSV_URL_TEMPLATE
        ini:
            - section: dsv_lookup
              key: url_template
        required: false
"""

RETURN = r"""
_list:
    description:
        - One or more JSON responses to C(GET /secrets/{path}).
        - See U(https://dsv.secretsvaultcloud.com/api/index.html#operation/getSecret).
    type: list
    elements: dict
"""

EXAMPLES = r"""
- hosts: localhost
  vars:
      secret: "{{ lookup('delinea.core.dsv', '/test/secret') }}"
  tasks:
      - ansible.builtin.debug:
          msg: 'the password is {{ secret["data"]["password"] }}'
"""

import json  # noqa: E402

from ansible.errors import AnsibleError, AnsibleOptionsError  # noqa: E402
from ansible.plugins.lookup import LookupBase  # noqa: E402
from ansible.utils.display import Display  # noqa: E402

sdk_is_missing = False

try:
    from thycotic.secrets.vault import SecretsVault, SecretsVaultError  # noqa: E402
except ImportError:
    sdk_is_missing = True


display = Display()


class LookupModule(LookupBase):
    @staticmethod
    def Client(vault_parameters):
        try:
            dsv_client = SecretsVault(**vault_parameters)
        except TypeError:
            raise AnsibleError(
                "python-dsv-sdk==0.0.1 must be installed to use this plugin"
            )
        return dsv_client

    def run(self, terms, variables, **kwargs):
        if sdk_is_missing:
            raise AnsibleError(
                "python-dsv-sdk==0.0.1 must be installed to use this plugin"
            )

        self.set_options(var_options=variables, direct=kwargs)

        dsv_client = LookupModule.Client(
            {
                "tenant": self.get_option("tenant"),
                "client_id": self.get_option("client_id"),
                "client_secret": self.get_option("client_secret"),
                "tld": self.get_option("tld"),
                "url_template": self.get_option("url_template"),
            }
        )

        data_key = self.get_option("data_key")

        result = []

        for term in terms:
            display.v("delinea.core.dsv: term: %s" % term)

            path = term.lstrip("[/:]")
            if path == "":
                raise AnsibleOptionsError("Invalid secret path: %s" % term)

            display.v("delinea.core.dsv: path: %s" % path)

            if data_key:
                dsv_secret = self._get_secret_data_key(dsv_client, path, data_key)
            else:
                dsv_secret = self._get_secret(dsv_client, path)

            result.append(dsv_secret)

        return result

    def _get_secret(self, dsv_client, path):
        try:
            dsv_secret = dsv_client.get_secret_json(path)
        except SecretsVaultError as error:
            raise AnsibleError("DSV lookup failure: %s" % error.message)
        return dsv_secret

    def _get_secret_data_key(self, dsv_client, path, data_key):
        try:
            response_body = dsv_client.get_secret(path)
        except SecretsVaultError as error:
            raise AnsibleError("DSV lookup failure: %s" % error.message)

        try:
            dsv_secret_data = response_body["data"][data_key]
        except KeyError:
            raise AnsibleOptionsError(
                "DSV lookup failure: cannot find data key in secret data"
            )

        if isinstance(dsv_secret_data, str):
            return dsv_secret_data

        return json.dumps(dsv_secret_data)
