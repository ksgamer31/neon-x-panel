[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/3x-ui-dark.png">
    <img alt="Neon X" src="./media/3x-ui-light.png" width="380">
  </picture>
</p>

<h1 align="center">Neon X — پنل نیونی Xray</h1>

<p align="center">
  <b>زیبا · سریع · چندادمین</b> — بهترین فورک فارسی 3X-UI با تم دارک نیونی و مدیریت ادمین چندسطحی<br>
  <sub>بر پایهٔ <a href="https://github.com/MHSanaei/3x-ui">MHSanaei/3x-ui</a> · Xray-core 26.x · Go + Vue 3 · SQLite / PostgreSQL</sub>
</p>

  <a href="https://github.com/ksgamer31/neon-x-panel/releases"><img src="https://img.shields.io/github/v/release/ksgamer31/neon-x-panel?label=Neon X&labelColor=0a0e1a&color=8b5cf6&style=for-the-badge" alt="Release"></a>
  <a href="https://github.com/ksgamer31/neon-x-panel/actions"><img src="https://img.shields.io/github/actions/workflow/status/ksgamer31/neon-x-panel/release.yml?label=build&labelColor=0a0e1a&color=06ffa5&style=for-the-badge" alt="Build"></a>
  <a href="https://github.com/ksgamer31/neon-x-panel/releases/latest"><img src="https://img.shields.io/github/downloads/ksgamer31/neon-x-panel/total?label=downloads&labelColor=0a0e1a&color=22d3ee&style=for-the-badge" alt="Downloads"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0.en.html"><img src="https://img.shields.io/badge/license-GPL%20V3-0a0e1a?labelColor=22d3ee&color=8b5cf6&style=for-the-badge" alt="License"></a>
  <img src="https://img.shields.io/badge/theme-neon%20dark-0a0e1a?labelColor=8b5cf6&color=06ffa5&style=for-the-badge" alt="Theme">
  <img src="https://img.shields.io/badge/RBAC-چندادمین-0a0e1a?labelColor=06ffa5&color=8b5cf6&style=for-the-badge" alt="RBAC">

<p align="center">
  <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=8b5cf6" alt=""> <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=06ffa5" alt=""> <img src="https://img.shields.io/badge/%20-%20?style=flat&labelColor=0a0e1a&color=22d3ee" alt="">
</p>

<p align="center">
  <code>bash &lt;(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh) v1.0.0</code>
</p>

---

## ✨ چرا Neon X؟

| قابلیت | توضیح |
|---|---|
| 🎨 **تم دارک نیونی** | پس‌زمینهٔ تیره `#0a0e1a` + گرادینت بنفش `#8b5cf6` → فیروزه‌ای `#06ffa5`، کارت‌های شیشه‌ای با هالهٔ نئونی |
| 👥 **چندادمین با سطح دسترسی** | هرکس یوزر/پسورد جدا · ۵ نقش: `viewer` / `creator` / `editor` / `admin` / `owner` + محدودسازی اینباند |
| 🛡️ **همهٔ پروتکل‌ها** | VLESS · VMess · Trojan · Shadowsocks · WireGuard · AmneziaWG · Hysteria2 · MTProto · HTTP · SOCKS · TUN |
| 🚀 **ترنسپورت مدرن** | Reality / XTLS / TLS / WS / gRPC / XHTTP / mKCP — چند پروتکل روی یک پورت با fallback |
| 📊 **مدیریت کلاینت** | حجم، انقضا، محدودیت IP & HWID، آنلاین لحظه‌ای، لینک/QR/سابسکریپشن یک‌کلیک |
| 🌐 **سابسکریپشن هوشمند** | raw / JSON / Clash — انتخاب خودکار از User-Agent + قالب سفارشی |
| 🧩 **چندنود** | کلون اینباند روی نودهای دیگر از یک پنل |
| 🤖 **تلگرام + API** | ربات تلگرام + REST API با توکن‌های محدود و تاریخ‌انقضا |

### 👥 سطوح دسترسی

| نقش | دیدن | ساخت کلاینت | ویرایش کلاینت | ساخت اینباند | تنظیمات |
|---|:---:|:---:|:---:|:---:|:---:|
| `viewer` | ✅ | — | — | — | — |
| `creator` | ✅ | ✅ | — | — | — |
| `editor` | ✅ | ✅ | ✅ | ✅ | — |
| `admin` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `owner` | ✅ | ✅ | ✅ | ✅ | ✅ (+ مدیریت owner) |

> مدیریت از **`/panel/admins`** — ساخت، تغییر نقش، فعال/غیرفعال، ریست پسورد. `inboundIds = []` یعنی همه، مثلا `[1,3]` فقط آن اینباندها.

**Neon X** نسخهٔ بهبودیافته و فارسی‌دوستِ 3X-UI — پنل متن‌باز مدیریت [Xray-core](https://github.com/XTLS/Xray-core) است. پروژهٔ اصلی از [MHSanaei](https://github.com/MHSanaei/3x-ui).

> [!IMPORTANT]
> فقط برای استفادهٔ شخصی. برای کار غیرقانونی یا محیط پروداکشن بدون امن‌سازی استفاده نکنید.

## ✨ نسخهٔ Neon X — چه چیزی جدید است

- **تم دارک نیونی** — هالهٔ بنفش-فیروزه‌ای، کارت‌های شیشه‌ای، برند گرادینت — دارک‌مود حالا خوشگله.
- **چندادمین با سطوح دسترسی** — نقش‌ها: `viewer` فقط دیدن، `creator` فقط ساخت کلاینت، `editor` ساخت+ویرایش بدون تنظیمات، `admin`/`owner` کامل.
- **پنل مدیریت ادمین‌ها** — مسیر `/panel/admins` برای ساخت/حذف/تغییر نقش، فعال/غیرفعال، ریست پسورد.
- **فورک: `ksgamer31/neon-x-panel`** — نصب آسان و سازگار.

## ویژگی‌ها

- **اینباندهای چندپروتکلی** — VLESS، VMess، Trojan، Shadowsocks، WireGuard، AmneziaWG، Hysteria2، MTProto، HTTP، SOCKS (Mixed)، Dokodemo-door / Tunnel و TUN.
- **ترنسپورت‌ها و امنیت مدرن** — TCP (Raw)، mKCP، WebSocket، gRPC، HTTPUpgrade و XHTTP، ایمن‌شده با TLS، XTLS و REALITY.
- **AmneziaWG داخلی** — مقاوم در برابر DPI روی پشتهٔ userspace — بدون ماژول کرنل.
- **پراکسی‌های MTProto** — سکرت FakeTLS و سهمیه به‌ازای هر کلاینت، زنده بدون قطع اتصال.
- **فال‌بک** — چند پروتکل روی یک پورت با fallback.
- **مدیریت کلاینت** — سهمیه، انقضا، محدودیت IP/HWID، تمدید زمان‌بندی، آنلاین زنده، لینک/QR/سابسکریپشن.
- **آمار ترافیک** — هر اینباند/کلاینت/اوتباند با reset.
- **چندنود** — مدیریت چند سرور از یک پنل، کلون اینباند.
- **اوتباند و مسیریابی** — WARP، NordVPN، PIA، routing سفارشی، load balancer، زنجیرهٔ پراکسی. geosite/geoip قابل مرور در ادیتور.
- **سرور سابسکریپشن داخلی** — raw / JSON / Clash + [قالب سفارشی](docs/custom-subscription-templates.md).
- **ربات تلگرام** · **RESTful API** · **PWA** · **SQLite / PostgreSQL** · **۱۳ زبان** · **Fail2ban**.

## 📸 اسکرین‌شات

<details>
<summary>برای باز شدن کلیک کنید</summary>

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

## 🚀 شروع سریع

### نصب با یک دستور (پیشنهادی)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

وقتی `v1.0.0` به عنوان latest ثبت شد، بدون ورژن هم کار می‌کند:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

### ورژن خاص / dev

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
```

حین نصب یوزر/پسورد/مسیر تصادفی می‌سازد. بعدش:

```bash
x-ui              # منوی مدیریت
x-ui settings     # نمایش تنظیمات
x-ui update       # آپدیت به آخرین Neon X
x-ui uninstall    # حذف
```

نتیجه در `/etc/x-ui/install-result.env` ذخیره می‌شود.

### نصب غیرتعاملی (cloud-init)

```bash
XUI_NONINTERACTIVE=1 bash <(curl -Ls https://raw.githubusercontent.com/ksgamer31/neon-x-panel/main/install.sh)
cat /etc/x-ui/install-result.env
```

نمونه‌ها: [`deploy/cloud-init/`](deploy/cloud-init/) و [`deploy/marketplace/hetzner/`](deploy/marketplace/hetzner/)

هر ریلیز با `.sha256` منتشر می‌شود — `install.sh` آن را چک می‌کند.

مستندات کامل: **[docs.sanaei.dev/fa](https://docs.sanaei.dev/fa)**.

## پلتفرم‌های پشتیبانی‌شده

**سیستم‌عامل:** Ubuntu، Debian، Armbian، Fedora، CentOS، RHEL، AlmaLinux، Rocky، Oracle، Amazon، Virtuozzo، Arch، Manjaro، Parch، openSUSE، Alpine و Windows.

**معماری:** `amd64` · `386` · `arm64` · `armv7` · `armv6` · `armv5` · `s390x`.

## گزینه‌های پایگاه‌داده

- **SQLite** (پیش‌فرض) در `/etc/x-ui/x-ui.db`
- **PostgreSQL** برای کلاینت زیاد یا چندنود

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:***@127.0.0.1:5432/xui?sslmode=disable
```

```bash
x-ui migrate-db --dsn "postgres://xui:***@127.0.0.1:5432/xui?sslmode=disable"
# سپس XUI_DB_* در /etc/default/x-ui و:
systemctl restart x-ui
```

Docker: `docker compose up -d` و `docker compose --profile postgres up -d` — برای Fail2ban با `docker run` مقدار `--cap-add=NET_ADMIN --cap-add=NET_RAW` بده.

## متغیرهای محیطی

| متغیر | توضیحات | پیش‌فرض |
| --- | --- | --- |
| `XUI_DB_TYPE` | `sqlite` یا `postgres` | `sqlite` |
| `XUI_DB_DSN` | کانکشن Postgres | — |
| `XUI_LOG_LEVEL` | `debug`/`info`/`warning`/`error` | `info` |
| `XUI_ENABLE_FAIL2BAN` | محدودیت IP با Fail2ban | `true` |

لیست کامل: [مرجع متغیرها](https://docs.sanaei.dev/fa/docs/reference/env-vars).

## زبان‌های پشتیبانی‌شده

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil)

## مشارکت

[CONTRIBUTING.md](/CONTRIBUTING.md) را ببینید.

## تشکر ویژه

- [alireza0](https://github.com/alireza0/)

## قدردانی

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (GPL-3.0)
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (GPL-3.0)

## حمایت

**اگر به دردت خورد یه** :star2: **بده!**

<p align="center">
  <a href="https://www.buymeacoffee.com/MHSanaei"><img src="./media/default-yellow.png" alt="Buy Me A Coffee" width="240"></a>
</p>

## ستاره‌ها در طول زمان

[![Stargazers over time](https://starchart.cc/ksgamer31/neon-x-panel.svg?variant=adaptive)](https://starchart.cc/ksgamer31/neon-x-panel)
