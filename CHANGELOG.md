# Changelog

## [0.2.0](https://github.com/meysam81/vigil/compare/v0.1.0...v0.2.0) (2026-03-18)


### Features

* add migration for old CSP reports ([7a2f27a](https://github.com/meysam81/vigil/commit/7a2f27a6f81c2f30a1bb8760c90112518ef58310))
* add slack daily aggregated reports ([68f15fd](https://github.com/meysam81/vigil/commit/68f15fda756ccccb937579516575113c5a4fafbe))
* add structured and contextual logging throughout the app ([afc8b35](https://github.com/meysam81/vigil/commit/afc8b35f86609f61133afa59838eeab58d17831c))
* add support for bacth CSP reports ([91e884c](https://github.com/meysam81/vigil/commit/91e884c9049036424353f285f788f184fcd3e9f7))


### Bug Fixes

* **CI:** add internal to dockerignore ([df869e9](https://github.com/meysam81/vigil/commit/df869e9b817f3768201a23edc5df226d03578e5d))
* **CI:** make link-checker happy ([3ea34f8](https://github.com/meysam81/vigil/commit/3ea34f814e461eae76e2c2b2c5f13fe2ce4a89da))
* **CI:** update outdated jobs ([a33a217](https://github.com/meysam81/vigil/commit/a33a21799a966f3eb0e23db978f6a75d96d7eca7))
* polish up and add request max bytes reader ([04086fa](https://github.com/meysam81/vigil/commit/04086fa74a3fb960a95be55e876f9b1ca87ea01d))
* rename the project to single-word ([ffdb1fd](https://github.com/meysam81/vigil/commit/ffdb1fdf15110d3aa622016f446619c5d31b560b))

## 0.1.0 (2025-08-25)


### Features

* add csp violation endpoint ([0cde52f](https://github.com/meysam81/vigil/commit/0cde52f807e24b695d3dae28341b0bec7acc06c8))
* add index healthcheck and retry after on 429 status ([7731002](https://github.com/meysam81/vigil/commit/77310028dda70cc2b32db7bf866cd97f44e8e255))
* add rate-limiting ([219897f](https://github.com/meysam81/vigil/commit/219897fa99513f71a1f2085bc7394428af9bd86b))
* add readme ([0fed417](https://github.com/meysam81/vigil/commit/0fed417b090774942138c7c9c6e5841a49eaab01))
* add tests and support for unicode in content type ([5a6cb4c](https://github.com/meysam81/vigil/commit/5a6cb4c074eb7a9d7422e9e421d53ff5b215d136))
* **docs:** make sure legacy and modern language is clearly stated ([0eb45c8](https://github.com/meysam81/vigil/commit/0eb45c8a08612be44c03d2e6b5e8dc9111e391bb))
* support both report-uri legacy and modern report-to ([dd58f57](https://github.com/meysam81/vigil/commit/dd58f57a766f4b2a9d2a4c54ca31ee306c32be68))


### Bug Fixes

* accept csp report content type ([988cf10](https://github.com/meysam81/vigil/commit/988cf10fc829fc8f5cde99a9488d9057d3956576))
* **deps:** update module github.com/meysam81/x to v1.11.2 ([#3](https://github.com/meysam81/vigil/issues/3)) ([67f4c80](https://github.com/meysam81/vigil/commit/67f4c8011b437ec5ba0a9a622af17ef21ee2dd65))
* **deps:** update module github.com/redis/go-redis/v9 to v9.12.1 ([#5](https://github.com/meysam81/vigil/issues/5)) ([a8a63d1](https://github.com/meysam81/vigil/commit/a8a63d13e5c0e61d659d7d68391048192168e097))
* remove obsolete config ([6403c66](https://github.com/meysam81/vigil/commit/6403c66ea938fb23471ba594b9ac4dffcb8e31f1))
* remove obsolete default from config ([6ab17ca](https://github.com/meysam81/vigil/commit/6ab17ca286bb894a800a71d1dcea5b3f8cd286b6))
