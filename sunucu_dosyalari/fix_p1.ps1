$word = New-Object -ComObject Word.Application
$word.Visible = $false
$doc = $word.Documents.Open("C:\Users\alara\Desktop\staj\staj-faaliyet-raporu-duzenlenmis.docx")

function SetPara($idx, $text) {
    $doc.Paragraphs.Item($idx).Range.Text = $text + "`r"
}

# Sayfa 1 - P212, P213, P215
SetPara 212 "Bilgi teknolojileri altyapısı hakkında genel tanıtım yapıldı. Şirket bünyesinde çalışanların ağırlıklı olarak dizüstü (laptop) bilgisayarlar kullandığı, bu cihazların HDMI kablolar aracılığıyla harici monitörlere bağlanarak çift ekran yapılandırmasıyla çalışıldığı gözlemlendi. Ağ altyapısında kullanılan switch, router ve güvenlik duvarı cihazlarının temel işlevleri açıklandı."
SetPara 213 "Elektronik donanımlarla çalışırken uyulması gereken elektrostatik deşarj (ESD) kuralları ayrıntılı olarak anlatıldı. Donanım işlemlerinde antistatik ekipman kullanımının önemi ve çalışma ortamında alınması gereken güvenlik önlemleri açıklandı."
SetPara 215 "Eğitimin devamında ofis ortamı tanıtıldı. Personelin kullandığı dizüstü bilgisayarlar, HDMI ile bağlanan harici monitörler, switch, router ve diğer ağ ekipmanları incelendi. Ağ altyapısının genel yapısı gözlemlendi ve staj süresince yürütülecek çalışmalar hakkında bilgilendirme yapıldı."

Write-Host "Sayfa 1 OK"
$doc.Save()
$doc.Close($false)
$word.Quit()
Write-Host "Done"