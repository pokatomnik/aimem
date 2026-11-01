#!/usr/bin/env sh

export HTTP_PROXY=socks5h://127.0.0.1:1080
export http_proxy=socks5h://127.0.0.1:1080

export HTTPS_PROXY=socks5h://127.0.0.1:1080
export https_proxy=socks5h://127.0.0.1:1080

export ALL_PROXY=socks5h://127.0.0.1:1080
export all_proxy=socks5h://127.0.0.1:1080

exec aimem web "$@"
