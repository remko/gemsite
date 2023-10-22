---
title: "An Age plugin for Apple's Secure Enclave"
date: 2023-07-14
featured: true
---

For my day-to-day encryption needs, I'm a big fan of
[age](https://github.com/FiloSottile/age). Age is a simple, modern and secure
file encryption tool, and serves as a better replacement for
tools such as GnuPG. You can even use it as [the backend for managing your passwords](https://github.com/FiloSottile/passage "passage").

For extra convenience and security, I wanted to be able to use my MacBook's Secure
Enclave (controlled by Touch ID) to encrypt files, so I created an age plugin for this: [`age-plugin-se`](https://github.com/remko/age-plugin-se).

## Usage

You can create public/private key pairs that are bound to the Secure Enclave of your machine by calling the plugin directly:

```
$ age-plugin-se keygen --access-control=any-biometry
# created: 2023-07-08T19:00:19Z
# access control: any biometry
# public key: age1se1qfn44rsw0xvmez3pky46nghmnd5up0jpj97nd39zptlh83a0nja6skde3ak
AGE-PLUGIN-SE-1QJPQZLE3SGQHKVYP75X6KYPZPQ3N44RSW0XVMEZ3QYUNTXXQ7UVQTPSPKY6TYQSZDNVLMZYCYSRQRWP
```

The public key can then be used to encrypt files using `age`:

```
$ tar cvz ~/data | age -r age1se1qfn44rsw0xvmez3pky46nghmnd5up0jpj97nd39zptlh83a0nja6skde3ak
```

Note that encryption can be done on any machine, even machines without Secure
Enclaves, or even machines running Linux or Windows.

When decrypting the encrypted file, the key will now require Touch ID to use
the Secure Enclave to decrypt it:

```
$ age --decrypt -i key.txt data.tar.gz.age > data.tar.gz
```

![Touch ID prompt](/blog/age-plugin-se/screenshot-biometry.png)

For each generated key, you have the choice of different combinations of
requiring biometry (e.g. Touch ID), passcodes, or both.

## Implementation

The plugin is implemented entirely in Swift, and uses Apple's
[CryptoKit](https://developer.apple.com/documentation/cryptokit) framework for
all crypto operations. On non-macOS platforms, the plugin uses [Swift
Crypto](https://github.com/apple/swift-crypto), a cross-platform open source
implementation of a subset of CryptoKit.

Other than CryptoKit (or Swift Crypto) and the core frameworks, the plugin has
no other dependencies.

## Build & Tests

To avoid depending on Xcode, the plugin uses the [Swift Package
Manager](https://www.swift.org/package-manager/) for its builds. This allows
you to compile it from the command-line, on any platform that has a Swift
distribution.

You can also run XCTest unit tests from the command-line using SwiftPM, but
there is no annotated coverage information. I therefore created a 
[Swift script](https://github.com/remko/age-plugin-se/blob/main/Scripts/ProcessCoverage.swift "`ProcessCoverage.swift`") that takes the raw coverage data from SwiftPM, and outputs [source code annotated with coverage data](https://remko.github.io/age-plugin-se/ci/coverage.html "`age-plugin-se` Coverage Report") (together with an SVG icon for the GitHub project page).

## Format

The plugin uses the `piv-p256` recipient stanza in encrypted files. This is the
same stanza type used by the [age YubiKey plugin](https://github.com/str4d/age-plugin-yubikey). This recipient stanza is [currently being standardized](https://github.com/C2SP/C2SP/pull/31 "age-piv-p256 recipient stanza"). 

Although the plugin is complete and tested, since the recipient stanza is still being standardized, I'm holding off releasing a 1.0 version of the plugin until the dust settles. 

In the meantime, feedback on the plugin is welcome!
