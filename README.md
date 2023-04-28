# bc-wallet-common-lib-healthcheck

## Description

Library for manage healthcheck config prepare http-server with healthcheck handlers

Library contains:
* Config structs and implementation of http-server for 3 healthcheck probes:
  * Startup
  * Rediness 
  * Liveness

Each healthcheck probe it is http-server with uniq config and listen address/port.

## Authors

## Contributors

* Author and maintainer - [@gudron (Alex V Kotelnikov)](https://github.com/gudron)

## Licence

**bc-wallet-common-lib-healthcheck** has a proprietary license.

Switched to proprietary license from MIT - [CHANGELOG.MD - v0.0.2](./CHANGELOG.md)