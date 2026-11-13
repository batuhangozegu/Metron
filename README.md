# AETHER v2

AETHER v2, kripto varlıklar için risk yönetimi ve portföy takibi platformudur. v1'in ayrı bir proje olarak sıfırdan yeniden yapılmış halidir: tasarım tamamen yenidir ve kapsam daha detaylıdır. Ana odak, her coinin kendi özelliklerine göre stop-loss ve pozisyon büyüklüğü hesaplayan bir risk motorudur.

## Durum

Planlama aşamasında, henüz kod yok.

## Hedef teknoloji

- **Go:** API gateway, kimlik doğrulama, kullanıcı/portföy CRUD, REST uçları.
- **Rust:** risk motoru, canlı fiyat akışı (WebSocket), PnL/pozisyon hesapları.
- **İstemciler:** web: Next.js (React + TypeScript), mobil: Flutter.

## v1 ile ilişki

v1 (Java/Spring Boot + Flutter) arşivlendi: https://github.com/batuhangozegu/AETHER

v1'den kod taşınmayacak. Referans alınacaklar: API sözleşmesi, Flyway SQL şeması ve risk hesaplama kuralları.

## Belgeler

- [Özellik listesi](docs/OZELLIKLER.md)
- [Mimari kararlar](docs/MIMARI.md)
