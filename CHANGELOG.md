# Change Log

## [v0.0.2] - 09.02.2023 10:38 MSK

### Addded
* Separated config by probe type. Added readme.md file

### Changed
* Refactored healthcheck service - moved to new repository - https://github.com/crypto-bundle/bc-wallet-common-lib-healthcheck
* Swithched-back to MIT license

## [v0.0.3] - 13.04.2024
### Changed
* Refactored healthcheck service:
  * Probe http-server now is internal package entity
  * Changed probe unit dependency requirements
  * Added new field to probe config structs
  * Changed shutdown flow