$word = New-Object -ComObject Word.Application
$word.Visible = $false
$word.DisplayAlerts = 0
$doc = $word.Documents.Open("C:\Users\alara\Desktop\staj\staj-faaliyet-raporu-duzenlenmis.docx")

function ReplaceBySearch($searchStr, $newText) {
    $range = $doc.Content
    $find = $range.Find
    $find.ClearFormatting()
    $find.Text = $searchStr
    $find.Forward = $true; $find.Wrap = 1; $find.Format = $false; $find.MatchCase = $false; $find.MatchWholeWord = $false
    if ($find.Execute()) {
        $range.Expand(4)
        $range.Text = $newText + "`r"
        Write-Host "Updated: $searchStr"
    } else { Write-Host "NOT FOUND: $searchStr" }
}

# Sayfa 1 missed
ReplaceBySearch "Merkezi sunucu" "Şirket bünyesinde çalışanların ağırlıklı olarak dizüstü (laptop) bilgisayarlar kullandığı, bu cihazların HDMI kablolar aracılığıyla harici monitörlere bağlanarak çift ekran yapılandırmasıyla çalışıldığı gözlemlendi. Ağ altyapısında kullanılan switch, router ve güvenlik duvarı cihazlarının temel işlevleri açıklandı."
ReplaceBySearch "ile sunucular" "Eğitimin devamında ofis ortamı tanıtıldı. Personelin kullandığı dizüstü bilgisayarlar, HDMI ile bağlanan harici monitörler, switch, router ve diğer ağ ekipmanları incelendi. Ağ altyapısının genel yapısı gözlemlendi ve staj süresince yürütülecek çalışmalar hakkında bilgilendirme yapıldı."

# Sayfa 2 missed
ReplaceBySearch "Sistemin merkezini" "Analiz kapsamında aşağıdaki bileşenler değerlendirildi:"
ReplaceBySearch "etkileyen termal" "Ayrıca dizüstü bilgisayarların termal yönetim sistemleri incelendi. Isı borusu (heat pipe) teknolojisi, fan yapısı ve termal macunun işlemci üzerindeki rolü gözlemlendi."

# Sayfa 3 missed
ReplaceBySearch "alanlarına uyum" " "

# Sayfa 4
ReplaceBySearch "bilgisayar sisteminin yazılımsal altyapısını kurmak" "Donanım yapılandırması tamamlanan dizüstü bilgisayarın yazılımsal altyapısını kurmak ve BIOS/UEFI (Unified Extensible Firmware Interface) yapılandırmasını yapmak amacıyla çalışmalar gerçekleştirildi. UEFI arayüzü sayesinde fare desteği ve gelişmiş grafikler üzerinden konfigürasyon yapıldı."
ReplaceBySearch "vadedilen hızında" "XMP (Extreme Memory Profile): Yeni takılan RAM modüllerinin fabrika hızında çalışması sağlandı."
ReplaceBySearch "modern işletim sistemiyle tam uyumlu" "AHCI Modu: NVMe SSD'nin modern işletim sistemiyle tam uyumlu çalışması için disk kontrolcüsü ayarlandı."
ReplaceBySearch "monitör aracılığıyla takip" "Önyükleme (Boot) sırası değiştirildi ve önceden UEFI destekli GPT formatında hazırlanan Windows 11 Pro kurulum USB belleği birinci sıraya alınarak sistem harici monitör üzerinden takip edilerek başlatıldı. Windows kurulum ekranında NVMe M.2 SSD tamamen sistem diski (C: sürücüsü) olarak biçimlendirildi."
ReplaceBySearch "ViewSonic monitör desteğiyle bilgisayar" "Kurulum başarıyla tamamlandıktan sonra yerel bir yönetici hesabı açıldı. Dizüstü bilgisayar HDMI ile bağlı harici monitör desteğiyle sorunsuz bir şekilde masaüstü ekranına ulaştı. Aygıt Yöneticisi açılarak tanınmayan (sarı ünlemli) donanımlar listelendi ve sürücü kurulumları ertesi güne planlandı."

# Sayfa 5
ReplaceBySearch "bilgisayarın donanımları ile işletim sistemi arasındaki" "İşletim sistemi kurulumu tamamlanan dizüstü bilgisayarın donanımları ile işletim sistemi arasındaki iletişimi sağlayan sürücülerin (driver) yüklenmesi ve sistemin genel performans optimizasyonunun yapılması işlemleri gerçekleştirildi."
ReplaceBySearch "Ekran kartı üreticisinden indirilen güncel paket" "3. Görüntü Sürücüsü (GPU): Hem dahili (Intel UHD) hem de HDMI üzerinden harici monitöre görüntü aktarımını optimize eden güncel sürücü paketi yüklendi."
ReplaceBySearch "Audio / Intel ME" "4. Ses ve Yönetim (Audio / Touchpad): Multimedya sürücüleri ve dizüstü bilgisayara özgü touchpad sürücüsü kuruldu."
ReplaceBySearch "işlemcinin frekans kısmasına gitmemesi" "Sistemin donanımsal olarak tam uyumlu çalışır hale gelmesinin ardından işletim sistemi bazında çeşitli performans optimizasyonları yapıldı. Dizüstü bilgisayarın pil yönetimi ayarları düzenlenerek, ofiste prize takılı çalışırken 'Yüksek Performans' modunda, pil kullanımında ise 'Dengeli' modda çalışması sağlandı."

# Sayfa 10
ReplaceBySearch "Yavaş İnternet" "2. Senaryo (HDMI Görüntü Aktarım Sorunu): Bir çalışanın dizüstü bilgisayarından HDMI ile bağlı harici monitöre görüntü gelmediği bildirildi. HDMI kablosunun fiziksel bağlantısı kontrol edildi, kablo değiştirilerek sorunun kablodaki kırık pinlerden kaynaklandığı tespit edildi."

# Sayfa 11
ReplaceBySearch "ağın merkezini oluşturan sunucu odasının" "Rewola yazılım şirketinde standart çalışma düzeninin temelini oluşturan HDMI (High-Definition Multimedia Interface) tabanlı görüntü aktarım sistemleri ve çoklu ekran yapılandırmaları detaylı bir şekilde incelendi."
ReplaceBySearch "biyometrik kapı sistemleri" "Şirket bünyesinde çalışanların dizüstü bilgisayarlarını HDMI kablolar aracılığıyla harici masaüstü monitörlere bağlayarak çift ekran (dual monitor) konfigürasyonuyla çalışması standart donanım uygulaması olarak benimsendi."
ReplaceBySearch "42U rack kabininin" " "
ReplaceBySearch "yerleşim şeması" "Şekil 4 - HDMI görüntü aktarım ve bağlantı şeması"
ReplaceBySearch "rack (raf) tipi sistem kabininin standart fiziksel" "HDMI teknolojisinin donanımsal mimarisi ve sinyal iletim prensipleri incelendi. Şirkette kullanılan HDMI 2.0 kablolarının 4K@60Hz çözünürlük ve 18 Gbps bant genişliği desteği sunduğu tespit edildi. Sinyallerin TMDS teknolojisiyle iletildiği öğrenildi."
ReplaceBySearch "Kabin içerisinde çalışan fiziksel sunucuların" "Farklı çalışma istasyonlarında çeşitli çoklu ekran yapılandırmaları test edildi. Windows 'Görüntü Ayarları' üzerinden 'Genişletilmiş Masaüstü' modunun en verimli yapılandırma olduğu değerlendirildi. Ayrıca USB-C to HDMI dönüştürücü adaptörlerin kullanımı da gözlemlendi."

# Sayfa 12
ReplaceBySearch "fiziksel sunucu donanımının işletmeye" "Rewola yazılım şirketinde dizüstü bilgisayar tabanlı çalışma düzeninin verimliliğini artırmak amacıyla kullanılan docking station (yerleştirme istasyonu) cihazlarının donanımsal yapısı, bağlantı protokolleri ve çevre birim entegrasyon süreçleri detaylı olarak incelendi."
ReplaceBySearch "sunucunun anakart arayüzüne" "Şirkette kullanılan USB-C docking station cihazının sunduğu donanımsal genişleme özellikleri şu şekilde analiz edildi:"
ReplaceBySearch "Server Core (Komut Satırı) ile Desktop" "HDMI Çıkış: Docking station üzerindeki HDMI portun dizüstü bilgisayardan gelen görüntü sinyalini harici monitöre aktarma prensibi incelendi. Tek kablo ile hem şarj hem görüntü aktarımı sağlayan USB-C Power Delivery teknolojisi gözlemlendi."
ReplaceBySearch "Server Manager (Sunucu Yöneticisi)" "USB Hub İşlevi: Docking station'ın birden fazla USB 3.0 portu sunarak harici klavye, fare ve diğer çevre birimlerin tek bir merkezi noktadan bağlanmasını sağladığı değerlendirildi."
ReplaceBySearch "Hostname (Bilgisayar Adı)" "Ethernet Portu: Kablosuz bağlantı yerine daha kararlı ve hızlı Gigabit Ethernet bağlantısı sunan dahili ağ adaptörünün yapısı incelendi."
ReplaceBySearch "Ağ Yapılandırması: Sunucularda dinamik IP" "Yeni bir çalışma istasyonu için docking station kurulumu uygulamalı olarak gerçekleştirildi."
ReplaceBySearch "Uzak Yönetim: IT personelinin müdahalesi" "Dizüstü bilgisayar USB-C kablosuyla docking station'a bağlandığında otomatik olarak harici monitör (HDMI), klavye, fare ve Ethernet bağlantısının tek seferde aktif olduğu gözlemlendi."
ReplaceBySearch "Active Directory ve dosya paylaşımı gibi ek kurumsal rollerin" "Bu yapılandırma sayesinde personelin sabah geldiğinde tek bir kablo takarak tüm çevre birimlerine anında erişebildiği ve akşam tek bir kablo çıkararak dizüstü bilgisayarını alıp gidebildiği verimli bir çalışma düzeni sağlandı."
ReplaceBySearch "Güvenlik Duvarı & Güncelleme" " "

$doc.Save(); $doc.Close($false); $word.Quit(); Write-Host "Done"