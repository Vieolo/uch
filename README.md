# uch (Universal Command Helper)

`uch` is a CLI that stores your re-used terminal commands by nickname, encrypts the ones that hold secrets, and prompts for typed variables before running them.

## Install

**Homebrew (macOS / Linux)**:

```bash
brew install vieolo/tap/uch
```

**From source** (using Go):

```bash
go install github.com/vieolo/uch@latest
```


**Pre-built binaries** for Linux, macOS, and Windows are attached to every [GitHub release](https://github.com/vieolo/uch/releases).


## Quick start

```bash
# Plain command
uch add ll --cmd 'ls -la'
uch run ll

# Command with a variable (edit the file to add variables)
uch add fb --cmd 'flutter build {{platform}}'
uch edit                # add a `variables: { platform: { type: select, options: [ios, android, web] } }` block
uch run fb              # prompts for platform via a select TUI

# Sensitive command (encrypted at rest)
uch init                # set a uch password (once)
uch add db-prod --sensitive --cmd 'psql postgres://user:foo@bar/db'
uch run db-prod         # prompts for the uch password, then runs
```

## Storage

- `~/.uch/config.yaml`: your commands (mode 0600).
- `~/.uch/identity.age`: the age identity used to encrypt sensitive commands, itself protected by your uch password (mode 0600). **You should Back up this file** as losing it means losing access to every sensitive command.

## Commands

| Command | Purpose |
|---|---|
| `uch init` | Set the uch password and create the encryption identity. |
| `uch add <name>` | Add a command. Flags: `--cmd`, `--sensitive`, `--description`. |
| `uch list` (`ls`) | Show all stored commands. |
| `uch remove <name>` (`rm`) | Delete a command. `--force` to skip confirmation. |
| `uch run <name>` | Resolve variables, decrypt if sensitive, execute. |
| `uch edit` | Open `config.yaml` in `$EDITOR`. Sensitive commands are masked. |
| `uch edit --admin` | Same as `edit` but decrypts sensitive commands for editing. |
| `uch passwd` | Change the uch password (commands stay valid). |

## Variables

Three types, all static. Variables are an **ordered list** — the user is prompted in the order you declare:

```yaml
commands:
  flutter-build:
    cmd: flutter build {{platform}}
    variables:
      - name: platform
        type: select
        prompt: Pick a platform
        options: [ios, android, web]
  greet:
    cmd: echo "hello, {{name}}"
    variables:
      - name: name
        type: string
        prompt: What's your name?
        default: world
  deploy:
    cmd: ./deploy.sh {{env}} {{confirm}}
    variables:
      - name: env             # prompted first
        type: select
        options: [staging, prod]
      - name: confirm         # prompted second; resolves to "yes" or "no"
        type: confirm
```

## Execution overrides

Each command can optionally pin a working directory and add environment variables:

```yaml
commands:
  flutter-build:
    cmd: flutter build {{platform}}
    cwd: ~/projects/myapp
    env:
      NODE_OPTIONS: --max-old-space-size=8192
    variables:
      - name: platform
        type: select
        options: [ios, android, web]
```

`cwd` accepts a leading `~/` for your home directory; anything else is passed verbatim. `env` is merged onto the inherited environment with per-command entries winning.

## Encryption design

- Each sensitive command is encrypted with an [age](https://github.com/FiloSottile/age) X25519 recipient and stored ASCII-armored in `config.yaml`.
- The X25519 identity itself lives in `~/.uch/identity.age`, encrypted with your uch password via age's scrypt passphrase recipient.
- The uch password is never written to disk. `uch` prompts for it on every sensitive run; session caching is on the roadmap.

## Schema stability

`config.yaml` carries a `version: 1` field at the top. This is the **schema version**, not the CLI version — it increments only when the file structure changes in a way that requires migration. Bug fixes, new commands, and new flags do not bump it.

The following fields are committed-stable: renaming or removing them is a breaking change that requires a schema bump and a migration:

- `version`, `encryption.enabled`, `commands`
- `commands.<name>.cmd`, `commands.<name>.cmd_encrypted`, `commands.<name>.sensitive`
- `commands.<name>.variables.<name>.type` with values `string`, `select`, `confirm`
- `commands.<name>.variables.<name>.options`, `prompt`, `default`

Unknown fields are rejected on load so typos and version mismatches surface immediately.

A `created_by` field is written at every save (e.g. `created_by: uch 0.2.0`) as a diagnostic breadcrumb. It is never used for compatibility logic.

## Roadmap
- [ ] Improve the UX of config manipulation commands, such as interactive `add` commnad
- [ ] Add password caching for the sessions
- [ ] Dynamic `select` variable using the output of the command