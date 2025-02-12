# IPRangeSource module for caddy

This module retrieves Icelandic ips from RIX, [ipv4](https://rix.is/is-net.txt) and [ipv6](https://rix.is/is-net6.txt).

## Installation

Build Caddy using [xcaddy](https://github.com/caddyserver/xcaddy):

```shell
xcaddy build --with github.com/Gunni/caddy-iceland-ip
```

## Usage

Intended for use with [caddy-dynamic-remoteip](https://github.com/lanrat/caddy-dynamic-remoteip) to match Icelandic users.

Put following config in global options under corresponding server options

```
:8880 {
    @not_icelandic not dynamic_remote_ip icelandic
    abort @not_icelandic

    reverse_proxy localhost:8080
}
```

# Defaults

| Name     | Description                                            | Type     | Default    |
|----------|--------------------------------------------------------|----------|------------|
| interval | How often icelandic ip lists are retrieved             | duration | 1h         |
| timeout  | Maximum time to wait to get a response from RIX        | duration | no timeout |
