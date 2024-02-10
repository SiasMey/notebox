# Changelog

## [v1.3.0](https://github.com/SiasMey/notebox/compare/v1.2.1...v1.3.0) (2024-02-10)

### Features

* dont build locally
([dcf9d02](https://github.com/SiasMey/notebox/commit/dcf9d024f6d805f0a94770ede2059b13ccdce3d2))

### [v1.2.1](https://github.com/SiasMey/notebox/compare/v1.2.0...v1.2.1) (2024-02-10)

#### Fixes

* ok, use the clone in gh
([0d708cb](https://github.com/SiasMey/notebox/commit/0d708cb1f25ff4f11f4ee494873d65f914b037b2))
* unused time import
([57bbd73](https://github.com/SiasMey/notebox/commit/57bbd7323a72d38125404a3b1f8eb5e0c35e8751))
* just use locally cloned files, you have them anyway
([00d24c9](https://github.com/SiasMey/notebox/commit/00d24c9e004f78b502e00a99c5942436667688b9))

## [v1.2.0](https://github.com/SiasMey/notebox/compare/v1.1.6...v1.2.0) (2024-02-10)

### Features

* add linting
([2f18659](https://github.com/SiasMey/notebox/commit/2f18659d86bc694d4a23a808d7f82ce21a501729))
* add checking formating of files
([b47110f](https://github.com/SiasMey/notebox/commit/b47110fb93bcf46a32c8b82c1b35870ab467b7f0))
* limit changelog for local and remote
([bd91f65](https://github.com/SiasMey/notebox/commit/bd91f653ec167ec31c1bc614423319a9fe97665f))
* local prints changelog, but no commit and no tag
([0216c5e](https://github.com/SiasMey/notebox/commit/0216c5ec518f01380021017f0d576526bd0fa9f1))
* dev versions for local execution
([fd3af46](https://github.com/SiasMey/notebox/commit/fd3af46596e3ae1e601139adb2052a58c25039c9))
* if running local, use local folder as src
([ed37f21](https://github.com/SiasMey/notebox/commit/ed37f21f6da9b05b30b5bfd9ca44c47352c419be))

### Fixes

* call to publish
([d5ce9bf](https://github.com/SiasMey/notebox/commit/d5ce9bf2748e462a8e60cf4fd7f7d17bfbc0f330))

### [v1.1.6](https://github.com/SiasMey/notebox/compare/v1.1.5...v1.1.6) (2024-02-09)

#### Fixes

* publishing on remote
([247497f](https://github.com/SiasMey/notebox/commit/247497f5e611cf37400fdb4511dff3d4a0a40d04))

### [v1.1.5](https://github.com/SiasMey/notebox/compare/v1.1.4...v1.1.5) (2024-02-09)

#### Fixes

* so many flips
([2765aab](https://github.com/SiasMey/notebox/commit/2765aabbab2e573a72b7c95fc7a050adaa9b84c3))
* flip remote check
([b7b0545](https://github.com/SiasMey/notebox/commit/b7b0545f70cc8d9da013ab21e3340c4b10a7fc12))
* versioning result
([cdde881](https://github.com/SiasMey/notebox/commit/cdde881c30597a16a7783ea9c4d1ec6cd5825f30))
* flip check to make it exit when there is no bump
([a461301](https://github.com/SiasMey/notebox/commit/a4613014151b358e2adff5976ad3f6b6f8b01aa8))

### [v1.1.4](https://github.com/SiasMey/notebox/compare/v1.1.3...v1.1.4) (2024-02-09)

#### Fixes

* still needs checkout, just not much
([d1f2d67](https://github.com/SiasMey/notebox/commit/d1f2d67ccc2814562d863c989a0b215a382b0208))
* clean up deprecated files
([2d92d49](https://github.com/SiasMey/notebox/commit/2d92d49e60cd5c5bb64f13b0277172937bbe9591))

### [v1.1.3](https://github.com/SiasMey/notebox/compare/v1.1.2...v1.1.3) (2024-02-09)

#### Fixes

* no, you need that cache bust
([9382625](https://github.com/SiasMey/notebox/commit/9382625b51b0b37746f5e51a1ad89bb4dc50bb2b))
* remove cache bust and version.txt
([1914eab](https://github.com/SiasMey/notebox/commit/1914eab8f683ab323169924432ea4a8a4cd3eecf))

### [v1.1.2](https://github.com/SiasMey/notebox/compare/v1.1.1...v1.1.2) (2024-02-09)

#### Fixes

* update container as you go
([b01ef3d](https://github.com/SiasMey/notebox/commit/b01ef3d42dfc88ee752d56452e6b7be2413b8735))
* containers are copied, not pointered
([5732417](https://github.com/SiasMey/notebox/commit/57324172e29cfdaebed3d74b29ad93e31e4da07a))
* print changelog
([0d1eedd](https://github.com/SiasMey/notebox/commit/0d1eeddff892274e8ee77ee2c4e8ee084a4249e1))
* make version and changelog temp files
([28cf9d4](https://github.com/SiasMey/notebox/commit/28cf9d4dc740a44cedfee4612b17f65922c331c4))

### [v1.1.1](https://github.com/SiasMey/notebox/compare/v1.1.0...v1.1.1) (2024-02-09)

#### Fixes

* update ci to release changelog
([fbf2a2d](https://github.com/SiasMey/notebox/commit/fbf2a2df021e859a5d2d1a3930c1da25f9b86328))

## [v1.1.0](https://github.com/SiasMey/notebox/compare/v1.0.0...v1.1.0) (2024-02-09)

### Features

* add changelog generation
([c19bfb1](https://github.com/SiasMey/notebox/commit/c19bfb146fe23efcb4871204cae317461eba93ab))

### Fixes

* use gitconfig for secrets
([48d07dc](https://github.com/SiasMey/notebox/commit/48d07dc58527f7efec6a77a0033876871879e793))
* make it output version and changelog files
([bb390f6](https://github.com/SiasMey/notebox/commit/bb390f6e4e9dde69c848e679347634796b30d48b))
* change message
([fd94a1c](https://github.com/SiasMey/notebox/commit/fd94a1c7c4fcd24b073624854d14e428a62f6ad7))

## [v1.0.0](https://github.com/SiasMey/notebox/compare/v0.1.0...v1.0.0) (2024-02-03)

## v0.1.0 (2024-02-09)

### Features

* add changelog generation
([c19bfb1](https://github.com/SiasMey/notebox/commit/c19bfb146fe23efcb4871204cae317461eba93ab))

### Fixes

* use gitconfig for secrets
([48d07dc](https://github.com/SiasMey/notebox/commit/48d07dc58527f7efec6a77a0033876871879e793))
* make it output version and changelog files
([bb390f6](https://github.com/SiasMey/notebox/commit/bb390f6e4e9dde69c848e679347634796b30d48b))
* change message
([fd94a1c](https://github.com/SiasMey/notebox/commit/fd94a1c7c4fcd24b073624854d14e428a62f6ad7))
