# Robotaksi Slalom / Engelden Kaçınma (Obstacle Avoidance) Dokümantasyonu

Bu doküman, Teknofest 2026 Robotaksi yarışması için geliştirilen **3 Aşamalı Dinamik Sollama Manevrası** mimarisinin çalışma mantığını, LiDAR algılama stratejisini ve güvenlik mekanizmalarını açıklar.

> **Son Güncelleme:** 25 Haziran 2026  
> **Dosyalar:**  
> - Beyin (FSM): `src/robotaksi_autonomy/robotaksi_autonomy/fsm_decision_node.py`  
> - Gözler (Perception): `src/robotaksi_autonomy/robotaksi_autonomy/perception_node.py`

---

## 1. Genel Bakış

Araç, önündeki statik engelleri (kutu, duba vb.) algıladığında **3 aşamalı yarı-kapalı çevrim** bir sollama manevrası uygular. Eski sinüs tabanlı açık-çevrim sisteme göre en büyük farkı: aracın sol şeritte ilerlerken **PID şerit takibini** kullanması ve sağa dönüş kararını **LiDAR verisine dayalı olarak dinamik** vermesidir.

### Mimari Avantajları
| Özellik | Eski Sistem (Sinüs/5-Faz) | Yeni Sistem (3-Aşama Dinamik) |
|---|---|---|
| Şerit değiştirme | Kör (zamana dayalı) | Yarı-akıllı (PID + zaman) |
| Engel geçiş süresi | Sabit (ör. 1.5 sn) | Dinamik (sağ şerit boşalana kadar) |
| Çoklu engel desteği | ❌ (tek engel için tasarlandı) | ✅ (her engel bağımsız ele alınır) |
| Şeride dönüş kararı | Zamanlayıcı | LiDAR Kartezyen tarama |
| Ardışık sollama güvenliği | ❌ | ✅ (4 sn soğuma süresi) |

---

## 2. LiDAR Algılama Stratejisi (Perception Node)

LiDAR verileri polar koordinatlardan (r, θ) **Kartezyen koordinatlara** (x, y) dönüştürülür. Bu sayede şerit bazlı engel tespiti yapılır.

```
         x (ileri)
         ↑
         |
    y ←──┼──→ -y
    (sol) |  (sağ)
         |
      [ARAÇ]
```

### 2.1 Tarama Bölgeleri

| Bölge | X Aralığı | Y Aralığı | Amaç |
|---|---|---|---|
| **Ön Engel** (Kendi Şeridi) | 0 → 30m | -1.0 → 1.0m | Önümüzdeki en yakın engelin mesafesini ölç |
| **Sol Şerit** | 0 → 30m | 1.2 → 3.6m | Sollamaya çıkmadan önce sol şeridin boş olup olmadığını kontrol et |
| **Sağ Şerit** (Yakın Alan) | -4.0 → **8.0m** | -4.5 → -0.8m | Engeli geçtikten sonra sağa dönmenin güvenli olup olmadığını kontrol et |

> **Kritik Tasarım Kararı:** Sağ şerit taraması bilerek **sadece 8 metreye** sınırlandırılmıştır.  
> **Neden?** Uzağa (30m) bakıldığında, 25m ilerideki ikinci bir engel de "sağ şerit dolu" olarak algılanır ve araç sağa hiç dönemez. 8 metrelik kısa menzil sayesinde **her engel bağımsız** bir sollama manevrası olarak ele alınır:
> ```
> Engel 1'i geç → Yanımda engel kalmadı (8m içinde) → Sağa dön → CRUISE
>   → Engel 2'ye yaklaş → ACC → Yeni sollama başlat → ...
> ```

### 2.2 Yayınlanan Topic'ler

| Topic | Tip | Açıklama |
|---|---|---|
| `/perception/obstacle_dist` | Float32 | Ön engel mesafesi (metre, EMA filtreli) |
| `/perception/left_lane_clear` | Bool | Sol şerit boş mu? |
| `/perception/right_lane_clear` | Bool | Sağ şerit (yakın alan) boş mu? |

### 2.3 Gürültü Filtreleme
- **Şasi Filtresi:** `r < 1.0m` olan LiDAR okumaları filtrelenir (aracın kendi tekerlekleri/şasisi).
- **EMA Filtresi:** Engel mesafesi Exponential Moving Average ile yumuşatılır (alpha=0.4).

---

## 3. FSM Durum Makinesi — 3 Aşamalı Sollama

### 3.1 Durum Geçiş Diyagramı

```
CRUISE ──(engel ≤ 8m & sol boş & soğuma OK)──→ OVERTAKING_ENTER
                                                      │
                                                      │ (5.0 sn geçti)
                                                      ▼
                                                OVERTAKING_PASS
                                                      │
                                         (sağ şerit 0.5sn kesintisiz boş)
                                                      │
                                                      ▼
                                                OVERTAKING_EXIT
                                                      │
                                                      │ (5.0 sn geçti)
                                                      ▼
                                                   CRUISE
                                              (4 sn soğuma başlar)
```

### 3.2 Aşama Detayları

#### Aşama 1: `OVERTAKING_ENTER` — Sola Çıkış
- **Süre:** 5.0 saniye (parametre: `phase1_dur`)
- **Hız:** 1.0 m/s (parametre: `maneuver_speed`)
- **Direksiyon:** +0.15 rad sabit sola (parametre: `maneuver_steer_angle`)
- **Kontrol Tipi:** Açık çevrim (open-loop)
- **Açıklama:** Araç yumuşak bir kavis çizerek sol şeride geçer. Küçük direksiyon açısı (0.15 rad) kameranın şerit çizgilerini kaybetmesini önler.

#### Aşama 2: `OVERTAKING_PASS` — Sol Şeritte İlerleme
- **Süre:** ♾️ Sınırsız (sağ şerit boşalana kadar)
- **Hız:** 1.0 m/s
- **Direksiyon:** **PID kontrolcü** ile şerit takibi
- **Kontrol Tipi:** Kapalı çevrim (closed-loop, kamera tabanlı)
- **Çıkış Koşulu:** `/perception/right_lane_clear` sinyali **kesintisiz 0.5 saniye** boyunca `True` olduğunda
- **Açıklama:** Bu aşamanın zaman sınırı yoktur! Araç sol şeritte PID ile şerit çizgilerini takip ederek dümdüz gider. Sağ şerit taraması (8m yakın alan) sürekli kontrol edilir:
  - Engel hâlâ yanındaysa → Sol şeritte kalmaya devam eder
  - Engel geride kaldıysa → 0.5 sn debounce sonrası sağa dönüşe geçer

> **Debounce Filtresi (0.5 sn):** LiDAR gürültüsünden kaynaklanan anlık "boş" okumalarının erken dönüşe yol açmasını engeller. Sağ şerit kesintisiz 0.5 saniye boş kalmalıdır.

#### Aşama 3: `OVERTAKING_EXIT` — Sağa Dönüş
- **Süre:** 5.0 saniye (parametre: `phase3_dur`)
- **Hız:** 1.0 m/s
- **Direksiyon:** -0.15 rad sabit sağa
- **Kontrol Tipi:** Açık çevrim (open-loop)
- **Açıklama:** Araç simetrik bir kavis çizerek kendi şeridine geri döner. Manevra bittiğinde PID sıfırlanır.

---

## 4. Güvenlik Mekanizmaları

### 4.1 Soğuma Zamanlayıcısı (Cooldown — 4 saniye)
Ardışık engellerde araç her sollama sonrası biraz daha sola kayıyordu (PID henüz aracı ortalayamadan yeni manevra başlıyordu → kümülatif sapma → yoldan çıkma).

**Çözüm:** Sollama tamamlandıktan sonra **4 saniye boyunca yeni sollama başlatılamaz.** Bu sürede:
- PID aracı şeridin ortasına hizalar
- Engele yaklaşılırsa `ACC_FOLLOW` moduna geçilir (yavaşlar ama sola çıkmaz)
- 4 saniye dolduktan sonra güvenle yeni sollama başlatılabilir

### 4.2 Sol Şerit Kontrolü
Sollama başlamadan önce sol şerit 30 metre ileriye kadar taranır. Sol şeritte herhangi bir engel varsa, araç sollama yapmaz ve `ACC_FOLLOW` modunda yavaşlayarak bekler.

### 4.3 ACC (Adaptive Cruise Control)
Engele yaklaşırken hız, mesafeyle doğru orantılı olarak düşürülür:
```python
speed_factor = (obstacle_dist - critical_dist) / (safe_dist - critical_dist)
speed = max(0.2, max_speed * speed_factor)
```

### 4.4 Acil Fren (Critical Stop)
Engel mesafesi 1.5 metrenin altına düşerse araç derhal durur.

### 4.5 Watchdog (Sensör Zaman Aşımı)
Sensör verisi 2.0 saniye boyunca gelmezse araç `FAIL_SAFE` moduna geçerek durur.

---

## 5. Parametre Tablosu

Tüm parametreler ROS 2 üzerinden **çalışma anında** (`ros2 param set`) değiştirilebilir.

| Parametre | Varsayılan | Açıklama |
|---|---|---|
| `safe_dist` | 8.0 m | Sollama tetikleme mesafesi |
| `critical_dist` | 1.5 m | Acil fren mesafesi |
| `maneuver_speed` | 1.0 m/s | Manevra hızı |
| `maneuver_steer_angle` | 0.15 rad | Manevra direksiyon açısı |
| `phase1_dur` | 5.0 sn | Sola çıkış süresi |
| `phase3_dur` | 5.0 sn | Sağa dönüş süresi |
| `max_speed` | 1.5 m/s | Normal seyir hızı |
| `pid_kp` | 1.2 | PID Oransal kazanç |
| `pid_ki` | 0.01 | PID İntegral kazanç |
| `pid_kd` | 0.3 | PID Türevsel kazanç |

---

## 6. Test Senaryoları ve Sonuçları

### Senaryo A: Tekli Engel ✅
Tek bir kutu, sağ şeritte. Araç 8m kala sola çıkar, engeli PID ile geçer, engel geride kalınca sağa döner.

### Senaryo B: Çoklu Ayrık Engeller ✅
Birbirinden 15-20m uzaklıkta birden fazla kutu. Her biri bağımsız bir sollama manevrası olarak ele alınır. Soğuma zamanlayıcısı (4 sn) sayesinde araç her manevra arasında düzgünce ortalanır.

### Senaryo C: Ardışık (Bitişik) Engeller ✅
Art arda dizilmiş kutular (aralarında <8m boşluk). Araç sağ şerit yakın alan taramasında (8m) sürekli engel gördüğü için sol şeritte kalmaya devam eder. Tüm kutular geride kalınca (8m içinde engel yok + 0.5 sn debounce) sağa döner.

### Senaryo D: İki Şerit Dolu ✅
Her iki şeritte de engel varsa, araç sol şerit kontrolünden (`left_lane_clear = False`) sollama yapmaz ve `ACC_FOLLOW` modunda yavaşlayarak engelin arkasında durur.

---

## 7. Geçmiş Sürümler (Arşiv)

### v1 — Sinüs Tabanlı Açık Çevrim (Yedeklendi)
`math.sin()` fonksiyonu ile parabolik S-eğrisi çizen 3 fazlı kör manevra. Tek engel için çalışıyordu ancak PID entegrasyonu yoktu, çoklu engel desteği yoktu.

### v2 — 5 Fazlı Açık Çevrim
Sola kır → Düzle → Düz git → Sağa kır → Düzle şeklinde 5 adımlı zamana dayalı manevra. Gazebo fizik motorundaki kayma (understeer) ve zamanlama hataları biriktiği için overshoot ve yoldan çıkma sorunları yaşanıyordu.

### v3 — 3 Aşamalı Dinamik (Güncel ✅)
Bu dokümanda anlatılan sistem. PID entegrasyonu, Kartezyen LiDAR tarama, debounce filtresi ve soğuma zamanlayıcısı ile en stabil ve güvenilir sürüm.
