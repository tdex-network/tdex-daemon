# Macaroons

This package makes the [lnd/macaroons](https://github.com/lightningnetwork/lnd/tree/42099ef5e1e0f927290531fe2bff18406141567b/macaroons) package a standalone go module.

Commit: [42099ef5](https://github.com/lightningnetwork/lnd/tree/42099ef5e1e0f927290531fe2bff18406141567b).

### CHANGELOG

* Filename is provided as argument to `NewService()` instead of being hardcoded into a string variable.