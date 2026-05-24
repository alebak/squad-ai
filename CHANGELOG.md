# Changelog

## [0.4.0](https://github.com/alebak/squad-ai/compare/squad-ai-v0.3.0...squad-ai-v0.4.0) (2026-05-24)


### Features

* add GoReleaser config, install script, and ldflags version support ([6a20578](https://github.com/alebak/squad-ai/commit/6a20578cbf951da841e951f540cc2b37aee34e3d))
* bootstrap Go project with squad version command ([1d1321b](https://github.com/alebak/squad-ai/commit/1d1321b6772218ec814f3b858e37e0c266107e3d))
* **cli:** add auto-update on startup ([7a43f4b](https://github.com/alebak/squad-ai/commit/7a43f4b38feebeb22a6c414a314a430de62850c9))
* **cli:** add install, list, add, remove, update, and info commands ([70305b7](https://github.com/alebak/squad-ai/commit/70305b7f39a05ce87d68817718f6d2649339ff29))
* **cli:** add squad reset command to clean config and cache ([295af26](https://github.com/alebak/squad-ai/commit/295af26edb4c29d8223083dd07768c5ebe0ad370))
* **cli:** add squad self-update command ([cb2be73](https://github.com/alebak/squad-ai/commit/cb2be738e9e7e5de61547a8d078b6f54b7232cfd))
* **config:** add user config with atomic load/save ([9850a6a](https://github.com/alebak/squad-ai/commit/9850a6a49dcbccba8c862272b51b2c1fa5d50449))
* **installer:** add agent detection and installation pipeline ([a2aa886](https://github.com/alebak/squad-ai/commit/a2aa886e54c45f01827c4242d34ab6588fb323b1))
* **installer:** add non_interactive flag to pipe yes for interactive installers ([8f908ec](https://github.com/alebak/squad-ai/commit/8f908ec6db8bdc15f1e13c7c618d5c5cdce18765))
* **registry:** add agent types, HTTP fetch, and local cache ([23be0e4](https://github.com/alebak/squad-ai/commit/23be0e4d88220ef094be28d0f0e5e0ed3b3e7104))
* **runtime:** add Node.js, Go, and Python version detection ([bba5dc0](https://github.com/alebak/squad-ai/commit/bba5dc0a2963a4632f3d4cab02056ec48f3265ae))
* **tui:** add interactive agent selection with Bubbletea ([a93a285](https://github.com/alebak/squad-ai/commit/a93a285cb272ca7e0b43c2b7d17305973e556e47))
* **tui:** add select all / deselect all toggle with 'a' key ([d45b2d2](https://github.com/alebak/squad-ai/commit/d45b2d282ab532aa57b55ed5acbf9e3465dd9ff2))
* **tui:** show all agents including installed in TUI selection ([3dc6711](https://github.com/alebak/squad-ai/commit/3dc6711683131b069ecf34ecb036a2d6b9a3c326))


### Bug Fixes

* **cli:** skip new agent notification on first run ([8f67d24](https://github.com/alebak/squad-ai/commit/8f67d24c7701764c002022dfc140f038d7415e44))
* display Squad AI product name in version output ([bd528f7](https://github.com/alebak/squad-ai/commit/bd528f78dc3c9167fe11a27ab9976b2c34b08927))
* feat commits should bump minor version pre-1.0 ([e9b0480](https://github.com/alebak/squad-ai/commit/e9b048019dbccf21f465b0256c71dca000a03615))
* **installer:** verify binary exists after successful install ([687e254](https://github.com/alebak/squad-ai/commit/687e254ffa513934bcd4e3b6df5000e315a89561))
* **install:** move tmpdir to global scope for EXIT trap ([f684a06](https://github.com/alebak/squad-ai/commit/f684a06db7f30b5d0e8a0091a25814c41d09281e))
* **registry:** add install commands and SHA-256 checksums for all agents ([35492c7](https://github.com/alebak/squad-ai/commit/35492c7b2a064a2a438848961a6e22a2bc855279))
* use clean tag format for release-please ([9e2294a](https://github.com/alebak/squad-ai/commit/9e2294accbf80ada9c66cca3c49ec9737eab9cb1))


### Documentation

* add project documentation ([4281cc2](https://github.com/alebak/squad-ai/commit/4281cc2a9e3b3dbed507544a63bc0cf9af0da616))
* rewrite README with bilingual content and usage examples ([c7d9a67](https://github.com/alebak/squad-ai/commit/c7d9a673ffcb15068cb38e07c1e8ca6fe81dec59))
* split README into English and Spanish, update author name ([8524856](https://github.com/alebak/squad-ai/commit/8524856fce77779bbb387eefed504cd0d3fc3c16))


### Code Refactoring

* **cli:** split long functions into helpers, remove dead code ([73fba2d](https://github.com/alebak/squad-ai/commit/73fba2db9dd29748cb241b9a8e302d0b5937ec00))
* replace manual agent installs with Squad AI in post-create.sh ([18bbe83](https://github.com/alebak/squad-ai/commit/18bbe83d36be01b20f71a86860dea0dbb16746c7))

## [0.3.0](https://github.com/alebak/squad-ai/compare/squad-ai-v0.2.0...squad-ai-v0.3.0) (2026-05-24)


### Features

* add GoReleaser config, install script, and ldflags version support ([6a20578](https://github.com/alebak/squad-ai/commit/6a20578cbf951da841e951f540cc2b37aee34e3d))
* bootstrap Go project with squad version command ([1d1321b](https://github.com/alebak/squad-ai/commit/1d1321b6772218ec814f3b858e37e0c266107e3d))
* **cli:** add install, list, add, remove, update, and info commands ([70305b7](https://github.com/alebak/squad-ai/commit/70305b7f39a05ce87d68817718f6d2649339ff29))
* **cli:** add squad reset command to clean config and cache ([295af26](https://github.com/alebak/squad-ai/commit/295af26edb4c29d8223083dd07768c5ebe0ad370))
* **cli:** add squad self-update command ([cb2be73](https://github.com/alebak/squad-ai/commit/cb2be738e9e7e5de61547a8d078b6f54b7232cfd))
* **config:** add user config with atomic load/save ([9850a6a](https://github.com/alebak/squad-ai/commit/9850a6a49dcbccba8c862272b51b2c1fa5d50449))
* **installer:** add agent detection and installation pipeline ([a2aa886](https://github.com/alebak/squad-ai/commit/a2aa886e54c45f01827c4242d34ab6588fb323b1))
* **installer:** add non_interactive flag to pipe yes for interactive installers ([8f908ec](https://github.com/alebak/squad-ai/commit/8f908ec6db8bdc15f1e13c7c618d5c5cdce18765))
* **registry:** add agent types, HTTP fetch, and local cache ([23be0e4](https://github.com/alebak/squad-ai/commit/23be0e4d88220ef094be28d0f0e5e0ed3b3e7104))
* **runtime:** add Node.js, Go, and Python version detection ([bba5dc0](https://github.com/alebak/squad-ai/commit/bba5dc0a2963a4632f3d4cab02056ec48f3265ae))
* **tui:** add interactive agent selection with Bubbletea ([a93a285](https://github.com/alebak/squad-ai/commit/a93a285cb272ca7e0b43c2b7d17305973e556e47))
* **tui:** add select all / deselect all toggle with 'a' key ([d45b2d2](https://github.com/alebak/squad-ai/commit/d45b2d282ab532aa57b55ed5acbf9e3465dd9ff2))
* **tui:** show all agents including installed in TUI selection ([3dc6711](https://github.com/alebak/squad-ai/commit/3dc6711683131b069ecf34ecb036a2d6b9a3c326))


### Bug Fixes

* **cli:** skip new agent notification on first run ([8f67d24](https://github.com/alebak/squad-ai/commit/8f67d24c7701764c002022dfc140f038d7415e44))
* display Squad AI product name in version output ([bd528f7](https://github.com/alebak/squad-ai/commit/bd528f78dc3c9167fe11a27ab9976b2c34b08927))
* feat commits should bump minor version pre-1.0 ([e9b0480](https://github.com/alebak/squad-ai/commit/e9b048019dbccf21f465b0256c71dca000a03615))
* **installer:** verify binary exists after successful install ([687e254](https://github.com/alebak/squad-ai/commit/687e254ffa513934bcd4e3b6df5000e315a89561))
* **install:** move tmpdir to global scope for EXIT trap ([f684a06](https://github.com/alebak/squad-ai/commit/f684a06db7f30b5d0e8a0091a25814c41d09281e))
* **registry:** add install commands and SHA-256 checksums for all agents ([35492c7](https://github.com/alebak/squad-ai/commit/35492c7b2a064a2a438848961a6e22a2bc855279))
* use clean tag format for release-please ([9e2294a](https://github.com/alebak/squad-ai/commit/9e2294accbf80ada9c66cca3c49ec9737eab9cb1))


### Documentation

* add project documentation ([4281cc2](https://github.com/alebak/squad-ai/commit/4281cc2a9e3b3dbed507544a63bc0cf9af0da616))
* rewrite README with bilingual content and usage examples ([c7d9a67](https://github.com/alebak/squad-ai/commit/c7d9a673ffcb15068cb38e07c1e8ca6fe81dec59))
* split README into English and Spanish, update author name ([8524856](https://github.com/alebak/squad-ai/commit/8524856fce77779bbb387eefed504cd0d3fc3c16))


### Code Refactoring

* **cli:** split long functions into helpers, remove dead code ([73fba2d](https://github.com/alebak/squad-ai/commit/73fba2db9dd29748cb241b9a8e302d0b5937ec00))
* replace manual agent installs with Squad AI in post-create.sh ([18bbe83](https://github.com/alebak/squad-ai/commit/18bbe83d36be01b20f71a86860dea0dbb16746c7))

## [0.2.0](https://github.com/alebak/squad-ai/compare/squad-ai-v0.1.0...squad-ai-v0.2.0) (2026-05-24)


### Features

* add GoReleaser config, install script, and ldflags version support ([6a20578](https://github.com/alebak/squad-ai/commit/6a20578cbf951da841e951f540cc2b37aee34e3d))
* bootstrap Go project with squad version command ([1d1321b](https://github.com/alebak/squad-ai/commit/1d1321b6772218ec814f3b858e37e0c266107e3d))
* **cli:** add install, list, add, remove, update, and info commands ([70305b7](https://github.com/alebak/squad-ai/commit/70305b7f39a05ce87d68817718f6d2649339ff29))
* **config:** add user config with atomic load/save ([9850a6a](https://github.com/alebak/squad-ai/commit/9850a6a49dcbccba8c862272b51b2c1fa5d50449))
* **installer:** add agent detection and installation pipeline ([a2aa886](https://github.com/alebak/squad-ai/commit/a2aa886e54c45f01827c4242d34ab6588fb323b1))
* **registry:** add agent types, HTTP fetch, and local cache ([23be0e4](https://github.com/alebak/squad-ai/commit/23be0e4d88220ef094be28d0f0e5e0ed3b3e7104))
* **runtime:** add Node.js, Go, and Python version detection ([bba5dc0](https://github.com/alebak/squad-ai/commit/bba5dc0a2963a4632f3d4cab02056ec48f3265ae))
* **tui:** add interactive agent selection with Bubbletea ([a93a285](https://github.com/alebak/squad-ai/commit/a93a285cb272ca7e0b43c2b7d17305973e556e47))


### Bug Fixes

* display Squad AI product name in version output ([bd528f7](https://github.com/alebak/squad-ai/commit/bd528f78dc3c9167fe11a27ab9976b2c34b08927))
* feat commits should bump minor version pre-1.0 ([e9b0480](https://github.com/alebak/squad-ai/commit/e9b048019dbccf21f465b0256c71dca000a03615))
* **registry:** add install commands and SHA-256 checksums for all agents ([35492c7](https://github.com/alebak/squad-ai/commit/35492c7b2a064a2a438848961a6e22a2bc855279))


### Documentation

* add project documentation ([4281cc2](https://github.com/alebak/squad-ai/commit/4281cc2a9e3b3dbed507544a63bc0cf9af0da616))
* rewrite README with bilingual content and usage examples ([c7d9a67](https://github.com/alebak/squad-ai/commit/c7d9a673ffcb15068cb38e07c1e8ca6fe81dec59))
* split README into English and Spanish, update author name ([8524856](https://github.com/alebak/squad-ai/commit/8524856fce77779bbb387eefed504cd0d3fc3c16))


### Code Refactoring

* **cli:** split long functions into helpers, remove dead code ([73fba2d](https://github.com/alebak/squad-ai/commit/73fba2db9dd29748cb241b9a8e302d0b5937ec00))

## [0.1.0] — 2026-05-24

### Features

- Bootstrap Go project with `squad version` command
- Agent registry with HTTP fetch and local cache (7 MVP agents)
- User config with atomic read/write (`~/.config/squad-ai/config.json`)
- Runtime detection: Node.js, Go, Python version parsing
- Agent detection via `exec.LookPath`
- Installation pipeline with checksum verification and log capture
- CLI commands: install, list, add, remove, update, info
- Interactive TUI for agent selection (Bubbletea)
- GoReleaser config and install script for distribution
- Release-please for automated changelog and versioning

[0.1.0]: https://github.com/alebak/squad-ai/releases/tag/v0.1.0
