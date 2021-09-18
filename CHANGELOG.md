# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v1.0.2 - 2021-09-18
## Fixed
* Replaced usage of ```time.Tick``` with ```time.NewTicker``` to prevent leaking of the underlying ticker.

## v1.0.2 - 2021-09-07
## Fixed
* Fixed unhandled error.

## v1.0.1 - 2021-08-21
### Added
* Added Go 1.17 support.
* Added a changelog.
* Added a code of conduct.

## v1.0.0 - 2021-06-22
### Changed
* **BREAKING**: Updated to v2 of the AWS SDK.

## v0.3.6 - 2021-03-20
### Changed
* Updated dependencies.

## v0.3.5 - 2021-02-20
### Added
* Added Go 1.16 support.

## v0.3.4 - 2021-02-06
### Changed
* Updated dependencies.

## v0.3.3 - 2021-02-06
### Changed
* Updated GitHub CI action.

## v0.3.2 - 2021-02-06
### Fixed
* Fixed several race condition issues with the ```CloudWatchLogger```.
* Fixed several issues with logs exceeding CloudWatch's limits.

## v0.3.1 - 2020-12-21
### Changed
* Updated dependencies.

## v0.3.0 - 2020-12-31
### Added
* Added Go 1.15 support.

### Changed
* Improved ```CloudWatchLogger``` tests.

## v0.2.0 - 2020-06-20
### Changed
* **BREAKING**: Renamed package from ```logger``` to ```log```.

## v0.1.3 - 2020-06-16
### Changed
* Removed capitalization of error messages.

## v0.1.2 - 2020-06-16
### Changed
* Updated dependencies.

## v0.1.1 - 2020-06-11
### Changed
* Updated dependencies.

## v0.1.0 - 2020-06-11
### Added
* Added ```Logger``` interface.
* Added ```StandardLogger``` struct.
* Added ```CloudWatchLogger``` struct.
