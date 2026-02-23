# Changelog

## [0.3.2](https://github.com/corazawaf/libinjection-go/compare/v0.3.1...v0.3.2) (2026-02-23)


### Performance Improvements

* optimize sqli detection with safe, zero-alloc patterns ([#97](https://github.com/corazawaf/libinjection-go/issues/97)) ([9f17d96](https://github.com/corazawaf/libinjection-go/commit/9f17d9614b0f585ecfa71f61cdee9646be983c63))
* optimize xss detection with zero-alloc patterns ([#98](https://github.com/corazawaf/libinjection-go/issues/98)) ([28df6ea](https://github.com/corazawaf/libinjection-go/commit/28df6ea908b9af91b3b7671592ace1381503c46d))

## [0.3.1](https://github.com/corazawaf/libinjection-go/compare/v0.3.0...v0.3.1) (2026-02-23)


### Bug Fixes

* align test file parser with C testdriver behavior ([#95](https://github.com/corazawaf/libinjection-go/issues/95)) ([9196e75](https://github.com/corazawaf/libinjection-go/commit/9196e75fa33ac495fcd4ce019a28a7acf51e1720))
* correct off-by-one in XML comment detection in XSS checker ([#93](https://github.com/corazawaf/libinjection-go/issues/93)) ([1978215](https://github.com/corazawaf/libinjection-go/commit/1978215e84d9ff2564618b7935b2d7c917be06f3))
* correct SVG tag detection typo and use prefix matching in isBlackTag ([#92](https://github.com/corazawaf/libinjection-go/issues/92)) ([75a7f79](https://github.com/corazawaf/libinjection-go/commit/75a7f79456bd788bb3394ee5a9e1991f942478ee))
* implement XSS test driver for test-xss-* files ([#94](https://github.com/corazawaf/libinjection-go/issues/94)) ([f05bbb8](https://github.com/corazawaf/libinjection-go/commit/f05bbb83ceafcc8d27ded4a7481d4bcc9a7227dd))
* use HasPrefix instead of Contains in htmlEncodeStartsWith ([#91](https://github.com/corazawaf/libinjection-go/issues/91)) ([51891ca](https://github.com/corazawaf/libinjection-go/commit/51891cabdad8a507d1fe2e927d76e25111d8b380)), closes [#46](https://github.com/corazawaf/libinjection-go/issues/46)

## [0.2.4](https://github.com/corazawaf/libinjection-go/compare/v0.2.3...v0.2.4) (2026-02-14)


### Bug Fixes

* scientific notation MySQL bypass ([69e28f8](https://github.com/corazawaf/libinjection-go/commit/69e28f853b8e86e5faea5d194fa0b5cfe90f3853))
