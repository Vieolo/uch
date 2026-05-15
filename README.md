# uch (Universal Command Helper)

`uch` is a CLI that stores your re-used terminal commands by nickname, encrypts the ones that hold secrets, and prompts for typed variables before running them.

## Install

```bash
go install github.com/vieolo/uch@latest
```

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
uch init                # set a master password (once)
uch add db-prod --sensitive --cmd 'psql postgres://user:foo@bar/db'
uch run db-prod         # prompts for the master password, then runs
```

## Storage

- `~/.uch/config.yaml` — your commands (mode 0600).
- `~/.uch/identity.age` — the age identity used to encrypt sensitive commands, itself protected by your master password (mode 0600). **Back this up.** Losing it means losing every sensitive command.

## Commands

| Command | Purpose |
|---|---|
| `uch init` | Set the master password and create the encryption identity. |
| `uch add <name>` | Add a command. Flags: `--cmd`, `--sensitive`, `--description`. |
| `uch list` (`ls`) | Show all stored commands. |
| `uch remove <name>` (`rm`) | Delete a command. `--force` to skip confirmation. |
| `uch run <name>` | Resolve variables, decrypt if sensitive, execute. |
| `uch edit` | Open `config.yaml` in `$EDITOR`. Sensitive commands are masked. |
| `uch edit --admin` | Same as `edit` but decrypts sensitive commands for editing. |
| `uch passwd` | Change the master password (commands stay valid). |

## Variables

Three types, all static:

```yaml
commands:
  flutter-build:
    cmd: flutter build {{platform}}
    variables:
      platform:
        type: select
        prompt: Pick a platform
        options: [ios, android, web]
  greet:
    cmd: echo "hello, {{name}}"
    variables:
      name:
        type: string
        prompt: What's your name?
        default: world
  deploy:
    cmd: ./deploy.sh {{confirm}}
    variables:
      confirm:
        type: confirm        # resolves to "yes" or "no"
```

## Encryption design

- Each sensitive command is encrypted with an [age](https://github.com/FiloSottile/age) X25519 recipient and stored ASCII-armored in `config.yaml`.
- The X25519 identity itself lives in `~/.uch/identity.age`, encrypted with your master password via age's scrypt passphrase recipient.
- The master password is never written to disk. `uch` prompts for it on every sensitive run; session caching is on the roadmap.
