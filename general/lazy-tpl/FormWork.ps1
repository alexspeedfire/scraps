# ����, ����
# �������� ����� ���� ������
# �������� ����-��� �� ��������� �������
# ������ ������ � �������
# ��������� ��� � ������ ����� ��� ������ ������
# �������� ���.

$strSourceRTFTpl = "template\template.rtf"
$datePlannedDay = (get-date).AddDays(11 - (Get-Date).DayOfWeek.value__)
$strPlannedDay = $datePlannedDay.ToString("dd.MM.yyyy")
$strCurrYear = get-date -Format yyyy
$strOutputPlan = $env:userprofile + "\Documents\������� �����\����� �����\" + $datePlannedDay.ToString("yyyyMMdd") + "_���� �����.rtf"
Get-Content $strSourceRTFTpl | ForEach-Object {$_ -replace 'curryear', $strCurrYear} | ForEach-Object {$_ -replace 'workdate', $strPlannedDay} | Set-Content $strOutputPlan
Start-Process -FilePath $strOutputPlan -Verb Print

[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")

$objNotifyIcon = New-Object System.Windows.Forms.NotifyIcon

$objNotifyIcon.Icon = "C:\Windows\Microsoft.NET\Framework64\v4.0.30319\SetupCache\v4.5.51209\RUS\Graphics\warn.ico"
$objNotifyIcon.BalloonTipIcon = "Warning"
$objNotifyIcon.BalloonTipText = "�� ������� ��������� ����� ���� �����. ����� � �������!"
$objNotifyIcon.BalloonTipTitle = "����������� ���� �����"

$objNotifyIcon.Visible = $True
$objNotifyIcon.ShowBalloonTip(10000)