# Changelog

## [1.25.5](https://github.com/awesome-foundation/cfnrain/compare/v1.25.4...v1.25.5) (2026-11-06)


### Bug Fixes

* **build:** remove the AuthorizerId string AWS::NoValue from the webapp ([fac1b9b](https://github.com/awesome-foundation/cfnrain/commit/fac1b9ba1a51a1085899ac2d2a808df3e54179a3))
* **build:** use the s3:ReplicateTags action in replication policies ([fac1b9b](https://github.com/awesome-foundation/cfnrain/commit/fac1b9ba1a51a1085899ac2d2a808df3e54179a3))
* **build:** use the s3:ReplicateTags action in replication policies ([#100](https://github.com/awesome-foundation/cfnrain/issues/100)) ([4c2875a](https://github.com/awesome-foundation/cfnrain/commit/4c2875a9fc456ae38bbea3c373230039b0d442c1))
* **deploy:** report all unknown parameters and silence ignored ones ([#74](https://github.com/awesome-foundation/cfnrain/issues/74)) ([94d2ad3](https://github.com/awesome-foundation/cfnrain/commit/94d2ad339ff23849c3b9e0a0c2d39d225d45de2e)), closes [#68](https://github.com/awesome-foundation/cfnrain/issues/68)
* expand all short-form intrinsic tags and keep their source form ([#80](https://github.com/awesome-foundation/cfnrain/issues/80)) ([a91be98](https://github.com/awesome-foundation/cfnrain/commit/a91be9843795e5c81cfbbdbf5748faf615521086))
* **fmt:** keep comments in templates with CRLF line endings ([#75](https://github.com/awesome-foundation/cfnrain/issues/75)) ([c2023bc](https://github.com/awesome-foundation/cfnrain/commit/c2023bc3d1ec48c09254416f65e018838abfa2ed)), closes [#64](https://github.com/awesome-foundation/cfnrain/issues/64)
* **pkg:** find the !Rain::S3 Run script with an absolute template path ([#61](https://github.com/awesome-foundation/cfnrain/issues/61)) ([8390718](https://github.com/awesome-foundation/cfnrain/commit/8390718770793fb632dc44a98aef0bd4c7f4abe2)), closes [#60](https://github.com/awesome-foundation/cfnrain/issues/60)
* **pkg:** honor an absolute Run path and upload again after Run ([#88](https://github.com/awesome-foundation/cfnrain/issues/88)) ([637ab3f](https://github.com/awesome-foundation/cfnrain/commit/637ab3f30db2c0531f48f4c4eb39298b9eb1dbbd)), closes [#60](https://github.com/awesome-foundation/cfnrain/issues/60)
* **rm:** handle missing termination protection ([#98](https://github.com/awesome-foundation/cfnrain/issues/98)) ([cde0f2f](https://github.com/awesome-foundation/cfnrain/commit/cde0f2f78e901b588b2ebdb8a53927a82b51473e))
* **rm:** stop the panic on a stack in REVIEW_IN_PROGRESS ([#73](https://github.com/awesome-foundation/cfnrain/issues/73)) ([2df7070](https://github.com/awesome-foundation/cfnrain/commit/2df7070ed8ce6ad4b2b2a70f95a48c2d0b8d6023))

## [1.25.4](https://github.com/awesome-foundation/cfnrain/compare/v1.25.3...v1.25.4) (2026-10-27)


### Bug Fixes

* **pkg:** reject zip entries outside the destination ([89e08ef](https://github.com/awesome-foundation/cfnrain/commit/89e08ef34f382aa4ee3eb0ddaa5adc9b5e96065c))

## [1.25.3](https://github.com/awesome-foundation/cfnrain/compare/v1.25.2...v1.25.3) (2026-10-27)


### Bug Fixes

* **deps:** bump pkl-go, go-runewidth and testify ([#27](https://github.com/awesome-foundation/cfnrain/issues/27)) ([d560313](https://github.com/awesome-foundation/cfnrain/commit/d5603133f2fd9fc0b54507a6aa536924332e41f2))

## [1.25.2](https://github.com/awesome-foundation/cfnrain/compare/v1.25.1...v1.25.2) (2026-10-26)


### Bug Fixes

* **deps:** update Go dependencies ([#18](https://github.com/awesome-foundation/cfnrain/issues/18)) ([3c50f7c](https://github.com/awesome-foundation/cfnrain/commit/3c50f7c9a0d992323846baf0225bebf5680aae3c))

## [1.25.1](https://github.com/awesome-foundation/cfnrain/compare/v1.25.0...v1.25.1) (2026-10-26)


### Continuous Integration

* publish releases as drafts first and build fewer targets at once ([#21](https://github.com/awesome-foundation/cfnrain/issues/21)) ([c97fb21](https://github.com/awesome-foundation/cfnrain/commit/c97fb21e7d4a07852c1b37fc1062df712a820db1))

## [1.25.0](https://github.com/awesome-foundation/cfnrain/compare/v1.24.4...v1.25.0) (2026-10-26)


### Features

* rename the project to cfnrain ([#2](https://github.com/awesome-foundation/cfnrain/issues/2)) ([2f41823](https://github.com/awesome-foundation/cfnrain/commit/2f41823dfa55aff8de965e4c725dba8097a32059))
* stop writing AWS usage metrics into packaged templates ([#3](https://github.com/awesome-foundation/cfnrain/issues/3)) ([c0f97f3](https://github.com/awesome-foundation/cfnrain/commit/c0f97f387f06ce439034c97c762b8d25f8af89fa))
