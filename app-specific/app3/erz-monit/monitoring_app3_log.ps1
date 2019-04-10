[regex]$pattern = '^(.{28})(.{9})(.{14})(.{2})(.*)'
$arrLines = Get-Content '\\host1\c$\Program Files\ibm\WebSphere\AppServer\profiles\profile_1\logs\server1\SystemOut.log'
$starttime = get-date

Foreach ($strLine in $arrLines)
{
 if ($strLine.StartsWith("["))
  {
    $arrayL = $strLine -split $pattern
    $strdate = $arrayL[1].Remove(0,1) -replace ".{6}$"
    $strdate = $strdate.Trim()
    $dtevent = [datetime]::ParseExact($strdate, "dd.MM.yy H:mm:ss:fff",$null)
    $strevid = $arrayL[2].Trim()
    $strsource = $arrayL[3].Trim()
    $strtype = $arrayL[4].Trim()
    $strmessage = $arrayL[5].Trim()
    $span = New-TimeSpan $dtevent $starttime
    if ($span.TotalHours -le "1")
    {
       if (($strtype -eq "E") -and ($strsource -eq "webapp") -and ($strevid -eq "0000002a") )
        {
            $message = $strmessage -split ":"
            $realerror = $message[2].Trim()
            if ($realerror -eq "java.lang.OutOfMemoryError")
            {
            $wshell = New-Object -ComObject Wscript.Shell
            $wshell.Popup("Требуется проверить память на сервере host1, в логе OutOfMemoryError! Дата события: " + $dtevent)
            $errorstring = "host1 profile_1 " + $strLine
            Out-File -FilePath logz\app3error.log -InputObject $errorstring
            }
        }
    }

  }

 }

$arrLines = Get-Content '\\host1\c$\Program Files\ibm\WebSphere\AppServer\profiles\profile_2\logs\server1\SystemOut.log'
$starttime = get-date

Foreach ($strLine in $arrLines)
{
 if ($strLine.StartsWith("["))
  {
    $arrayL = $strLine -split $pattern
    $strdate = $arrayL[1].Remove(0,1) -replace ".{6}$"
    $strdate = $strdate.Trim()
    $dtevent = [datetime]::ParseExact($strdate, "dd.MM.yy H:mm:ss:fff",$null)
    $strevid = $arrayL[2].Trim()
    $strsource = $arrayL[3].Trim()
    $strtype = $arrayL[4].Trim()
    $strmessage = $arrayL[5].Trim()
    $span = New-TimeSpan $dtevent $starttime
    if ($span.TotalHours -le "1")
    {
       if (($strtype -eq "E") -and ($strsource -eq "webapp") -and ($strevid -eq "0000002a") )
        {
            $message = $strmessage -split ":"
            $realerror = $message[2].Trim()
            if ($realerror -eq "java.lang.OutOfMemoryError")
            {
            $wshell = New-Object -ComObject Wscript.Shell
            $wshell.Popup("Требуется проверить память на сервере host1, в логе OutOfMemoryError! Дата события: " + $dtevent)
            $errorstring = "host1 profile_2 " + $strLine
            Out-File -FilePath logz\app3error.log -InputObject $errorstring
            }
        }
    }

  }

 }


$arrLines = Get-Content '\\host2\c$\Program Files\ibm\WebSphere\AppServer\profiles\profile_1\logs\server1\SystemOut.log'
$starttime = get-date

Foreach ($strLine in $arrLines)
{
 if ($strLine.StartsWith("["))
  {
    $arrayL = $strLine -split $pattern
    $strdate = $arrayL[1].Remove(0,1) -replace ".{6}$"
    $strdate = $strdate.Trim()
    $dtevent = [datetime]::ParseExact($strdate, "dd.MM.yy H:mm:ss:fff",$null)
    $strevid = $arrayL[2].Trim()
    $strsource = $arrayL[3].Trim()
    $strtype = $arrayL[4].Trim()
    $strmessage = $arrayL[5].Trim()
    $span = New-TimeSpan $dtevent $starttime
    if ($span.TotalHours -le "1")
    {
       if (($strtype -eq "E") -and ($strsource -eq "webapp") -and ($strevid -eq "0000002a") )
        {
            $message = $strmessage -split ":"
            $realerror = $message[2].Trim()
            if ($realerror -eq "java.lang.OutOfMemoryError")
            {
            $wshell = New-Object -ComObject Wscript.Shell
            $wshell.Popup("Требуется проверить память на сервере host2, в логе OutOfMemoryError! Дата события: " + $dtevent)
            $errorstring = "host2 profile_1 " + $strLine
            Out-File -FilePath logz\app3error.log -InputObject $errorstring
            }
        }
    }

  }

 }

 $arrLines = Get-Content '\\host2\c$\Program Files\ibm\WebSphere\AppServer\profiles\profile_2\logs\server1\SystemOut.log'
$starttime = get-date

Foreach ($strLine in $arrLines)
{
 if ($strLine.StartsWith("["))
  {
    $arrayL = $strLine -split $pattern
    $strdate = $arrayL[1].Remove(0,1) -replace ".{6}$"
    $strdate = $strdate.Trim()
    $dtevent = [datetime]::ParseExact($strdate, "dd.MM.yy H:mm:ss:fff",$null)
    $strevid = $arrayL[2].Trim()
    $strsource = $arrayL[3].Trim()
    $strtype = $arrayL[4].Trim()
    $strmessage = $arrayL[5].Trim()
    $span = New-TimeSpan $dtevent $starttime
    if ($span.TotalHours -le "1")
    {
       if (($strtype -eq "E") -and ($strsource -eq "webapp") -and ($strevid -eq "0000002a") )
        {
            $message = $strmessage -split ":"
            $realerror = $message[2].Trim()
            if ($realerror -eq "java.lang.OutOfMemoryError")
            {
            $wshell = New-Object -ComObject Wscript.Shell
            $wshell.Popup("Требуется проверить память на сервере host2, в логе OutOfMemoryError! Дата события: " + $dtevent)
            $errorstring = "host2 profile_2 " + $strLine
            Out-File -FilePath logz\app3error.log -InputObject $errorstring
            }
        }
    }

  }

 }
