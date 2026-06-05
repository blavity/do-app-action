# Changelog

## [1.1.1](https://github.com/blavity/do-app-action/compare/v1.1.0...v1.1.1) (2026-06-05)


### Bug Fixes

* **ci:** add workflow_dispatch to release.yml for manual retrigger ([#35](https://github.com/blavity/do-app-action/issues/35)) ([09ba1b8](https://github.com/blavity/do-app-action/commit/09ba1b8c517daf1021de7804c07b0a8e31c6d4d0))

## [1.1.0](https://github.com/blavity/do-app-action/compare/v1.0.3...v1.1.0) (2026-06-05)


### Features

* publish pre-built Docker image to GHCR for ARC k8s container mode compatibility ([#33](https://github.com/blavity/do-app-action/issues/33)) ([77461fb](https://github.com/blavity/do-app-action/commit/77461fbede43e610a03e52427e96019976af7b76))


### Miscellaneous

* **deps:** bump actions/checkout from 6.0.2 to 6.0.3 ([#32](https://github.com/blavity/do-app-action/issues/32)) ([f5df248](https://github.com/blavity/do-app-action/commit/f5df2480088a0ec4744aa085b827ab27665238e6))
* **deps:** bump github.com/digitalocean/godo ([#20](https://github.com/blavity/do-app-action/issues/20)) ([370624d](https://github.com/blavity/do-app-action/commit/370624d1ce1f4312d93bee87fdb8efb7b4c80b9d))
* **deps:** bump golangci/golangci-lint-action from 9.2.0 to 9.2.1 ([#31](https://github.com/blavity/do-app-action/issues/31)) ([bc0b87f](https://github.com/blavity/do-app-action/commit/bc0b87f5fc9e988850abd0ae4f7f7c64ae9c0e00))

## [1.0.3](https://github.com/blavity/do-app-action/compare/v1.0.2...v1.0.3) (2026-05-17)


### Miscellaneous

* **deps:** bump actions/upload-artifact from 7.0.0 to 7.0.1 ([#24](https://github.com/blavity/do-app-action/issues/24)) ([f9e799b](https://github.com/blavity/do-app-action/commit/f9e799b27c28a5b0c4068133aaf7f83439886126))
* **deps:** bump googleapis/release-please-action from 4.4.0 to 5.0.0 ([#26](https://github.com/blavity/do-app-action/issues/26)) ([a8c3574](https://github.com/blavity/do-app-action/commit/a8c3574cbdcd553aa97ab4a990f3a4b262755caf))
* **deps:** bump jdx/mise-action from 4.0.0 to 4.0.1 ([#21](https://github.com/blavity/do-app-action/issues/21)) ([9d992c7](https://github.com/blavity/do-app-action/commit/9d992c7c10cdcc9803ef4930d835fc4a338bdc71))

## [1.0.2](https://github.com/blavity/do-app-action/compare/v1.0.1...v1.0.2) (2026-03-14)


### Bug Fixes

* **ci:** use GITHUB_TOKEN for release-please instead of GitHub App ([#8](https://github.com/blavity/do-app-action/issues/8)) ([3c8b2ab](https://github.com/blavity/do-app-action/commit/3c8b2ab559855e39e511e04d4e0780e507840e5b))


### Miscellaneous

* **deps:** bump github.com/digitalocean/godo in the go-modules group ([#17](https://github.com/blavity/do-app-action/issues/17)) ([abbabdc](https://github.com/blavity/do-app-action/commit/abbabdcaa8233319f6f55991b2305c5dd266492a))
* **deps:** bump jdx/mise-action from 3.6.3 to 4.0.0 ([#16](https://github.com/blavity/do-app-action/issues/16)) ([c1b4815](https://github.com/blavity/do-app-action/commit/c1b4815996ef8f38ff2019aff035146e27ef2fb4))

## [1.0.1](https://github.com/blavity/do-app-action/compare/v1.0.0...v1.0.1) (2026-03-11)


### Bug Fixes

* **delete:** guard against nil resp on Delete error ([87ad57d](https://github.com/blavity/do-app-action/commit/87ad57d3d77f92d715ff8887949a6a1cfc6942a1))
* **deps:** update go-modules ([#5](https://github.com/blavity/do-app-action/issues/5)) ([38c6b49](https://github.com/blavity/do-app-action/commit/38c6b495243757c9ba748e59b0c0d8c415c8b7fc))
* **unarchive:** wait_timeout=0 now means poll indefinitely ([2c12951](https://github.com/blavity/do-app-action/commit/2c129518b933e8c5a722554393227fd0df501ceb))
* **utils:** InputAsBool must not clobber true defaults ([049620c](https://github.com/blavity/do-app-action/commit/049620c2cdc6aa53646f6c62ea48ed0916d6c1b7))


### Miscellaneous

* **deps:** update github-actions-versions ([#6](https://github.com/blavity/do-app-action/issues/6)) ([c287701](https://github.com/blavity/do-app-action/commit/c287701404d362cf4098e51be6fb8d840f50c8ba))
* **deps:** update golang docker tag to v1.26 ([#4](https://github.com/blavity/do-app-action/issues/4)) ([2de7206](https://github.com/blavity/do-app-action/commit/2de7206926ecbf1017e6298af856bee36d35cf83))
* **renovate:** add shared preset config ([df714c7](https://github.com/blavity/do-app-action/commit/df714c7d6fa16caca382d4e9a8c91077a400bf12))
* **renovate:** add shared preset config ([270dbdf](https://github.com/blavity/do-app-action/commit/270dbdf7a62aa21c13dbaee449f3dde8bec39ca8))
