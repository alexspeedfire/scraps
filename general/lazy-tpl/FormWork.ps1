# Итак, план
# Понимаем какой день недели
# Забиваем ворк-дей на следующий четверг
# Делаем замену в шаблоне
# Сохраняем его в нужную папку под нужным именем
# Печатаем его.

$strSourceRTFTpl = "template\template.rtf"
$datePlannedDay = (get-date).AddDays(11 - (Get-Date).DayOfWeek.value__)
$strPlannedDay = $datePlannedDay.ToString("dd.MM.yyyy")
$strCurrYear = get-date -Format yyyy
$strOutputPlan = $env:userprofile + "\Documents\Рабочий архив\Планы работ\" + $datePlannedDay.ToString("yyyyMMdd") + "_план работ.rtf"
Get-Content $strSourceRTFTpl | ForEach-Object {$_ -replace 'curryear', $strCurrYear} | ForEach-Object {$_ -replace 'workdate', $strPlannedDay} | Set-Content $strOutputPlan
Start-Process -FilePath $strOutputPlan -Verb Print

[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")

$objNotifyIcon = New-Object System.Windows.Forms.NotifyIcon

$objNotifyIcon.Icon = "C:\Windows\Microsoft.NET\Framework64\v4.0.30319\SetupCache\v4.5.51209\RUS\Graphics\warn.ico"
$objNotifyIcon.BalloonTipIcon = "Warning"
$objNotifyIcon.BalloonTipText = "На принтер отправлен новый план работ. Сходи и подпиши!"
$objNotifyIcon.BalloonTipTitle = "Подготовлен план работ"

$objNotifyIcon.Visible = $True
$objNotifyIcon.ShowBalloonTip(10000)