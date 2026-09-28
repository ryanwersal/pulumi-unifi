# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to
follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once released.

## [Unreleased]

## [0.2.0] - 2026-09-27

- Add UniFi OS custom roles, local administrators, and integration API keys,
  including local-console authentication support.

- Add `Vlan.wan.dhcpv6PdSizeAuto` for explicit DHCPv6 prefix-size requests.
- Preserve live network fields outside the managed input delta on updates,
  including controller routing, firewall-zone, and WAN failover settings.
- Apply the `networkGroup=LAN` default only to non-WAN networks, so importing
  an existing WAN does not add an unrelated interface group.

## [0.1.0] - 2026-07-01

Initial release:

- Resources for UniFi Network (Vlan, Wlan, Device, PortProfile, PortForward,
  FirewallGroup, FirewallRule, FirewallZonePolicy, StaticRoute, User, UserGroup,
  DnsRecord) and Protect (Camera, AlarmAutomation).
- Closed value sets modeled as enums; controller defaults surfaced via
  `SetDefault`; identity fields force replacement.
- Idempotent deletes, drift-to-deleted reads, and contextual error messages.
- Hermetic lifecycle tests, a golden schema snapshot, and a shared CI gate.
