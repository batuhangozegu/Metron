# AETHER v2 — Özellik Listesi

13 Kasım 2026

Not: Bu liste fikir aşamasındadır; kesinleşmiş kapsam değildir.

## 1. v1'deki durum (çıkış noktası)

- **Stop fiyatı elle giriliyor.** v1'de kullanıcı stop fiyatını kendisi giriyor; sistem sadece `risk tutarı / |giriş − stop|` ile lot hesaplıyor. Stop'un doğru yerde olup olmadığını sistem bilmiyor.
- **Risk profili alanları.** İşlem başı risk yüzdesi (varsayılan %2), günlük kayıp limiti (%5), hedef risk/ödül oranı (2.0), otomatik stop-loss anahtarı, azami açık pozisyon sayısı (5).
- **Temel fikir.** Bitcoin ile altcoin'e aynı stop mantığı uygulanmamalı; v2 risk motoru bunun üzerine kurulacak.

## 2. Coine göre stop-loss (öncelik 1)

- **Volatiliteye göre stop (ATR).** Stop mesafesi sabit yüzde değil, coinin son günlerdeki ortalama hareketine göre belirlenir. Örnek (temsili rakamlar): BTC günde ortalama %2 oynuyorsa stop %3, bir altcoin %8 oynuyorsa stop %12. Stop geniş olunca lot otomatik küçülür, riske edilen para aynı kalır.
- **Coin sınıfları.** BTC/ETH, büyük altcoin, küçük altcoin, stablecoin. Her sınıfın kendi ATR çarpanı, azami pozisyon büyüklüğü ve işlem başı risk yüzdesi olur (örnek: BTC %2, küçük altcoin %0.5).
- **Likidite kontrolü.** Emir defteri sığ olan coinde stop tetiklense bile fiyat kayar. Sistem emir defterinden tahmini kaymayı hesaplar, gerekirse lotu küçültür veya uyarır.

## 3. Portföy seviyesinde risk (öncelik 2)

- **Toplam açık risk limiti.** Bütün stoplar aynı anda tetiklenirse kasanın yüzde kaçı gider, buna üst sınır konur. (v1'de sadece pozisyon sayısı limiti var.)
- **Korelasyon.** Beş farklı altcoinde long beş ayrı risk değildir, BTC düşerse birlikte düşerler. Sistem açık pozisyonların toplam riskini korelasyonla hesaplar ve uyarır.
- **Kayıp freni.** Günlük limite ek olarak haftalık limit ve art arda kayıp sonrası bekleme süresi.
- **Piyasa rejimi.** Piyasa genel olarak çok oynaksa tüm risk yüzdeleri otomatik düşer.

## 4. Stop türleri (öncelik 3)

- **İz süren stop.** Fiyat lehe hareket ettikçe stop da onu takip eder.
- **Başa baş stop.** Belli kâra ulaşınca stop girişe çekilir.
- **Kademeli kâr alma.** Pozisyon tek seferde değil, kademeli olarak kapatılır.
- **Zaman stop'u.** Belli sürede hareket etmezse pozisyon kapatılır.

## 5. Portföy yönetimi

Üç konu (portföy yönetimi, stablecoin, tokenize hisse) tek fikir altında toplanıyor: hedef dağılım.

### 5.1 Hedef dağılım (ana fikir)

- **Hedef dağılım tanımı.** Kullanıcı portföyün nasıl görünmesini istediğini bir kez tanımlar (örnek: %40 BTC/ETH, %20 altcoin, %25 stablecoin, %15 tokenize hisse). Sistem gerçek dağılımı sürekli bununla karşılaştırır. v1 sadece "elinde ne var" gösteriyordu; v2 "olması gerekenden ne kadar saptın" sorusunu da cevaplar.
- **Sapma uyarısı.** Bir sınıfın payı hedeften belirgin uzaklaşınca (örnek: altcoin %20'den %35'e) sistem uyarır ve dengeye dönmek için gereken alım satımları listeler.
- **Yeniden dengeleme önerisi.** Sistem emirleri kendisi göndermez, öneriyi hazırlar; kullanıcı onaylarsa uygulanır. Otomatik uygulama sonraki aşamaya kalabilir.
- **Sınıf limitleri.** Risk motorundaki coin sınıflarıyla aynı sınıflar kullanılır. "Küçük altcoin toplamı kasanın %10'unu geçemez" gibi bir kural yeni emri baştan engeller.

### 5.2 Stablecoin (nakit rolü)

- **Nakit kısmı.** Stablecoin ayrı bir özellik değil, portföyün nakit kısmı olarak ele alınır.
- **Riskten kaçış.** Kayıp freni tetiklendiğinde veya piyasa çok oynakken sistem stablecoin payını artırmayı önerir. Risk motoruyla portföy tarafını birbirine bağlar.
- **Kâr park etme.** Kademeli kâr alımlarından çıkan para stablecoin'e geçer ve hedef dağılımda nakit sayılır.
- **Stablecoin'in kendi riski.** Tek stablecoin'e yığılma ve sabit fiyattan kopma (depeg) izlenir; fiyat 1 dolardan belli oranda uzaklaşırsa alarm verilir.

### 5.3 Tokenize hisse (ilk sürümde yok)

- **Amaç.** Kripto dışı varlık sınıfı olarak portföyü çeşitlendirir, ama ilk sürüme alınmayacak.
- **Erişim belirsiz.** Hangi borsada ve hangi ülkeden alınabildiği değişiyor. Binance'in sunup sunmadığı ve Türkiye'den erişim başlamadan önce kontrol edilecek (açık soru).
- **Farklı davranış.** Hisse piyasası kapalıyken token işlem görebilir, fiyat gerçek hisseden kopabilir, likidite düşüktür. Risk motorunda ayrı sınıf gerekir.
- **Ek entegrasyon.** İkinci bir borsa entegrasyonu gerekebilir.
- **Karar.** Veri modeli baştan "varlık sınıfı" kavramıyla esnek kurulacak, hisse sonradan eklenecek.

### 5.4 Portföy takibi

- **Geçmiş performans.** Kasanın zaman içindeki değeri, en büyük düşüş (drawdown), "sadece BTC tutsaydın" karşılaştırması.
- **İşlem günlüğü ve istatistik.** Kazanma oranı, ortalama kazanç/kayıp, hangi coin sınıfında kazanıp hangisinde kaybedildiği. v1'in "duygusal kararların önüne geçme" amacına en çok hizmet eden kısım.
- **Maliyet takibi.** Ortalama alış fiyatı ve ödenen komisyonlar.

### 5.5 Portföy tarafında öncelik sırası

1. Geçmiş performans ve işlem istatistikleri.
2. Hedef dağılım, sapma uyarısı ve stablecoin'in nakit rolü.
3. Yeniden dengeleme önerisi.
4. Tokenize hisse (erişim netleşince).

## 6. v1'den devralınan yapılmamış işler

- **Market service.** Coin listesi, mum verisi (klines), WebSocket + Redis canlı fiyat akışı.
- **Ortak mantık.** Kullanıcı bulma mantığının tek yerde toplanması; PnL hesabının tek ortak modülde olması (v1'de iki yerde tekrar ediyordu).
- **Bildirim sistemi.**
- **İki faktörlü doğrulama.** v1'de backend hazırdı, mobil entegrasyon yoktu.

## 7. Mimari notu

- Bölüm 2-4'teki tüm hesaplar Rust risk motorunda yapılır; Go tarafı API ve kullanıcı işlemlerini yürütür.
- İlk adım: kapsam kesinleşince v1 API'sini OpenAPI'ye dökmek ve veritabanı şemasını uyarlamak.

## 8. Ek özellikler

Risk motoru ve portföy dışında kalan özellikler.

- **Emir öncesi kontrol ekranı.** Emir gönderilmeden önce tek ekranda bu işlemde ne kadar kaybedilebileceği, toplam açık riskin ne olacağı, hedef dağılımdan sapma ve çiğnenen kurallar görünür. AETHER'in ana fikrini ("her emirden önce risk hesabı") görünür kılar; veriyi risk motoru ve portföy tarafı zaten üretir.
- **Kâğıt üstünde işlem (paper trading).** Gerçek para kullanmadan, canlı fiyatlarla sanal kasada işlem yapılır. Kullanıcı kurallarını risksiz dener, sistem de gerçek para riske edilmeden test edilir. v1'deki Binance demo desteğinin borsadan bağımsız hâlidir.
- **Geriye dönük test (backtest).** "ATR çarpanı 1.5 yerine 2 olsaydı geçen ay ne olurdu" sorusunu geçmiş veriyle cevaplar; stop ayarları tahminle değil veriyle seçilir. Rust tarafına uygun, en çok emek isteyen maddedir.
- **Disiplin kuralları.** Kullanıcının kendi koyduğu kurallardır (örnek: "kayıptan sonra 1 saat işlem yok", "gece 00:00'dan sonra emir yok", "stop zarar yönünde genişletilemez"). Sistem kuralı çiğnetmez ya da en azından ihlali kaydeder.
- **Çoklu borsa.** Binance'e ek ikinci bir borsa eklenir ve hepsi tek portföyde birleşir. Tokenize hisse için de gerekebilir.
- **Akıllı alarmlar.** v1'deki fiyat alarmına ek olarak volatilite sıçraması, stop'a yaklaşma, hedef dağılımdan sapma ve depeg alarmları gelir. Mobilde push, web'de anlık bildirim kullanılır.
- **Düzenli alım (DCA).** "Her hafta 50 dolarlık BTC al" gibi planlar tanımlanır. Uzun vadeli kısmı işlem kısmından ayırır.
- **Haftalık rapor.** Hafta sonunda otomatik özet çıkar (kazanç/kayıp, çiğnenen kurallar, en iyi ve en kötü işlem).
- **Vergi/kazanç dökümü.** Yıllık gerçekleşen kazanç raporu ve dışa aktarma sunulur.
- **Vadeli işlemler.** Kaldıraç, likidasyon fiyatı ve fonlama maliyeti gelir. Risk motorunu ciddi büyütür; spot sağlam çalışmadan girilmeyecek.
- **Yapay zekâ yorumu.** İşlem geçmişinden gözlem çıkaran asistandır (örnek: "kayıpların çoğu küçük altcoinlerde, gece açılan işlemlerde").

## 9. Sürüm kademeleri

Listeye yazılmış olmak yapılacak demek değildir. Kural: ilk sürüm bitmeden ikinci kademeden iş başlamaz.

Not: Bu bölüm, önceki bölümlerdeki öncelik notlarının hangi sürüme düştüğünü belirler.

### 9.1 İlk sürüm

- ATR tabanlı stop ve coin sınıfları (bölüm 2)
- Toplam açık risk limiti ve kayıp freni (bölüm 3)
- Portföy takibi: geçmiş performans, işlem istatistikleri, maliyet takibi (bölüm 5.4)
- Emir öncesi kontrol ekranı
- Disiplin kuralları
- Paper trading
- Market service: coin listesi, mum verisi, canlı fiyat akışı (bölüm 6; ATR hesabı için ön koşul)
- İki faktörlü doğrulama (bölüm 6)

### 9.2 İkinci sürüm

- Likidite kontrolü, korelasyon, piyasa rejimi (bölüm 2 ve 3)
- Stop türleri: iz süren, başa baş, kademeli kâr alma, zaman stop'u (bölüm 4)
- Hedef dağılım, sapma uyarısı, stablecoin'in nakit rolü, yeniden dengeleme önerisi (bölüm 5.1, 5.2)
- Backtest
- Akıllı alarmlar ve bildirim sistemi
- Haftalık rapor
- Düzenli alım (DCA)

### 9.3 Fikir havuzu (söz yok)

- Çoklu borsa
- Tokenize hisse (bölüm 5.3)
- Vergi/kazanç dökümü
- Vadeli işlemler
- Yapay zekâ yorumu
