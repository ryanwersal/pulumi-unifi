# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to
follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once released.

## [Unreleased]

The provider is pre-1.0 and unreleased; the schema may still change. Notable
work to date:

- Resources for UniFi Network (Vlan, Wlan, Device, PortProfile, PortForward,
  FirewallGroup, FirewallRule, FirewallZonePolicy, StaticRoute, User, UserGroup,
  DnsRecord) and Protect (Camera, AlarmAutomation).
- Closed value sets modeled as enums; controller defaults surfaced via
  `SetDefault`; identity fields force replacement.
- Idempotent deletes, drift-to-deleted reads, and contextual error messages.
- Hermetic lifecycle tests, a golden schema snapshot, and a shared CI gate.
