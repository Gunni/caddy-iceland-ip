# trusted_proxy module for caddy

This module retrieves Icelandic ips from RIX, [ipv4](https://rix.is/is-net.txt) and [ipv6](https://rix.is/is-net6.txt).

# Example config

Put following config in global options under corresponding server options

```
trusted_proxies iceland {
    interval 12h
    timeout 15s
}
```

# Defaults

| Name     | Description                                            | Type     | Default    |
|----------|--------------------------------------------------------|----------|------------|
| interval | How often icelandic ip lists are retrieved             | duration | 1h         |
| timeout  | Maximum time to wait to get a response from RIX        | duration | no timeout |
