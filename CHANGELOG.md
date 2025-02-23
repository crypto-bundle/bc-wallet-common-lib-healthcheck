# Change Log

## [v0.0.7] - 03.10.2024
### Added
* Added linters checks:
  * Added linters config
  * Fixed all liners issues
* Changed error formatter interface - need for new version of lib-errors

## [v0.0.5, v0.0.6] - 28.09.2024
### Added
* Added support of new version of lib-logger library
* Added support of new ver of lib-errors library
* Added slog.Logger
### Changed
* Removed zap.Logger dependency
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