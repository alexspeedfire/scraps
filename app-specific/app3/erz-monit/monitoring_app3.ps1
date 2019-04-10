$strPass = ConvertTo-SecureString -String "PASSWORD" -AsPlainText -Force
$strUser = "USER"
$pscredCred = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $strUser, $strPass
$listHosts = @('host1', 'host2')


function ReportMe([string] $reporthost, [string] $message)
{
$Global:vncHost = $reporthost
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$objNotifyIcon = New-Object System.Windows.Forms.NotifyIcon
$objNotifyIcon.Icon = "C:\Windows\Microsoft.NET\Framework64\v4.0.30319\SetupCache\v4.5.51209\RUS\Graphics\warn.ico"
$objNotifyIcon.BalloonTipIcon = "Warning"
$objNotifyIcon.BalloonTipText = $message
$objNotifyIcon.BalloonTipTitle = "Сообщение от скрипта"
$objNotifyIcon.Visible = $True
$objNotifyIcon.ShowBalloonTip(30000)


Unregister-Event -SourceIdentifier click_event -ErrorAction SilentlyContinue
Register-ObjectEvent $objNotifyIcon BalloonTipClicked -SourceIdentifier click_event -Action {

$runchmd = "C:\Program Files\uvnc bvba\UltraVNC\vncviewer.exe";
[array]$argz = $vncHost, "-password", "PASSWORD2";
& $runchmd $argz
} | Out-Null

Wait-Event -Timeout 15 -SourceIdentifier click_event > $null
Remove-Event click_event -ErrorAction SilentlyContinue
Unregister-Event -SourceIdentifier click_event -ErrorAction SilentlyContinue
$objNotifyIcon.Dispose()
}

foreach ($strHost in $listHosts)
{
    if (Test-Connection -ComputerName $strHost -BufferSize 16 -Count 1 -ErrorAction 0 -Quiet)
    {
        $listProcess = Get-Process -ComputerName $strHost -Name java
        foreach ($strProcess in $listProcess)
        {
           
            if ([int64] $strProcess.WorkingSet64 -ge 2684354560) 
                { 
                    $strReport = "Обнаружено повышенное потребление памяти процессом c id " + $strProcess.id + " на сервере " + $strHost + ". Текущее потребление " + $strProcess.WorkingSet64 /1MB + " MB."
                    ReportMe -reporthost $strHost -message $strReport
                }
        }
    }
    else
    {
        $strReport = "Нарушена связь с "+$strHost
        ReportMe -message $strReport -reporthost $strHost
    }


}
