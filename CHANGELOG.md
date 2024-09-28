# Change Log

## [v0.0.5] - 28.09.2024
### Added
* Added support of new version of lib-logger library
### Changed
* Changed MIT License to NON-AI MIT
* Added License banner to all *.go files

## [v0.0.4] - 16.04.2024
### Changed
* Bump golang version 1.19 -> 1.22

## [v0.0.3] - 13.04.2024
### Changed
* Refactored healthcheck service:
  * Probe http-server now is internal package entity
  * Changed probe unit dependency requirements
  * Added new field to probe config structs
  * Changed shutdown flow

## [v0.0.2] - 09.02.2023 10:38 MSK
### Added
* Separated config by probe type. Added readme.md file
### Changed
* Refactored healthcheck service - moved to new repository - https://github.com/crypto-bundle/bc-wallet-common-lib-healthcheck
* Switched-back to MIT license