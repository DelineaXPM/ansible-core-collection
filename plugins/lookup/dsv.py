# -*- coding: utf-8 -*-
# Copyright: (c) 2022, Delinea <https://delinea.com>
# Copyright: (c) 2020, Adam Migus <adam@migus.org>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)
from __future__ import absolute_import, division, print_function
from ansible.errors import AnsibleError, AnsibleOptionsError
from ansible.utils.display import Display
from ansible.plugins.lookup import LookupBase


__metaclass__ = type

DOCUMENTATION = r"""
name: dsv
author:
    - Delinea (!UNKNOWN) (https://delinea.com/)
    - Adam Migus (@amigus) <adam@migus.org>
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


sdk_is_missing = False

try:
    from thycotic.secrets.vault import (
        SecretsVault,
        SecretsVaultError,
    )
except ImportError:
    sdk_is_missing = True


display = Display()


class LookupModule(LookupBase):
    @staticmethod
    def Client(vault_parameters):
        try:
            vault = SecretsVault(**vault_parameters)
            return vault
        except TypeError:
            raise AnsibleError(
                "python-dsv-sdk==0.0.1 must be installed to use this plugin"
            )

    def run(self, terms, variables, **kwargs):
        if sdk_is_missing:
            raise AnsibleError(
                "python-dsv-sdk==0.0.1 must be installed to use this plugin"
            )

        self.set_options(var_options=variables, direct=kwargs)

        vault = LookupModule.Client(
            {
                "tenant": self.get_option("tenant"),
                "client_id": self.get_option("client_id"),
                "client_secret": self.get_option("client_secret"),
                "tld": self.get_option("tld"),
                "url_template": self.get_option("url_template"),
            }
        )
        result = []

        for term in terms:
            display.debug("dsv_lookup term: %s" % term)

            path = term.lstrip("[/:]")
            if path == "":
                raise AnsibleOptionsError("Invalid secret path: %s" % term)

            display.vvv("DevOps Secrets Vault GET /secrets/%s" % path)

            try:
                result.append(vault.get_secret_json(path))
            except SecretsVaultError as error:
                raise AnsibleError(
                    "DevOps Secrets Vault lookup failure: %s" % error.message
                )
        return result