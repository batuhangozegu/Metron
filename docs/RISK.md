# Pre-trade risk: spot and USDT-M futures / İşlem öncesi risk

Rust owns decimal sizing and rule decisions. Go authenticates the session,
reads the owned paper account's equity and market from `app.accounts`, and
calls `metron.engine.v1.RiskService/Check`. This previews a trade; it never
submits an order. Account type (paper/live) and market (spot/futures) are
independent. Live exchange execution remains step 8.

## Common calculation

The engine reads the latest stored volume class and Wilder daily ATR(14)
from `market.symbols`; neither comes from the client. BTC/ETH are core,
named stablecoins are cash and unsupported for ATR sizing. Other coins are
large at the configured mean daily USDT volume threshold (default 10,000,000),
otherwise small; unknown volume is small. Missing symbols return 404.
Missing/non-positive ATR, unavailable storage, snapshots older than one hour
or more than one minute in the future return 503. Entry is the caller's
proposed price, independent of Redis.

| Class | Risk per trade cap | ATR multiplier | Spot class allocation cap |
| --- | --- | --- | --- |
| core | 2% | 1.5 | 40% |
| large | 1% | 1.5 | 15% |
| small | 0.5% | 2 | 8% |
| stable | Unsupported | None | No ATR sizing |

These are provisional system defaults. Profile `risk_per_trade_percent`
defaults to `"2"`, `target_risk_reward` to `"2"`. Empty/omitted decimal profile
fields and absent count fields use defaults; explicit zero limits are invalid.
Overrides are editable per request; profiles and checks are not persisted yet.
Effective risk is `min(profile risk, class cap)` in both markets.

```text
suggested_stop = entry - ATR * multiplier       # long; + for short
stop = manual_stop if supplied, otherwise suggested_stop
distance = abs(entry - stop)
risk_amount = equity * effective_risk_percent / 100
raw_size = risk_amount / distance              # v1 formula
size = floor(raw_size / quantity_step) * quantity_step
notional = size * entry
actual_risk = size * distance                  # <= risk_amount
target_price = entry +/- distance * target_risk_reward
```

Prices, balances, percentages and quantities travel as plain decimal strings,
never floats. Inputs accept up to eight fractional digits and at most
1,000,000,000,000; no exponent, sign or whitespace. Prices, equity and profile
limits must be positive; activity percentages and holdings may be zero.
Profile percentages and target risk/reward are at most 100. Count limits are
positive uint32 integers. Manual stops must be below entry for longs and above
entry for shorts; equality is a zero-distance error. A non-positive ATR
suggestion is rejected even when a manual stop exists.

Stored ATR can have 28 fractional digits; round it up to eight before sizing.
`source_atr` exposes storage precision; `atr` is the effective value. The
provisional quantity step is `0.00000001` base units for every pair; exchange
lot/tick/minimum-notional filters must be supplied before execution. The pure
function supports arbitrary decimal steps such as `0.05`. It removes
`budget % (distance * step)` before division, preventing a repeating quotient
from rounding across a quantity boundary. Exact products/additions that
overflow or lose precision return `arithmetic_range`. Ratios (including margin
and liquidation displays) use Decimal division up to 28 places; limit decisions
use exact cross-products rather than rounded display ratios. Output drops
trailing zeros. `risk_amount` is the budget; `actual_risk` reflects size rounding.

## Spot rules

Spot requires absent leverage or leverage 1; shorts fail `spot_side`.
The `position_limit` allocation rule compares:

```text
current_class_holding + proposed_notional <= equity * class_allocation_cap / 100
```

`current_class_holding` is the quote-currency value already held in the same
coin class, default zero. These 40%/15%/8% ceilings are portfolio allocation
rules, exclusively for spot. They flag the size rather than silently shrinking
it. Spot responses have effective `leverage: 1`; optional futures fields
`liquidation_price` and `initial_margin` are absent.

## Futures rules and approximation

USDT-M futures requires integer leverage 1..125 and supports both directions.
There is no class allocation check; notional/equity is informational only and
`max_position_percent` is empty. Per-trade risk, ATR and discipline still apply.

Assume one isolated linear USDT-M position, no extra collateral, entry E,
leverage L and provisional maintenance margin rate m = 0.005 (0.5%):

```text
initial_margin = notional / L
long_liquidation  = E * (L - 1) / (L * (1 - m))
short_liquidation = E * (L + 1) / (L * (1 + m))
```

These follow from isolated collateral plus unrealized PnL equaling maintenance
margin on mark notional. A 1x long has estimated liquidation at zero.
`initial_margin` fails when notional > equity * L. `liquidation_stop` fails
for long stop <= liquidation or short stop >= liquidation; equality fails.
It cautions on the safe side within 1% of entry price of liquidation, including
that buffer boundary. The engine compares exact cross-products before dividing
for display, so a rounded liquidation price cannot permit an unsafe stop.

This is not an exchange liquidation quote: maintenance tiers/deductions, fees,
liquidation fees, funding, cross margin, other positions, extra collateral,
mark-versus-last price triggers, slippage and execution gaps are not modeled.
It cannot guarantee a fill before liquidation. Futures-specific market data and
exchange filters are not connected yet: this preview uses the existing
spot-derived daily ATR/class snapshot. Paper execution is step 4 and real
exchange integration is step 8.

## Discipline in both markets

Inputs describe activity before the proposed trade. Days start at 00:00 UTC;
weeks start Monday 00:00 UTC. The engine expects pre-aggregated counters and
does not calculate calendar windows itself.

| Rule id | Activity input | Editable profile field | Default | Fail condition |
| --- | --- | --- | --- | --- |
| total_open_risk | open_risk_percent | total_open_risk_limit_percent | 6% | current + new effective trade risk > limit |
| open_positions | open_positions | max_open_positions | 5 | current + 1 > limit |
| losing_trades_today | losing_trades_today | max_losing_trades_today | 3 | current >= limit; block rest of day |
| daily_loss | daily_loss_percent | daily_loss_limit_percent | 5% | current >= limit; block rest of day |
| stopped_trades_this_week | stopped_trades_this_week | max_stopped_trades_this_week | 5 | current >= limit; block rest of week |
| weekly_loss | weekly_loss_percent | weekly_loss_limit_percent | 10% | current >= limit; block rest of week |

Percentage rules caution at 75% of the limit; count rules at limit - 1, using
the proposed count for open positions. Total open risk includes the conservative
effective risk budget, not the slightly smaller rounded actual risk. Exactly
filling open-risk/position capacity is a caution; reaching a loss/count brake
is a failure. Every check includes a stable rule id, pass/caution/fail status,
current/proposed/limit values and English/Turkish messages. Any fail makes
`allowed=false`; `verdict` is the worst status, with separate passed/cautions/
failed counts. Common checks also include `class_risk` (caution if capped),
`atr_stop` (caution if tighter than suggested), `quantity_step` (fail if zero
size), and `target_price` (fail if non-positive).

**TODO(step 4): API activity is incomplete.** Public clients cannot supply
activity counters, holdings, market, equity, class or ATR. Go currently passes
explicit zeros for every activity counter/percentage and current class holdings
because authoritative positions/fills do not exist yet. Discipline evaluates
zero history plus the proposed trade; spot allocation evaluates zero existing
holdings. These results do not establish the actual account's capacity.
Replace zeros with owner-filtered positions/fills and UTC day/week aggregates
in step 4. Never use this preview as execution authorization.

## HTTP, generation and verification

`POST /v1/risk/check` requires a full `metron_session` cookie and JSON content
type. Strict JSON rejects unknown fields/trailing values and bodies over 4 KiB.
Responses use `Cache-Control: no-store`; pending sessions cannot access it.
Go filters by owner and account UUID, accepting paper/USDT accounts, using
`paper_balance` until mark-to-market equity exists. Unknown/foreign/unsupported
accounts share 404. Engine deadline: four seconds within auth's five seconds.

Example for an account created with `"market":"futures"`:

```json
{
  "account_id": "11111111-1111-4111-8111-111111111111",
  "symbol": "BTCUSDT",
  "side": "short",
  "entry_price": "100",
  "manual_stop": "105",
  "leverage": 10,
  "risk_profile": {
    "risk_per_trade_percent": "1",
    "total_open_risk_limit_percent": "6",
    "max_open_positions": 5,
    "max_losing_trades_today": 3,
    "daily_loss_limit_percent": "5",
    "max_stopped_trades_this_week": 5,
    "weekly_loss_limit_percent": "10"
  }
}
```

`manual_stop`, `risk_profile` and spot leverage are optional; null means omitted.
Responses retain sizing fields and bilingual checks and add market, effective
integer leverage, and optional decimal-string liquidation_price/initial_margin
for futures. With equity 10000, this example gives budget 100, size 20, notional
2000, margin 200 and estimated liquidation 109.45273631840796019900497512.
Rule failures return 200 with allowed=false. Validation returns 400; session
failure 401; missing account/symbol 404; wrong content type 415; unavailable
dependencies/fresh data 503. Errors expose error/error_tr and a stable risk code,
never private dependency details.

Rust build.rs uses vendored protoc. API `go generate ./...` runs pinned Buf/Go
plugins; generated files stay ignored. Engine checks: `cargo fmt --check`,
`cargo clippy --all-targets -- -D warnings`, `cargo test`. API checks:
`go generate ./...`, `gofmt -l .` (empty), `go vet ./...`, `go test ./...`.
Proto: `buf lint` or API `go tool -modfile=tools/go.mod buf lint ../proto`.
Tests cover hand-calculated sizes, exact step and liquidation boundaries,
allocation with existing holdings, all discipline thresholds, overrides,
leverage, gRPC verdicts and HTTP ownership, trusted inputs and error policies.
DB integration requires TEST_DATABASE_URL and both app migrations.

## Türkçe

Rust kesin ondalık hesaplama ve kural kararlarını üretir. Go oturumu doğrular;
kullanıcıya ait kağıt hesabın bakiyesini ve market değerini veritabanından okur.
Spot ve USDT-M futures ayrı kurallara sahiptir; kağıt/gerçek hesap türünden
bağımsızdır. Bu uç nokta emir göndermez; gerçek borsa adım 8'dedir.

Her iki piyasada etkin risk profil yüzdesi ile sınıf sınırının küçüğüdür.
Varsayılan profil riski %2, hedef risk/getiri oranı 2'dir. Core/large/small risk
sınırları %2/%1/%0,5; ATR çarpanları 1,5/1,5/2'dir. Sınıf ve günlük Wilder
ATR(14) sunucudan alınır; sabit coinler nakittir, ATR hesabı desteklenmez.
Uzun stop girişin altında, kısa stop üstünde olmalıdır; eşitlik reddedilir.
Elle stop yoksa girişten ATR × çarpan çıkarılır (uzun) veya eklenir (kısa).
Risk bütçesi = öz sermaye × etkin risk / 100; miktar = bütçe / stop mesafesi,
daima miktar adımına aşağı yuvarlanır; gerçek risk bütçeyi aşmaz.

Fiyat/bakiye/miktarlar kayan nokta değil ondalık dizelerdir. Girdiler en fazla
8 ondalık basamak ve 1.000.000.000.000 kabul eder; işaret, üs ve boşluk kabul
edilmez. Profil yüzdeleri ve hedef oran en fazla 100'dür. Sınırlar pozitif,
faaliyet ve mevcut tutarlar sıfır olabilir. ATR 8 basamağa yukarı yuvarlanır.
Geçici miktar adımı 0,00000001'dir; borsa filtreleri henüz bağlı değildir.
Kesin çarpım/toplam taşarsa hata verilir; tekrarlayan oranların gösterimi
Decimal bölmesinin en fazla 28 basamaklı hassasiyetini kullanır.

Spot yalnızca uzun işlem ve kaldıraç 1 (veya boş) kabul eder. Mevcut sınıf
tutarı + yeni pozisyon tutarı öz sermayenin core için %40, large için %15,
small için %8'ini aşarsa position_limit başarısızdır. Bunlar yalnızca spot
portföy dağılım sınırlarıdır; vadeli işlemlerde uygulanmaz. Miktar sessizce
küçültülmez. Spot yanıtında tasfiye/başlangıç marjini alanları bulunmaz.

USDT-M vadeli işlemler 1–125 tam sayı kaldıraç ve her iki yönü destekler.
Başlangıç marjini = pozisyon tutarı / kaldıraç; öz sermayeyi aşarsa başarısızdır.
Geçici bakım marjini m = 0,005 ile izole tasfiye tahmini uzun için
E × (L−1) / (L × (1−m)), kısa için E × (L+1) / (L × (1+m)) olur. Uzun stop
tasfiyenin kesinlikle üstünde, kısa stop altında olmalıdır; eşitlik başarısızdır.
Güvenli tarafta ama tasfiyeye giriş fiyatının %1'i kadar veya daha yakın stop
uyarı üretir. Karşılaştırmalar yuvarlanmış fiyattan yapılmaz. Tahmin borsa
fiyatı değildir: bakım kademeleri, komisyon, fonlama, çapraz marjin, ek teminat,
diğer pozisyonlar ve kayma yoktur; stopun önce gerçekleşmesini garanti etmez.
Henüz spot kaynaklı ATR/sınıf kullanılır; kağıt işlem yürütmesi adım 4'tedir.

Disiplin her iki piyasada değerlendirilir. Toplam açık risk varsayılanı yeni
etkin risk bütçesi dahil %6; en fazla açık pozisyon 5'tir. Bu iki kapasiteyi
tam doldurmak uyarı, aşmak başarısızlıktır. Günlük 3 zararlı işleme veya %5
zarara ulaşılması gün sonuna kadar; haftalık 5 stop olan işleme veya %10
zarara ulaşılması hafta sonuna kadar yeni işlemleri engeller. Gün UTC
00:00'da, hafta pazartesi UTC 00:00'da başlar. Yüzde sınırının %75'inde, sayı
sınırının bir eksiğinde uyarı verilir. Tüm profil sınırları istek başına
değiştirilebilir; sıfır sınır geçersizdir. Her kural kimlik, durum, değerler ve
iki dilde metin döndürür; tek başarısızlık allowed=false yapar.

**Adım 4 eksikliği:** API faaliyet sayaçları ve mevcut sınıf tutarını istemciden
kabul etmez; pozisyon/dolum kayıtları olmadığından hepsini açıkça sıfır gönderir.
Sonuç gerçek geçmişi doğrulamaz. Adım 4'te sahiplik filtreli pozisyon/dolumlardan
UTC gün/hafta toplamları hesaplanmalıdır. Profil ve kontroller henüz kaydedilmez.

POST /v1/risk/check tam oturum çerezi ve JSON gerektirir. Örnekteki alan adları
dilden bağımsızdır; istemci piyasa, bakiye, faaliyet, sınıf veya ATR gönderemez.
Yanıtta piyasa/kaldıraç; vadeli işlemlerde tasfiye fiyatı/başlangıç marjini vardır.
Geçerli başarısız kontrol HTTP 200, hatalı girdi 400, eksik oturum 401,
bulunamayan hesap/çift 404, yanlış içerik türü 415, erişilemeyen/eski veri 503
döndürür. Hatalar ve kural metinleri iki dillidir. Bir saatten eski veya bir
dakikadan fazla gelecekteki veriyle hesap yapılmaz. Yukarıdaki üretim/test
komutları her iki piyasa için geçerlidir; veritabanı testi TEST_DATABASE_URL
ve her iki app geçişini gerektirir.
