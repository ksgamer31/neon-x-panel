[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/3x-ui-dark.png">
    <img alt="Neon X" src="./media/3x-ui-light.png" width="380">
  </picture>
</p>

<h1 align="center">Neon X Panel</h1>

<p align="center">
  <b>زیبا · سریع · چندادمین</b> — بهترین فورک فارسی 3X-UI با تم دارک نیونی و مدیریت ادمین سطح‌بندی‌شده<br>
  <sub>Independent panel · Based on Xray-core · Neon dark theme · Xray-core 26.x · Go + Vue 3 · SQLite / PostgreSQL</sub>
</p>

  <a href="https://github.com/ksgamer31/neon-x-panel/releases"><img src="https://img.shields.io/github/v/release/ksgamer31/neon-x-panel?label=Neon X&labelColor=0a0e1a&color=8b5cf6&style=for-the-badge" alt="Release"></a>
  <a href="https://github.com/ksgamer31/neon-x-panel/actions"><img src="https://img.shields.io/github/actions/workflow/status/ksgamer31/neon-x-panel/release.yml?label=build&labelColor=0a0e1a&color=06ffa5&style=for-the-badge" alt="Build"></a>
  <a href="https://github.com/ksgamer31/neon-x-panel/releases/latest"><img src="https://img.shields.io/github/downloads/ksgamer31/neon-x-panel/total?label=downloads&labelColor=0a0e1a&color=22d3ee&style=for-the-badge" alt="Downloads"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0.en.html"><img src="https://img.shields.io/badge/license-GPL%20V3-0a0e1a?labelColor=22d3ee&color=8b5cf6&style=for-the-badge" alt="License"></a>
  <img src="https://img.shields.io/badge/theme-neon%20dark-0a0e1a?labelColor=8b5cf6&color=06ffa5&style=for-the-badge" alt="Theme">
  <img src="https://img.shields.io/badge/RBAC-multi--admin-0a0e1a?labelColor=06ffa5&color=8b5cf6&style=for-the-badge" alt="RBAC">

<p align="center">
  <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=8b5cf6" alt=""> <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=06ffa5" alt=""> <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=22d3ee" alt="">
</p>

<p align="center">
  <code>bash &lt;(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh) v1.0.0</code>
</p>

---

## ✨ Why Neon X?

| Feature | Details |
|---|---|
| 🎨 **Neon Dark Theme** | Deep navy `#0a0e1a` + purple `#8b5cf6` → cyan `#06ffa5` gradient, glass cards with neon glow, gradient sidebar & buttons |
| 👥 **Multi-Admin RBAC** | Each admin has own username/password · 5 roles: `viewer` / `creator` / `editor` / `admin` / `owner` + per-admin inbound scoping via `inboundIds` |
| 🛡️ **Every Protocol** | VLESS · VMess · Trojan · Shadowsocks · WireGuard · AmneziaWG · Hysteria2 · MTProto · HTTP · SOCKS · TUN |
| 🚀 **Modern Transports** | Reality / XTLS / TLS / WS / gRPC / XHTTP / mKCP — fallbacks let multiple protocols share one port |
| 📊 **Per-Client Power** | Traffic quota, expiry, IP & HWID limits, live online, one-click share link / QR / subscription |
| 🌐 **Smart Subscription** | raw / JSON / Clash — auto-selected from User-Agent + [custom templates](docs/custom-subscription-templates.md) |
| 🧩 **Multi-Node** | Clone inbounds to other nodes from a single panel |
| 🤖 **Telegram + API** | Telegram bot + scoped REST API with expiring tokens |

### 👥 Admin Roles

| Role | View | Create client | Edit client | Create inbound | Settings |
|---|:---:|:---:|:---:|:---:|:---:|
| `viewer` | ✅ | — | — | — | — |
| `creator` | ✅ | ✅ | — | — | — |
| `editor` | ✅ | ✅ | ✅ | ✅ | — |
| `admin` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `owner` | ✅ | ✅ | ✅ | ✅ | ✅ (+ manage owners) |

> Manage at **`/panel/admins`** — create, change role, enable/disable, reset password. `inboundIds = []` means all, e.g. `[1,3]` limits to those inbounds.

**Neon X** is an enhanced, Persian-friendly fork of 3X-UI — an advanced open-source web panel for [Xray-core](https://github.com/XTLS/Xray-core). Original project by [MHSanaei](https://github.com/MHSanaei/3x-ui).

> [!IMPORTANT]
> For personal use only. Do not use for illegal purposes or in production without hardening.

## ✨ Neon X Edition — What's New

- **Neon Dark Theme** — purple-cyan glow, glass cards, gradient brand — dark mode is now gorgeous.
- **Multi-Admin RBAC** — `viewer` (read-only), `creator` (create clients only), `editor` (create+edit, no settings), `admin` / `owner` (full). Per-admin inbound scoping.
- **Admin Panel** — `/panel/admins` to manage users, roles, enable/disable, reset password.
- **Fork: `ksgamer31/neon-x-panel`** — easy install stays compatible.

## Screenshots

<details>
<summary>Click to expand (dark / light)</summary>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/01-overview-dark.png">
  <img alt="Overview" src="./media/01-overview-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/02-add-inbound-dark.png">
  <img alt="Inbounds" src="./media/02-add-inbound-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/03-add-client-dark.png">
  <img alt="Add client" src="./media/03-add-client-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/05-add-nodes-dark.png">
  <img alt="Configs" src="./media/05-add-nodes-light.png">
</picture>

</details>

> Neon X neon in dark mode: gradient **Neon X** logo in sidebar, neon-bordered cards, purple→cyan gradient buttons with glow on hover.

---

## 🚀 Quick Start

### One-line install (recommended)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

Once `v1.0.0` is `latest` (it is), the short form also works:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

### Specific version / dev channel

```bash
# a specific tag
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)

# rolling dev (latest commit on main, not stable)
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

During install a random username / password / path is generated. Afterwards:

```bash
x-ui              # management menu
x-ui settings     # show settings
x-ui update       # update to latest Neon X
x-ui uninstall    # remove
```

Result is saved to `/etc/x-ui/install-result.env` (mode 600).

### Unattended (cloud-init)

```bash
XUI_NONINTERACTIVE=1 bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
cat /etc/x-ui/install-result.env
```

See [`deploy/cloud-init/`](deploy/cloud-init/) and [`deploy/marketplace/hetzner/`](deploy/marketplace/hetzner/).

Every release ships with a `.sha256` file — `install.sh` verifies it and aborts on mismatch.

Full docs: **[docs.sanaei.dev](https://docs.sanaei.dev)**.

## Features

- **Multi-protocol inbounds** — VLESS, VMess, Trojan, Shadowsocks, WireGuard, AmneziaWG, Hysteria2, MTProto, HTTP, SOCKS (Mixed), Dokodemo-door / Tunnel, and TUN.
- **Modern transports & security** — TCP (Raw), mKCP, WebSocket, gRPC, HTTPUpgrade, and XHTTP, secured with TLS, XTLS, and REALITY.
- **AmneziaWG built in** — DPI-resistant WireGuard on a userspace stack — no kernel module needed.
- **MTProto proxies** — per-client FakeTLS secrets, ad-tags, quotas, applied live.
- **Fallbacks** — serve multiple protocols on one port via Xray fallbacks.
- **Per-client management** — quotas, expiry, IP limits, HWID, renewal cycles, live online, share links / QR / subscriptions.
- **Traffic statistics** — per inbound / client / outbound with reset.
- **Multi-node** — manage many servers from one panel, clone inbounds.
- **Outbound & routing** — WARP, NordVPN, PIA, custom routing, load balancers, proxy chaining. Bundled geosite/geoip browsable in editor.
- **Built-in subscription server** — raw, JSON, Clash auto-selected by User-Agent + [custom templates](docs/custom-subscription-templates.md).
- **Telegram bot** · **RESTful API** with scoped tokens · **PWA** · **SQLite / PostgreSQL** · **13 languages** · **Fail2ban**.

## Supported Platforms

**OS:** Ubuntu, Debian, Armbian, Fedora, CentOS, RHEL, AlmaLinux, Rocky, Oracle, Amazon, Virtuozzo, Arch, Manjaro, Parch, openSUSE, Alpine, Windows.

**Arch:** `amd64` · `386` · `arm64` · `armv7` · `armv6` · `armv5` · `s390x`.

## Database Options

- **SQLite** (default) at `/etc/x-ui/x-ui.db`.
- **PostgreSQL** for large / multi-node. Installer can install it or take a DSN:

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:***@127.0.0.1:5432/xui?sslmode=disable
```

```bash
x-ui migrate-db --dsn "postgres://xui:***@127.0.0.1:5432/xui?sslmode=disable"
# set XUI_DB_* in /etc/default/x-ui then:
systemctl restart x-ui
```

Docker: `docker compose up -d` (SQLite) · `docker compose --profile postgres up -d`.

Fail2ban needs `NET_ADMIN` — `docker-compose.yml` already adds it; with `docker run` add `--cap-add=NET_ADMIN --cap-add=NET_RAW`.

## Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `XUI_DB_TYPE` | `sqlite` or `postgres` | `sqlite` |
| `XUI_DB_DSN` | Postgres DSN | — |
| `XUI_LOG_LEVEL` | `debug` / `info` / `warning` / `error` | `info` |
| `XUI_ENABLE_FAIL2BAN` | IP-limit via Fail2ban | `true` |

Full list: [environment variables reference](https://docs.sanaei.dev/docs/reference/env-vars).

## Supported Languages

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil)

## Contributing

See [CONTRIBUTING.md](/CONTRIBUTING.md).

## A Special Thanks to

- [alireza0](https://github.com/alireza0/)

## Acknowledgment

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (GPL-3.0)
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (GPL-3.0)

## Support project

**If this project is helpful, please give it a** :star2:

<a href="https://www.buymeacoffee.com/MHSanaei" target="_blank">
<img src="./media/default-yellow.png" alt="Buy Me A Coffee" style="height: 70px !important;width: 277px !important;" >
</a>

</br>
<a href="https://nowpayments.io/donation/hsanaei" target="_blank" rel="noreferrer noopener">
   <img src="./media/donation-button-black.svg" alt="Crypto donation button by NOWPayments">
</a>

## Stargazers over Time

[![Stargazers over time](https://starchart.cc/ksgamer31/neon-x-panel.svg?variant=adaptive)](https://starchart.cc/ksgamer31/neon-x-panel)
