# -*- coding: utf-8 -*-
# (c) 2022, Delinea <https://delinea.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

from unittest import TestCase

try:
    from unittest.mock import patch
except ImportError:
    from mock import patch

from ansible.errors import AnsibleOptionsError
from ansible.plugins.loader import lookup_loader
from ansible_collections.delinea.core.plugins.lookup import dsv


class TestLookupModule(TestCase):
    def setUp(self):
        dsv.sdk_is_missing = False
        self.lookup = lookup_loader.get("delinea.core.dsv")

    @patch("ansible_collections.delinea.core.plugins.lookup.dsv.LookupModule.Client")
    def test_run_removed_leading_slash(self, mock_client):
        instance = mock_client.return_value
        instance.get_secret_json.return_value = '{"data": "secret-val"}'

        result = self.lookup.run(
            ["/secret/path"],
            [],
            **{"tenant": "ten", "client_id": "cid", "client_secret": "csecret"}
        )

        instance.get_secret_json.assert_called_once_with("secret/path")
        assert len(result) == 1
        assert result[0] == '{"data": "secret-val"}'

    @patch("ansible_collections.delinea.core.plugins.lookup.dsv.LookupModule.Client")
    def test_run_removed_leading_colon(self, mock_client):
        instance = mock_client.return_value
        instance.get_secret_json.return_value = '{"data": "secret-val"}'

        result = self.lookup.run(
            [":secret:path"],
            [],
            **{"tenant": "ten", "client_id": "cid", "client_secret": "csecret"}
        )

        instance.get_secret_json.assert_called_once_with("secret:path")
        assert len(result) == 1
        assert result[0] == '{"data": "secret-val"}'

    @patch("ansible_collections.delinea.core.plugins.lookup.dsv.LookupModule.Client")
    def test_run_empty_path(self, mock_client):
        with self.assertRaises(AnsibleOptionsError) as exc:
            self.lookup.run(
                [""],
                [],
                **{"tenant": "ten", "client_id": "cid", "client_secret": "csecret"}
            )

        assert str(exc.exception) == "Invalid secret path: "

    @patch("ansible_collections.delinea.core.plugins.lookup.dsv.LookupModule.Client")
    def test_run_slash_path(self, mock_client):
        with self.assertRaises(AnsibleOptionsError) as exc:
            self.lookup.run(
                ["/"],
                [],
                **{"tenant": "ten", "client_id": "cid", "client_secret": "csecret"}
            )

        assert str(exc.exception) == "Invalid secret path: /"

    @patch("ansible_collections.delinea.core.plugins.lookup.dsv.LookupModule.Client")
    def test_run_colon_path(self, mock_client):
        with self.assertRaises(AnsibleOptionsError) as exc:
            self.lookup.run(
                [":"],
                [],
                **{"tenant": "ten", "client_id": "cid", "client_secret": "csecret"}
            )

        assert str(exc.exception) == "Invalid secret path: :"
