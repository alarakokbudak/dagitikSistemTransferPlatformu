# Robotaksi URDF ve Simülasyon Kurulum Rehberi

Bu klasör, Teknofest Robotaksi projesi için hazırlanan aracın 3 boyutlu model (Mesh) dosyalarını ve ROS 2 URDF/XACRO simülasyon tanımlamalarını içerir.

## 🛠️ Yapılan Optimizasyonlar ve Düzeltmeler

Onshape üzerinden dışa aktarılan ham STL dosyalarında aşağıdaki sorunlar tespit edilmiş ve düzeltilmiştir:
1. **Origin (Merkez) Kaymaları:** Tüm parçaların (şasi ve tekerlekler) kendi merkezleri uzayda rastgele konumlardaydı. Tüm parçalar kendi geometrik Bounding Box merkezlerine oturtuldu ve ROS koordinat sistemine (X: İleri, Y: Sol, Z: Yukarı) göre hizalandı.
2. **Birleşik Tekerleklerin Ayrılması:** `on_teker.stl` ve `arka_teker.stl` dosyalarının her biri birer "aks" gibi ikişer tekerlek içeriyordu. Bu tekerlekler Python (Trimesh & Open3D) kullanılarak sağ ve sol olmak üzere 4 bağımsız tekerleğe (`fl_wheel`, `fr_wheel`, `rl_wheel`, `rr_wheel`) ayrıştırıldı.
3. **Fren ve Kaliper Asimetrisi:** Tekerleklerin iç kısmında kalan fren ve süspansiyon detayları asimetrik bir Bounding Box oluşturuyordu. Bu nedenle tekerleklerin `robotaksi.urdf.xacro` içindeki montaj koordinatları, simetrik varsayımlarla değil; tamamen orijinal CAD modelindeki milimetrik ofset vektörleriyle (Örn: `X=0.7727, Y=0.4726, Z=-0.4239`) yerleştirildi.
4. **Decimation (Poligon Düşürme):** Orijinal `on_teker.stl` dosyasındaki **553.000 poligonluk** gereksiz yüksek detay seviyesi (özellikle fren kaliperi ve cıvatalar), Open3D Quadric Decimation algoritmasıyla **15.000 poligona** düşürüldü. Bu işlem sayesinde RViz ve Gazebo'nun çökmesi/donması engellendi.

## 🚀 Sistemi Ayağa Kaldırma (Kullanım Talimatları)

Modelin ROS 2 Jazzy ortamında, RViz veya Gazebo Harmonic üzerinde sorunsuz şekilde çalışması için aşağıdaki adımları sırasıyla uygulayınız.

### Adım 1: RViz'de Görüntüleme (TF Ağacı Testi)

Her bir komutu ayrı bir terminalde, çalışma alanınızın kök dizininde (`~/robotaksi_ws`) çalıştırın:

**Terminal 1 (URDF'i yayınla):**
```bash
# XACRO'yu URDF'e derle
xacro src/robot_description/urdf/robotaksi.urdf.xacro > src/robot_description/urdf/robotaksi.urdf

# Robot State Publisher'ı başlat
ros2 run robot_state_publisher robot_state_publisher src/robot_description/urdf/robotaksi.urdf
```

**Terminal 2 (Tekerlek mafsallarını tetikle):**
```bash
ros2 run joint_state_publisher_gui joint_state_publisher_gui src/robot_description/urdf/robotaksi.urdf
```

**Terminal 3 (RViz2 Başlat):**
```bash
rviz2
```
*RViz açıldığında:*
- `Global Options -> Fixed Frame` kısmını **base_link** yapın.
- Sol alt köşeden **Add -> RobotModel** ekleyin.
- RobotModel menüsünü genişletip **Description Topic** alanına manuel olarak `/robot_description` yazıp Enter'a basın.

### Adım 2: Gazebo Harmonic Üzerinde Başlatma

Eğer aracı fizik motoruyla birlikte Gazebo'da (ROS 2 Jazzy varsayılanı) görmek isterseniz, Adım 1'deki `robot_state_publisher` düğümünün açık olduğundan (Terminal 1) emin olun ve aşağıdaki komutları çalıştırın:

**Terminal 4 (Boş bir Gazebo dünyası başlat):**
```bash
gz sim empty.sdf
```

**Terminal 5 (Aracı Gazebo'ya aktar):**
```bash
ros2 run ros_gz_sim create -topic robot_description -name robotaksi -z 0.5
```
*Bu komut, yayınlanan URDF dosyasını yakalar ve aracı Gazebo ortamında Z ekseninde 0.5 metre havadan bırakır.*

---
**Not:** Otonomi testlerine (FSM ve Navigasyon) geçmeden önce aracı hareket ettirmek isterseniz, `robotaksi.urdf.xacro` dosyasına Gazebo motor (Ackermann veya Diff-Drive) eklentilerini (plugins) dahil etmeniz gerekecektir.
