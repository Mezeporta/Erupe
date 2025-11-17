# Basic Server Settings

Basic configuration options for Erupe server hosting.

## Configuration

```json
{
  "Host": "127.0.0.1",
  "BinPath": "bin",
  "Language": "en",
  "DisableSoftCrash": false,
  "HideLoginNotice": true,
  "LoginNotices": [
    "<BODY><CENTER><SIZE_3><C_4>Welcome to Erupe!"
  ],
  "PatchServerManifest": "",
  "PatchServerFile": "",
  "ScreenshotAPIURL": "",
  "DeleteOnSaveCorruption": false
}
```

## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `Host` | string | auto-detect | Server IP address. Leave empty to auto-detect. Use `"127.0.0.1"` for local hosting |
| `BinPath` | string | `"bin"` | Path to binary game data (quests, scenarios, etc.) |
| `Language` | string | `"en"` | Server language. `"en"` for English, `"jp"` for Japanese |
| `DisableSoftCrash` | boolean | `false` | When `true`, server exits immediately on crash (useful for auto-restart scripts) |
| `HideLoginNotice` | boolean | `true` | Hide the Erupe welcome notice on login |
| `LoginNotices` | array | `[]` | Custom MHFML-formatted login notices to display to players |
| `PatchServerManifest` | string | `""` | Override URL for patch manifest server (optional) |
| `PatchServerFile` | string | `""` | Override URL for patch file server (optional) |
| `ScreenshotAPIURL` | string | `""` | Destination URL for screenshots uploaded to BBS (optional) |
| `DeleteOnSaveCorruption` | boolean | `false` | If `true`, corrupted save data will be flagged for deletion |

## Login Notices

Login notices use MHFML (Monster Hunter Frontier Markup Language) formatting:

```text
<BODY><CENTER><SIZE_3><C_4>Large Centered Red Text<BR>
<BODY><LEFT><SIZE_2><C_5>Normal Left-Aligned Yellow Text<BR>
<BODY><C_7>White Text
```

**Common MHFML Tags:**

- `<BODY>` - Start new line
- `<BR>` - Line break
- `<CENTER>`, `<LEFT>`, `<RIGHT>` - Text alignment
- `<SIZE_2>`, `<SIZE_3>` - Text size
- `<C_4>` (Red), `<C_5>` (Yellow), `<C_7>` (White) - Text color

## Examples

### Local Development Server

```json
{
  "Host": "127.0.0.1",
  "BinPath": "bin",
  "Language": "en",
  "DisableSoftCrash": false,
  "HideLoginNotice": false
}
```

### Production Server with Auto-Restart

```json
{
  "Host": "",
  "BinPath": "bin",
  "Language": "en",
  "DisableSoftCrash": true,
  "HideLoginNotice": false,
  "LoginNotices": [
    "<BODY><CENTER><SIZE_3><C_4>Welcome to Our Server!<BR><BODY><LEFT><SIZE_2><C_5>Join our Discord: discord.gg/example"
  ]
}
```

## Related Documentation

- [Server Configuration](server-configuration.md) - Server types and ports
- [Configuration Overview](README.md) - All configuration options
