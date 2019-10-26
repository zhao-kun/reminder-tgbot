## Telegram reminder bot

A telegram bot which dedicated reminding to group member to check in, and recoding time and person for each checking in


## Build

- Go1.13+
- Go module

Run `make` to build binary target

## Configuration

Configuration file was written by json, named with `config.json`, which must be put in same directory with binary target file. 

```
{
    "channels": [
        XXXXX
    ],
    "tgbot_token": "your telegram bot token",
    "listen_addr": ":8888",
    "webhook_endpoint": "the endpoint which accept the tg webhook",
    "remind": {
        "time_range": {
            "begin": "12:00:00+08:00",
            "end": "18:40:00+08:00"
        },
        "remind_interval": "30m"
    },
    "check_users": [
        "some_one"
    ]
}
```
