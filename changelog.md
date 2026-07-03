# Change Log

## v0.2.1 (2026-07-03)
- Improved the structure of the outputs
- Improved the experience while using the `add` function

## v0.2.0 (2026-05-15)

#### Breaking changes
- Switched the central config from JSON to YAML (`~/.uch/config.yaml`).
- Introduced encryption for sensitive commands using [age](https://github.com/FiloSottile/age). A new uch password unlocks an X25519 identity stored at `~/.uch/identity.age`.
- Added commands: `init`, `add`, `list` (`ls`), `remove` (`rm`), `passwd`. Implemented `run` and rewrote `edit` (with a `--admin` tier for editing sensitive bodies).
- Added typed variables (`string`, `select`, `confirm`) with `{{name}}` placeholder substitution. Variables are an ordered list, so the user is prompted in authored order.
- Added per-command `cwd` and `env` execution overrides.
- Added schema version 1 to `config.yaml`. Unknown fields are now rejected on load; legacy unversioned configs auto-migrate. A diagnostic `created_by` breadcrumb is stamped on every save.
- `Load()` now runs a `Validate()` pass that catches inconsistent commands (sensitive without ciphertext, both cmd and cmd_encrypted set, duplicate or unnamed variables, select with no options) before they reach `run`.
- Removed the unused `encryption.enabled` field. Encryption is keyed off the existence of `identity.age`; the field was dead weight.
- Writes to `config.yaml` and `identity.age` are now atomic (temp + fsync + rename).

## v0.1.0 (2025-06-08)
- Initial release
