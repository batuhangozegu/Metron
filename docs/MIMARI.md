# AETHER v2 — Mimari Kararlar

13 Kasım 2026

Not: Kodlamaya başlanmadı; bu dosya kararları ve gerekçelerini tutar. Özellik listesi için [OZELLIKLER.md](OZELLIKLER.md) dosyasına bakın.

## 1. İstemci framework'leri

- **Web: Next.js (React + TypeScript).** v2'nin detaylı kısmı (grafikler, backtest sonuçları, portföy analizi, yoğun tablolar) web'de yaşayacak. Finans grafiği ve tablo/panel bileşenleri JavaScript tarafında daha olgun; Flutter web yoğun veri ekranlarında zayıf.
- **Mobil: Flutter (v1'deki gibi, Riverpod ile).** Flutter zaten biliniyor. Backend'de iki yeni dil (Go, Rust) öğrenilirken mobilde üçüncü bir yenilik projeyi yavaşlatır.
- **İki istemcinin maliyeti.** Ekranlar iki kez yazılır. Azaltmak için API OpenAPI ile tanımlanır, TypeScript ve Dart istemci kodu otomatik üretilir; elle yazılan sadece arayüz olur.
- **Elenen alternatifler.** Flutter'ı web için de kullanmak (tek kod tabanı ama zayıf web) ve React Native'e geçmek (kod paylaşımı ama öğrenilecek yeni bir şey daha) elendi.

## 2. Web ve mobilin rolleri

İki istemci aynı uygulamanın kopyası olmayacak.

- **Web, analiz masası.** Portföy analizi, backtest, risk ayarları, coin sınıfı kuralları ve işlem günlüğü burada olur. Geniş ekran, yoğun veri.
- **Mobil, takip ve hızlı işlem.** Açık pozisyonlar, alarmlar, emir öncesi kontrol ekranı, hızlı emir ve bildirimler burada olur. Backtest gibi ağır ekranlar mobilde olmayacak.

## 3. Tasarım yönü

- **Ana ekran riski gösterir.** Ana ekran fiyatları değil risk durumunu gösterir: toplam açık risk, günlük kayıp limitine kalan pay, çiğnenen kurallar.
- **İmza ekranı.** Emir öncesi kontrol ekranı tasarımın merkezi ve uygulamanın imza ekranıdır; en çok emek buraya gider.
- **Renk kullanımı.** Renk sadece kırmızı/yeşil değildir: kâr/zarar için renk yanında işaret ve ikon kullanılır. Risk seviyesi için ayrı bir skala vardır (güvenli, dikkat, limit aşıldı).
- **Tema ve rakamlar.** Koyu tema öncelikli, açık tema da desteklenir. Rakamlar eşit genişlikli yazılır ki tablolarda hizalı dursun.
- **Tek tasarım sistemi, iki uygulama.** Renk, yazı ve boşluk değerleri bir kez tanımlanır; web ve mobil aynı değerleri kullanır.

## 4. Servisler ve iletişim

- **Genel yapı.** Mikroservis değil, iki servis: bir Go servisi (API) ve bir Rust servisi (motor). Her biri kendi içinde modüllere ayrılır ama ayrı dağıtılan çok sayıda parça olmaz; tek kişilik projede fazlası yük getirir.

| Bağlantı | Yöntem | Ne için |
| --- | --- | --- |
| İstemci ↔ Go | REST + WebSocket | Normal istekler REST, canlı fiyat ve alarmlar WebSocket |
| Go → Rust | gRPC | "Bu emrin riskini hesapla" gibi soru-cevap çağrıları |
| Rust → Go | Redis yayın/abone | Fiyat akışı, alarm ve stop tetiklenmesi gibi olaylar |

- **gRPC gerekçesi.** İki dil arasındaki sözleşme tek dosyada tanımlanır, Go ve Rust kodu ondan üretilir; yanlış alan adı ya da tip derleme anında yakalanır.
- **Redis gerekçesi.** Son fiyatları da tutar; Go fiyat lazım olduğunda Rust'a sormadan Redis'ten okur.

### 4.1 Kim neyin sahibi

- **Rust motoru.** Binance fiyat akışına bağlanır, mumları ve ATR'yi hesaplar, risk kurallarını çalıştırır, açık pozisyonların stoplarını izler. Borsaya emir göndermez, kullanıcı verisine yazmaz.
- **Go servisi.** Kimlik doğrulama, kullanıcı ve portföy verisi, borsa API anahtarları (şifreli), veritabanına yazma, borsaya emir gönderme. İstemcilerin konuştuğu tek kapı.
- **Gerekçe.** API anahtarları tek serviste kalır ve borsaya tek yerden emir gider.

### 4.2 Bir emrin yolu

1. Kullanıcı emri girer, istemci Go'ya gönderir.
2. Go, emri ve kullanıcının risk profilini gRPC ile Rust'a sorar.
3. Rust lotu, stop seviyesini ve çiğnenen kuralları döner; emir öncesi kontrol ekranı bu cevabı gösterir.
4. Kullanıcı onaylarsa Go emri borsaya gönderir ve veritabanına yazar.

### 4.3 İki kural

- **Stop her zaman borsada da durur.** Sistem stop'u sadece kendi içinde izlerse sunucu çöktüğünde pozisyon korumasız kalır. Her pozisyon için borsaya gerçek bir stop emri konur; motor iz süren stop gibi durumlarda o emri günceller.
- **Paper trading aynı yoldan geçer.** Go tarafında borsa bağlantısı değiştirilebilir bir parçadır (biri gerçek Binance, biri sanal borsa). Risk motoru ikisini ayırt etmez; paper trading ayrı bir sistem olarak yazılmaz.

## 5. Açık kararlar

1. Veritabanı: v1 şemasından ne kalır, coin sınıfları ve "varlık sınıfı" kavramı nasıl modellenir.
2. Depo yapısı: tek depo mu, ayrı depolar mı.
