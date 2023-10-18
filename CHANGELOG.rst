==========================
Delinea.Core Release Notes
==========================

.. contents:: Topics


v1.1.0
======

Release Summary
---------------

New option to allow returning a specific key from the returned data. If it's defined, but not found it will error, otherwise by default the entire secret object will be returned.

Minor Changes
-------------

- dsv lookup plugin - add optional ``data_key`` parameter for filtering secret data.

v1.0.0
======

Release Summary
---------------

New plugin for getting secrets from Delinea DevOps Secrets Vault in Ansible.


New Plugins
-----------

Lookup
~~~~~~

- delinea.core.dsv - Get secrets from Delinea DevOps Secrets Vault
