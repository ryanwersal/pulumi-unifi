---
title: UniFi
meta_desc: Manage a Ubiquiti UniFi Dream Machine's Network and Protect applications with Pulumi.
layout: package
---

The UniFi provider for [Pulumi](https://www.pulumi.com) manages a Ubiquiti
**UniFi Dream Machine** — its **Network** and **Protect** applications — as
code. It is a **native** provider built with
[`pulumi-go-provider`](https://github.com/pulumi/pulumi-go-provider): it talks
directly to the **local UniFi controller API** and is **not** a Terraform
bridge.

## Example

{{< chooser language "typescript" >}}
{{% choosable language typescript %}}

```typescript
import * as unifi from "@ryanwersal/pulumi-unifi";

const lab = new unifi.network.Vlan("lab", {
    name: "lab",
    purpose: "corporate",
    vlan: 20,
    subnet: "192.168.20.1/24",
    dhcpEnabled: true,
    dhcpStart: "192.168.20.10",
    dhcpStop: "192.168.20.254",
});

new unifi.network.Wlan("lab-wifi", {
    name: "lab",
    networkId: lab.networkId,
    passphrase: cfg.requireSecret("wifiPassphrase"),
});
```

{{% /choosable %}}
{{< /chooser >}}

## Resources

| Token | What | Lifecycle |
|---|---|---|
| `unifi:network:Vlan` | A network / VLAN (DHCP, DNS, IGMP, IPv6, WAN, mDNS, …) | full CRUD + import |
| `unifi:network:Wlan` | A wireless network (SSID) | full CRUD + import |
| `unifi:network:Device` | Settings of an **existing** switch / AP / gateway / PDU | adoption: Create/Update merge, Read, no-op Delete |
| `unifi:network:PortProfile` | A reusable switch-port profile | full CRUD + import |
| `unifi:network:PortForward` | A port-forwarding rule | full CRUD + import |
| `unifi:network:FirewallGroup` | A firewall address / port group | full CRUD + import |
| `unifi:network:FirewallRule` | A classic per-ruleset firewall rule | full CRUD + import |
| `unifi:network:FirewallZonePolicy` | A zone-based firewall policy | full CRUD + import |
| `unifi:network:StaticRoute` | A static route | full CRUD + import |
| `unifi:network:User` | A known client (fixed IP, group, block, local DNS) | full CRUD + import |
| `unifi:network:UserGroup` | A bandwidth-limit user group | full CRUD + import |
| `unifi:network:DnsRecord` | A controller-local DNS record | full CRUD + import |
| `unifi:protect:Camera` | Settings of an **existing** Protect camera | adoption: Create/Update patch, Read, no-op Delete |
| `unifi:protect:AlarmAutomation` | An Alarm Manager rule (conditions → webhook actions) | full CRUD + import (private API) |

`Device` and `Camera` follow an **adoption model**: physical hardware is adopted
on the controller, not created via the API, so these manage settings on a device
you have already adopted (keyed by `mac` / `cameraId`) and a Pulumi delete only
releases it from management.

See [Installation & Configuration](installation-configuration) to get started.
