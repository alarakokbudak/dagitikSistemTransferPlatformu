# 🚀 1. Hafta PID ve Algı Geliştirme Raporu

**Tarih:** 20 Haziran 2026
**Geliştirilen Bileşenler:** PID Kontrolcüsü Kalibrasyon Ortamları, Perception Node (Algı) Optimizasyonu

Bu rapor, Teknofest Robotaksi yarışmasının 1. Hafta "Araç Şerit Takibi" hedeflerini tamamlamak ve sistemin zorlu senaryolarda (kesikli şeritler, engeller) gösterdiği yalpalamaları çözmek için yapılan iyileştirmeleri ve test ortamlarını belgelendirmektedir.

---

## 🛠️ Neler Yaptık? (Yapılan Güncellemeler)

### 1. Özel Test Haritaları (World Dosyaları) Oluşturuldu
Aracın PID kontrolcüsünü virajlı pistte ayarlamak çok zordur. Bu nedenle kalibrasyon ve test süreçlerini izole etmek için 3 yeni özel Gazebo haritası oluşturuldu:
*   **`duz_yol.world`**: Hiçbir virajı ve engeli olmayan, 120 metre uzunluğunda dümdüz bir şerit. (Amacı: P ve D katsayılarını pürüzsüz ayarlamak).
*   **`cift_serit.world`**: İki şeritten oluşan ve aracın şeridinde (sağ şeritte) 35. metrede bekleyen statik bir kırmızı engel bulunan harita. (Amacı: FSM'in LiDAR engel tespiti ve sol şeride geçiş kararını test etmek).
*   **`cift_serit_kesikli.world`**: Çift şeritli haritanın daha gerçekçi versiyonudur. İki şeridi ayıran orta çizgi kesiklidir (3m dolu, 3m boşluk).

### 2. Algı (Perception) Düğümündeki "Kör Nokta" Yalpalaması Çözüldü
Aracın kesikli şeritlerde giderken çizgi boşluklarına denk geldiğinde aniden sola/sağa kırması (yalpalama) hatası tespit edildi ve çözüldü.
*   **Sorun:** Kesikli çizgi kameranın açısından çıktığında, sistem o bölgede beyaz çizgi bulmak için yan şeridin (en soldaki) çizgisine odaklanıyordu. Bu da şeridin aniden 2 kat genişlediğini sanmasına ve hedefin sapmasına yol açıyordu.
*   **Çözüm:** `perception_node.py` dosyasına **Dinamik Şerit Genişliği Hafızası (`tracked_lane_width_px`)** eklendi. Araç artık iki çizgiyi düzgün görürken şeridin pixel genişliğini ezberliyor. Şeritlerden biri kaybolduğunda ve ekranda çok uzakta sahte bir şerit bulunduğunda genişlik filtremiz (`current_width > tracked_lane_width_px * 1.4`) bunu fark edip sahte çizgiyi iptal ediyor ve aracı eski hafızasındaki genişlikle merkezde tutmaya devam ediyor.

### 3. Debug Kamerasında "Kaybolan Sarı Çizgi" Sorunu Giderildi
*   Aracın hedefini gösteren Sarı Çizgi, şeritlerden sadece biri kaybolduğunda bile ekrandan siliniyordu (araç arka planda hafızadan doğru gitse de görsel olarak kayboluyordu).
*   Kod güncellenerek, şerit çizgileri tamamen görünmez olsa bile sarı çizginin (hedef rotanın) sistem hafızasından okunarak ekrana kesintisiz çizilmesi sağlandı.

---

## 🏃‍♂️ Sistemi Çalıştırmak İsteyenler İçin Rehber

Projeyi ilk kez indirecek veya bu test haritalarında deneme yapmak isteyecek takım arkadaşları için adım adım kullanım rehberi:

### Adım 1: Çalışma Alanını (Workspace) Derleme
Kodlarda Python güncellemesi yapıldığı için (perception_node.py) sistemin bir kez derlenmesi gerekir:
```bash
cd ~/Documents/GitHub/Tekonofest_OtonomveKaraVerme
colcon build --packages-select robotaksi_autonomy
```

### Adım 2: Simülasyonu Başlatma (Terminal 1)
Test etmek istediğiniz haritayı belirleyip `ros2 launch` komutunu çalıştırın. `setup.bash` dosyasını `source` yapmayı unutmayın!

*Dümdüz PID test pisti için:*
```bash
source install/setup.bash
ros2 launch robotaksi_autonomy robotaksi_full.launch.py world:=$(pwd)/src/robotaksi_sim/worlds/duz_yol.world
```

*Kesikli çift şerit ve Engel testi için:*
```bash
source install/setup.bash
ros2 launch robotaksi_autonomy robotaksi_full.launch.py world:=$(pwd)/src/robotaksi_sim/worlds/cift_serit_kesikli.world
```

### Adım 3: Aracın Gözünden Bakma (Terminal 2)
Aracın şeritleri ve rotayı nasıl hesapladığını (mavi, yeşil, kırmızı ve sarı çizgileri) görmek için yeni bir terminal açın:
```bash
source ~/Documents/GitHub/Tekonofest_OtonomveKaraVerme/install/setup.bash
ros2 run rqt_image_view rqt_image_view /perception/debug_image
```
*(Sarı çizgi hedeftir, beyaz dikey çizgi ise aracın merkezidir. Sarı çizgiyle beyaz çizginin üst üste bindiğinden emin olun.)*

### Adım 4: Canlı PID Ayarı Yapma (Terminal 3)
Araç şeritte giderken yalpalamasını durdurmak için simülasyonu kapatmanıza gerek yoktur. Başka bir terminal açarak `ros2 param set` ile PID katsayılarını anında değiştirebilirsiniz:

```bash
source ~/Documents/GitHub/Tekonofest_OtonomveKaraVerme/install/setup.bash

# Eğer araç çok hızlı sağ-sol yapıyorsa P değerini düşürün:
ros2 param set /fsm_decision_node pid_kp 0.8

# Direksiyon daha yumuşak toplasın istiyorsanız D değerini artırın:
ros2 param set /fsm_decision_node pid_kd 0.4

# Tüm FSM parametrelerini (hız dahil) görmek için:
ros2 param list /fsm_decision_node
```

Bu adımları takip ederek aracın PID kalibrasyonunu tamamlayabilir ve şerit değiştirme algoritmalarınızı güvenle test edebilirsiniz!
