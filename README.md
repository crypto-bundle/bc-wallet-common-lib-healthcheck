# bc-wallet-common-lib-healthcheck

## Description

Library for manage healthcheck config prepare http-server with healthcheck handlers

Library contains:
* Config structs and implementation of http-server for 3 healthcheck probes:
  * Startup
  * Rediness 
  * Liveness

Each healthcheck probe it is http-server with uniq config and listen address/port.

## Contributors

* Author and maintainer - [@gudron (Alex V Kotelnikov)](https://github.com/gudron)

## Licence

**bc-wallet-common-lib-healthcheck** is licensed under the [MIT](./LICENSE) License.