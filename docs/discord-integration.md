# Discord Integration

Real-time Discord bot integration for posting server activity to Discord channels.

## Configuration

```json
{
  "Discord": {
    "Enabled": false,
    "BotToken": "",
    "RealtimeChannelID": ""
  }
}
```

## Settings Reference

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable Discord integration |
| `BotToken` | string | Discord bot token from Discord Developer Portal |
| `RealtimeChannelID` | string | Discord channel ID where activity messages will be posted |

## How It Works

When enabled, the Discord bot:

1. **Connects on Server Startup**: The bot authenticates using the provided bot token
2. **Monitors Game Activity**: Listens for in-game chat messages and events
3. **Posts to Discord**: Sends formatted messages to the specified channel

### What Gets Posted

- Player chat messages (when sent to world/server chat)
- Player connection/disconnection events
- Quest completions
- Special event notifications

### Message Format

Messages are posted in this format:

```text
**PlayerName**: Hello everyone!
```

Discord mentions and emojis in messages are normalized for proper display.

## Setup Instructions

### 1. Create a Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Give your application a name (e.g., "Erupe Server Bot")
4. Go to the "Bot" section in the left sidebar
5. Click "Add Bot"
6. Under the bot's username, click "Reset Token" to reveal your bot token
7. **Copy this token** - you'll need it for the config

**Important:** Keep your bot token secret! Anyone with this token can control your bot.

### 2. Get Channel ID

1. Enable Developer Mode in Discord:
   - User Settings → Advanced → Developer Mode (toggle on)
2. Right-click the channel where you want bot messages
3. Click "Copy ID"
4. This is your `RealtimeChannelID`

### 3. Add Bot to Your Server

1. In Discord Developer Portal, go to OAuth2 → URL Generator
2. Select scopes:
   - `bot`
3. Select bot permissions:
   - `Send Messages`
   - `Read Message History`
4. Copy the generated URL at the bottom
5. Paste the URL in your browser and select your Discord server
6. Click "Authorize"

### 4. Configure Erupe

Edit your `config.json`:

```json
{
  "Discord": {
    "Enabled": true,
    "BotToken": "YOUR_BOT_TOKEN_HERE",
    "RealtimeChannelID": "YOUR_CHANNEL_ID_HERE"
  }
}
```

### 5. Start Erupe

The bot will connect automatically on server startup. You should see:

```text
INFO    Discord bot connected successfully
```

## Example Configuration

```json
{
  "Discord": {
    "Enabled": true,
    "BotToken": "MTIzNDU2Nzg5MDEyMzQ1Njc4OQ.AbCdEf.GhIjKlMnOpQrStUvWxYz123456789",
    "RealtimeChannelID": "987654321098765432"
  }
}
```

## Implementation Details

- **Bot Code**: [server/discordbot/discord_bot.go](../server/discordbot/discord_bot.go)
- **Library**: Uses [discordgo](https://github.com/bwmarrin/discordgo)
- **Message Normalization**: Discord mentions (`<@123456>`) and emojis (`:emoji:`) are normalized
- **Error Handling**: Non-blocking - errors are logged but don't crash the server
- **Threading**: Bot runs in a separate goroutine

## Troubleshooting

### Bot doesn't connect

**Error:** `Discord failed to create realtimeChannel`

**Solutions:**

- Verify the `RealtimeChannelID` is correct
- Ensure the bot has been added to your server
- Check that the bot has permission to read the channel

### Bot connects but doesn't post messages

**Solutions:**

- Verify the bot has `Send Messages` permission in the channel
- Check channel permissions - the bot's role must have access
- Look for error messages in server logs

### Invalid token error

**Error:** `Discord failed: authentication failed`

**Solutions:**

- Regenerate the bot token in Discord Developer Portal
- Copy the entire token, including any special characters
- Ensure no extra spaces in the config file

### Bot posts but messages are blank

**Issue:** Message normalization may be failing

**Solution:**

- Check server logs for Discord-related errors
- Verify game chat is being sent to world/server chat, not private chat

## Security Considerations

1. **Never commit your bot token** - Add `config.json` to `.gitignore`
2. **Regenerate compromised tokens** - If your token is exposed, regenerate immediately
3. **Limit bot permissions** - Only grant necessary permissions
4. **Monitor bot activity** - Check for unusual posting patterns

## Advanced Usage

### Multiple Channels

Currently, Erupe supports posting to a single channel. To post to multiple channels, you would need to modify the bot code.

### Custom Message Formatting

To customize message formatting, edit [sys_channel_server.go:354](../server/channelserver/sys_channel_server.go#L354):

```go
func (s *Server) DiscordChannelSend(charName string, content string) {
    if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
        // Customize this format
        message := fmt.Sprintf("**%s**: %s", charName, content)
        s.discordBot.RealtimeChannelSend(message)
    }
}
```

### Webhook Alternative

For simpler one-way messaging, consider using Discord webhooks instead of a full bot. This would require code modifications but wouldn't need bot creation/permissions.

## Related Documentation

- [In-Game Commands](commands.md) - Chat commands that may trigger Discord posts
- [Logging](logging.md) - Server logging configuration
