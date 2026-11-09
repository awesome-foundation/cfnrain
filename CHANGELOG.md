# Changelog

## [1.25.5](https://github.com/awesome-foundation/cfnrain/compare/v1.25.4...v1.25.5) (2026-11-09)


### Bug Fixes

* **build:** honor only-cache and reject conflicting cache flags ([#136](https://github.com/awesome-foundation/cfnrain/issues/136)) ([c0fc080](https://github.com/awesome-foundation/cfnrain/commit/c0fc080172278a46a9210d8c2d5abd3886982b85))
* **build:** remove the AuthorizerId string AWS::NoValue from the webapp ([fac1b9b](https://github.com/awesome-foundation/cfnrain/commit/fac1b9ba1a51a1085899ac2d2a808df3e54179a3))
* **build:** use the s3:ReplicateTags action in replication policies ([fac1b9b](https://github.com/awesome-foundation/cfnrain/commit/fac1b9ba1a51a1085899ac2d2a808df3e54179a3))
* **build:** use the s3:ReplicateTags action in replication policies ([#100](https://github.com/awesome-foundation/cfnrain/issues/100)) ([4c2875a](https://github.com/awesome-foundation/cfnrain/commit/4c2875a9fc456ae38bbea3c373230039b0d442c1))
* **cfn:** handle empty stack responses and fast deletions ([#105](https://github.com/awesome-foundation/cfnrain/issues/105)) ([e7c5185](https://github.com/awesome-foundation/cfnrain/commit/e7c51857e7fe86401935d8f93ea06f622462c8bb))
* **deploy:** confirm before deleting an existing stack ([#106](https://github.com/awesome-foundation/cfnrain/issues/106)) ([46d63c7](https://github.com/awesome-foundation/cfnrain/commit/46d63c714087c03f26607c65fb26311d4a9f11dc))
* **deploy:** handle stack lookup gaps ([#116](https://github.com/awesome-foundation/cfnrain/issues/116)) ([0da3440](https://github.com/awesome-foundation/cfnrain/commit/0da3440f5c602a916f02061d042b42daa9a34333))
* **deploy:** preserve failure when final lookup fails ([#130](https://github.com/awesome-foundation/cfnrain/issues/130)) ([cff264c](https://github.com/awesome-foundation/cfnrain/commit/cff264c99346fb500524f3b38ddb865f5aacc48f))
* **deploy:** report all unknown parameters and silence ignored ones ([#74](https://github.com/awesome-foundation/cfnrain/issues/74)) ([94d2ad3](https://github.com/awesome-foundation/cfnrain/commit/94d2ad339ff23849c3b9e0a0c2d39d225d45de2e)), closes [#68](https://github.com/awesome-foundation/cfnrain/issues/68)
* **deps:** update pflag to v1.1.0 and accept new help text ([#109](https://github.com/awesome-foundation/cfnrain/issues/109)) ([6a791c5](https://github.com/awesome-foundation/cfnrain/commit/6a791c59491e10a3304641110d773a998e5ba166))
* expand all short-form intrinsic tags and keep their source form ([#80](https://github.com/awesome-foundation/cfnrain/issues/80)) ([a91be98](https://github.com/awesome-foundation/cfnrain/commit/a91be9843795e5c81cfbbdbf5748faf615521086))
* **fmt:** keep comments in templates with CRLF line endings ([#75](https://github.com/awesome-foundation/cfnrain/issues/75)) ([c2023bc](https://github.com/awesome-foundation/cfnrain/commit/c2023bc3d1ec48c09254416f65e018838abfa2ed)), closes [#64](https://github.com/awesome-foundation/cfnrain/issues/64)
* **fmt:** preserve anchors on function arguments ([#113](https://github.com/awesome-foundation/cfnrain/issues/113)) ([d9cf4fb](https://github.com/awesome-foundation/cfnrain/commit/d9cf4fbaf6c194fa1893c9cc6072e46fcf3b5ba5))
* **fmt:** preserve comments on short-form function keys ([#112](https://github.com/awesome-foundation/cfnrain/issues/112)) ([8dfc62b](https://github.com/awesome-foundation/cfnrain/commit/8dfc62b055e3421e88ef58795334f3f99a4ce870))
* **fmt:** preserve YAML merge keys and reject JSON conversion ([#115](https://github.com/awesome-foundation/cfnrain/issues/115)) ([19287ca](https://github.com/awesome-foundation/cfnrain/commit/19287ca3214b1b041af97fc14d15f22231ea02b6))
* **merge:** return input parse errors without panicking ([#111](https://github.com/awesome-foundation/cfnrain/issues/111)) ([ca5439c](https://github.com/awesome-foundation/cfnrain/commit/ca5439cfb8577ba609b948d06fe673e600592567))
* **pkg:** align S3 option help descriptions ([#131](https://github.com/awesome-foundation/cfnrain/issues/131)) ([e47d24c](https://github.com/awesome-foundation/cfnrain/commit/e47d24c3e339cbc035924a620a1bd1c7700afa58))
* **pkg:** correct S3 options and module help ([#119](https://github.com/awesome-foundation/cfnrain/issues/119)) ([5d4b771](https://github.com/awesome-foundation/cfnrain/commit/5d4b77187f29e0c08bb30b31adb5e7e0a68d80de))
* **pkg:** find the !Rain::S3 Run script with an absolute template path ([#61](https://github.com/awesome-foundation/cfnrain/issues/61)) ([8390718](https://github.com/awesome-foundation/cfnrain/commit/8390718770793fb632dc44a98aef0bd4c7f4abe2)), closes [#60](https://github.com/awesome-foundation/cfnrain/issues/60)
* **pkg:** honor an absolute Run path and upload again after Run ([#88](https://github.com/awesome-foundation/cfnrain/issues/88)) ([637ab3f](https://github.com/awesome-foundation/cfnrain/commit/637ab3f30db2c0531f48f4c4eb39298b9eb1dbbd)), closes [#60](https://github.com/awesome-foundation/cfnrain/issues/60)
* **pkg:** preserve nested Sub constant strings ([#125](https://github.com/awesome-foundation/cfnrain/issues/125)) ([372e039](https://github.com/awesome-foundation/cfnrain/commit/372e03962d581c92abb6398a69c65ba26d490aa1))
* **pkg:** reject Sub constants in plain strings ([#132](https://github.com/awesome-foundation/cfnrain/issues/132)) ([9fc6ecf](https://github.com/awesome-foundation/cfnrain/commit/9fc6ecfef4acc3eeebbb5e30f05386f97034d6c8))
* **pkg:** report failed implicit uploads ([#114](https://github.com/awesome-foundation/cfnrain/issues/114)) ([4bf2570](https://github.com/awesome-foundation/cfnrain/commit/4bf25708446dea16379ef1924a4940eace91797c))
* **pkg:** report hard constant errors after deferred errors ([#133](https://github.com/awesome-foundation/cfnrain/issues/133)) ([652d7c1](https://github.com/awesome-foundation/cfnrain/commit/652d7c1c226295e139b045460794c0251679424c))
* **pkg:** resolve Fn::Select in CloudFormation argument order ([#123](https://github.com/awesome-foundation/cfnrain/issues/123)) ([41f117c](https://github.com/awesome-foundation/cfnrain/commit/41f117cd3649fc9b682921fb84dbb49d405a633c))
* **rm:** handle missing termination protection ([#98](https://github.com/awesome-foundation/cfnrain/issues/98)) ([cde0f2f](https://github.com/awesome-foundation/cfnrain/commit/cde0f2f78e901b588b2ebdb8a53927a82b51473e))
* **rm:** stop the panic on a stack in REVIEW_IN_PROGRESS ([#73](https://github.com/awesome-foundation/cfnrain/issues/73)) ([2df7070](https://github.com/awesome-foundation/cfnrain/commit/2df7070ed8ce6ad4b2b2a70f95a48c2d0b8d6023))
* **s3:** correct artifact bucket lifecycle rule ID ([#117](https://github.com/awesome-foundation/cfnrain/issues/117)) ([d8cdf7f](https://github.com/awesome-foundation/cfnrain/commit/d8cdf7fc424ef2b83edbe1c39aff69248d239df7))
* **stackset:** remove misleading optional template help ([#118](https://github.com/awesome-foundation/cfnrain/issues/118)) ([4b76c50](https://github.com/awesome-foundation/cfnrain/commit/4b76c506746a5c5774a3b34436b9042cedc4c328))
* **table:** isolate stdout capture in writer test ([#124](https://github.com/awesome-foundation/cfnrain/issues/124)) ([105a3a0](https://github.com/awesome-foundation/cfnrain/commit/105a3a083833f05b43bc50a4268aaed07df655dd))

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
