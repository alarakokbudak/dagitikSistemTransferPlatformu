### 1. Kontrol ve Araç Dinamiği Ekibi

Bu ekip, malzeme geldiğinde aracın hareket, yönlendirme, fren ve güvenli duruş komutlarını çalıştıracak ekip.

**Malzeme gelmeden yapılacaklar**

| Hazırlık | Açıklama |
| :--- | :--- |
| **Motor kontrol mantığı** | Hız komutu geldiğinde motor PWM / gaz komutu nasıl üretilecek netleştirilecek. |
| **Direksiyon kontrol mantığı** | Hedef açı geldiğinde sağ-sol dönüş nasıl yapılacak belirlenecek. |
| **Fren kontrol mantığı** | Fren komutu geldiğinde fren aktif/pasif davranışı tanımlanacak. |
| **Emergency stop mantığı** | Acil durumda motor kesme, fren aktif etme, direksiyonu nötrleme mantığı hazırlanacak. |
| **Komut aralıkları** | Hız, direksiyon ve fren için yazılımsal sınırlar belirlenecek. |
| **Test kod iskeleti** | Malzeme gelince direkt denenebilecek Arduino/kontrol kodu hazırlanacak. |

**Bu ekipten beklenen çıktı**

| Çıktı | Durum |
| :--- | :--- |
| `set_motor(speed)` fonksiyonu | Hazır olmalı |
| `set_steering(angle)` fonksiyonu | Hazır olmalı |
| `set_brake(active)` fonksiyonu | Hazır olmalı |
| `emergency_stop()` fonksiyonu | Hazır olmalı |
| Düşük hız test planı | Hazır olmalı |

**Malzeme gelince ilk yapacakları**
Motoru direkt tam güçte denemeyecekler. Önce:
1. Motor düşük PWM ile çalışıyor mu?
2. Direksiyon sağ-sol küçük açılara gidiyor mu?
3. Fren komutu fiziksel karşılık veriyor mu?
4. Emergency komutu motoru kesiyor mu?
5. Timeout olursa sistem güvenli moda geçiyor mu?

**Bu ekip için ana hedef:**
Araç düşük hızda ileri gidecek, sağ-sol tepki verecek, frenle duracak ve emergency durumda güvenli şekilde kesilecek.

---

### 2. Entegrasyon ve Test Ekibi

Bu ekip, yazılım-elektronik-araç arasındaki bağlantıyı yönetecek ekip. Yani sistemin "birlikte çalışmasını" sağlayacaklar.

**Malzeme gelmeden yapılacaklar**

| Hazırlık | Açıklama |
| :--- | :--- |
| **CAN/Serial mesaj protokolü** | Hangi mesaj ne anlama geliyor netleşecek. |
| **Test komutları** | İleri, dur, sağ, sol, fren, emergency komutları hazırlanacak. |
| **Haberleşme test scripti** | Laptop/Jetson'dan Arduino'ya komut gönderen test scripti hazırlanacak. |
| **Log formatı** | Terminalde neyin görüneceği belirlenip log yapısı hazırlanacak. |
| **Test checklist** | Malzeme gelince hangi sırayla ne test edilecek listelenecek. |
| **Video test akışı** | 1 Temmuz videosunda hangi sahneler gösterilecek taslaklanacak. |

**Bu ekipten beklenen çıktı**

| Çıktı | Durum |
| :--- | :--- |
| CAN/Serial mesaj tablosu | Hazır olmalı |
| Komut gönderme scripti | Hazır olmalı |
| Arduino mesaj okuma test kodu | Hazır olmalı |
| Sistem log yapısı | Hazır olmalı |
| Test checklist | Hazır olmalı |
| Video çekim senaryosu | Hazır olmalı |

**Önerilen minimum mesaj yapısı**

| Komut | Açıklama |
| :--- | :--- |
| `mode` | Manuel / otonom / emergency |
| `target_speed` | Hedef hız |
| `target_steering` | Hedef direksiyon açısı |
| `brake_cmd` | Fren aktif/pasif |
| `heartbeat` | Haberleşme canlılık bilgisi |

**Malzeme gelince ilk yapacakları**
1. Elektrikçiler hattı kurunca mesaj al-gönder testi yapılacak.
2. Motor bağlamadan önce seri monitörde mesajlar okunacak.
3. Sonra LED/röle testiyle komutlar doğrulanacak.
4. Daha sonra kontrol ekibiyle birlikte motor/direksiyon/fren çıkışlarına geçilecek.
5. Her test loglanacak ve video için kayıt altına alınacak.

**Bu ekip için ana hedef:**
Sistemi dağınık modüller hâlinden çıkarıp, tek bir test senaryosunda çalışır hâle getirmek.

---

### 3. Algılama ve Görüntü İşleme Ekibi

Bu ekip, kamera ve görüntü işleme tarafını hazır edecek. 1 Temmuz videosunda araç tam otonom sürmese bile yazılım gücünü gösterecek ekip burası.

**Malzeme gelmeden yapılacaklar**

| Hazırlık | Açıklama |
| :--- | :--- |
| **Kamera test altyapısı** | Webcam, araç kamerası veya kayıtlı video üzerinden çalışma ortamı hazırlanacak. |
| **Şerit/yol algılama** | Basit şerit veya yol merkezi algılama çıktısı hazırlanacak. |
| **Nesne/tabela algılama** | YOLO veya basit görüntü işleme ile kutucuklu algılama çıktısı alınacak. |
| **Algılama veri formatı** | Karar ekibine gönderilecek veriler netleştirilecek. |
| **Demo ekranı** | Ekranda görüntü + etiket + güven skoru + durum bilgisi gösterilecek. |
| **Hata durumları** | Kamera yoksa veya algılama yapılamazsa sistem ne döndürecek belirlenecek. |

**Bu ekipten beklenen çıktı**

| Çıktı | Durum |
| :--- | :--- |
| Kamera görüntüsü alma kodu | Hazır olmalı |
| Şerit/yol algılama demosu | Hazır olmalı |
| Nesne/tabela algılama demosu | Hazır olmalı |
| `lane_error` çıktısı | Hazır olmalı |
| `object_detected` çıktısı | Hazır olmalı |
| `stop_sign` / `traffic_light` benzeri çıktı | Hazır olmalı |

**Karar ekibine gönderecekleri örnek veriler**

| Veri | Örnek |
| :--- | :--- |
| `lane_valid` | `true` / `false` |
| `lane_error` | `-25 px` |
| `object_detected` | `true` / `false` |
| `object_class` | `stop` / `cone` / `pedestrian` |
| `object_confidence` | `0.82` |
| `obstacle_distance` | varsa mesafe |

**Malzeme gelince ilk yapacakları**
1. Araç üstü kamera varsa canlı görüntü alınacak.
2. Kamera yoksa hazır video/simülasyon görüntüsüyle demo gösterilecek.
3. Algılama çıktıları terminale veya arayüze düşürülecek.
4. Otonom karar verme ekibine veri gönderilecek.
5. Video için ekran kaydı alınacak.

**Bu ekip için ana hedef:**
Kamera görüntüsünden anlamlı veri üretip bunu karar sistemine aktarabilir seviyeye gelmek.

---

### 4. Otonom Karar Verme Ekibi

Bu ekip, algılama verilerini sürüş kararına çevirecek. 1 Temmuz için tam otonomi değil, basit ve güvenli karar sistemi yeterli.

**Malzeme gelmeden yapılacaklar**

| Hazırlık | Açıklama |
| :--- | :--- |
| **FSM durumları** | Başlat, kontrol, ilerle, dur, acil dur gibi durumlar belirlenecek. Hangi girişte hangi aksiyon verilecek tanımlanacak. |
| **Karar kuralları** | Hız, direksiyon ve fren komutu üretilecek. |
| **Sahte veri testi** | Algılama olmadan sahte verilerle FSM test edilecek. |
| **Fail-safe kararları** | Veri kaybı, kamera kaybı, emergency durumlarında karar üretilecek. |
| **Terminal logları** | Aktif state ve aksiyon terminalde görünecek. |

**Bu ekipten beklenen çıktı**

| Çıktı | Durum |
| :--- | :--- |
| FSM kod iskeleti | Hazır olmalı |
| Karar tablosu | Hazır olmalı |
| Sahte veriyle çalışan demo | Hazır olmalı |
| `target_speed` üretimi | Hazır olmalı |
| `target_steering` üretimi | Hazır olmalı |
| `brake_cmd` üretimi | Hazır olmalı |
| State log çıktısı | Hazır olmalı |

**Basit FSM önerisi**

| State | Ne yapar? |
| :--- | :--- |
| **INIT** | Sistemi başlatır |
| **SENSOR_CHECK** | Kamera/sensör verisi var mı kontrol eder |
| **MANUAL_READY** | Manuel test komutlarına hazır bekler |
| **LANE_FOLLOW** | Yol/şerit bilgisine göre ilerleme kararı verir |
| **OBSTACLE_CHECK** | Engel var mı kontrol eder |
| **STOP** | Aracı durdurur |
| **EMERGENCY_STOP** | Tüm hareket komutlarını keser |
| **SAFE_MODE** | Hata durumunda güvenli moda geçer |

**Malzeme gelince ilk yapacakları**
1. Önce sahte veriyle kontrol komutu üretmeye devam edecekler.
2. Sonra algılama ekibinden gerçek veri alacaklar.
3. Kontrol ekibine hız, direksiyon ve fren komutu gönderecekler.
4. Entegrasyon ekibiyle birlikte logları kontrol edecekler.
5. Video için state geçişlerini terminalde gösterecekler.

**Bu ekip için ana hedef:**
Araç ne zaman gidecek, ne zaman duracak, ne zaman güvenli moda geçecek kararını üretmek.

---

### Dört Ekibin Birbirine Bağlantısı

Burada ekiplerin birbirinden kopmaması önemli. Bağlantı şu şekilde olmalı:

| Veren Ekip | Alan Ekip | Veri / İş |
| :--- | :--- | :--- |
| **Algılama ve Görüntü İşleme** | Otonom Karar Verme | Şerit, nesne, tabela, engel bilgisi |
| **Otonom Karar Verme** | Kontrol ve Araç Dinamiği | Hedef hız, hedef direksiyon, fren komutu |
| **Kontrol ve Araç Dinamiği** | Entegrasyon ve Test | Motor, direksiyon, fren durum bilgisi |
| **Entegrasyon ve Test** | Tüm ekipler | Test sonucu, hata kaydı, video senaryosu |

**Basit zincir:**
Algılama veri üretir → Otonom karar verir → Kontrol fiziksel komuta çevirir → Entegrasyon test eder ve kaydeder.

---

### Malzeme Gelmeden Hazır Olması Gereken Ortak Dosyalar

Bunları bir klasörde toplayın:

| Dosya / Çıktı | Sorumlu |
| :--- | :--- |
| `can_serial_protocol.md` | Entegrasyon + Kontrol |
| `control_test_code.ino` | Kontrol |
| `send_command.py` | Entegrasyon |
| `state_machine.py` | Otonom Karar |
| `perception_demo.py` | Algılama |
| `test_checklist.md` | Entegrasyon |
| `video_senaryosu.md` | Entegrasyon + Kaptan |
| `logs/` klasörü | Entegrasyon |
| `demo_outputs/` klasörü | Algılama |

---

### Şu Anki En Doğru Görev Özeti

| Ekip | Malzeme Gelmeden Ana Görev | Malzeme Gelince Ana Görev |
| :--- | :--- | :--- |
| **Kontrol ve Araç Dinamiği** | Motor, direksiyon, fren kod iskeletini hazırlamak | Düşük hızda hareket, dönüş, fren ve emergency testi |
| **Entegrasyon ve Test** | Haberleşme protokolü, test scripti, checklist hazırlamak | Sistemi birleştirip video senaryosunu çalıştırmak |
| **Algılama ve Görüntü İşleme** | Kamera/şerit/nesne algılama demosu hazırlamak | Canlı kamera veya demo görüntüsünü karar sistemine bağlamak |
| **Otonom Karar Verme** | FSM ve karar tablosunu hazırlamak | Gerçek/sahte veriden hız, direksiyon ve fren komutu üretmek |