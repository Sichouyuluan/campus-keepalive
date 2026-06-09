; Campus Keepalive Installer
; NSIS 3.10

!include "MUI2.nsh"

; ========== Basic Info ==========
Name "Campus Keepalive"
OutFile "campus-keepalive-setup-v1.1.0.exe"
InstallDir "$PROGRAMFILES\CampusKeepalive"
InstallDirRegKey HKLM "Software\CampusKeepalive" "InstallDir"
RequestExecutionLevel admin

; ========== Version Info ==========
VIProductVersion "1.1.0.0"
VIAddVersionKey "ProductName" "Campus Keepalive"
VIAddVersionKey "CompanyName" "Campus Keepalive"
VIAddVersionKey "FileDescription" "Campus Network Auto Login Tool"
VIAddVersionKey "FileVersion" "1.1.0"
VIAddVersionKey "ProductVersion" "1.1.0"
VIAddVersionKey "LegalCopyright" "Copyright 2026"

; ========== UI Settings ==========
!define MUI_ABORTWARNING
!define MUI_ICON "winres\icon.ico"
!define MUI_UNICON "winres\icon.ico"

; ========== Install Pages ==========
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; ========== Uninstall Pages ==========
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; ========== Language ==========
!insertmacro MUI_LANGUAGE "SimpChinese"

; ========== Install Section ==========
Section "Install Main Program" SecMain
    SetOutPath "$INSTDIR"

    ; Install main program
    File "campus-keepalive.exe"

    ; Install icon
    File /nonfatal "winres\icon.ico"

    ; Write registry
    WriteRegStr HKLM "Software\CampusKeepalive" "InstallDir" "$INSTDIR"

    ; Write uninstall info
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "DisplayName" "Campus Keepalive"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "DisplayIcon" "$\"$INSTDIR\campus-keepalive.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "Publisher" "Campus Keepalive"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "DisplayVersion" "1.1.0"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive" \
        "NoRepair" 1

    ; Create uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

; ========== Desktop Shortcut ==========
Section "Create Desktop Shortcut" SecDesktop
    CreateShortcut "$DESKTOP\Campus Keepalive.lnk" "$INSTDIR\campus-keepalive.exe" "" "$INSTDIR\campus-keepalive.exe" 0
SectionEnd

; ========== Start Menu ==========
Section "Create Start Menu" SecStartMenu
    CreateDirectory "$SMPROGRAMS\Campus Keepalive"
    CreateShortcut "$SMPROGRAMS\Campus Keepalive\Campus Keepalive.lnk" "$INSTDIR\campus-keepalive.exe" "" "$INSTDIR\campus-keepalive.exe" 0
    CreateShortcut "$SMPROGRAMS\Campus Keepalive\Uninstall.lnk" "$INSTDIR\uninstall.exe" "" "$INSTDIR\uninstall.exe" 0
SectionEnd

; ========== Auto Start ==========
Section "Start with Windows" SecAutoStart
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "CampusKeepalive" "$\"$INSTDIR\campus-keepalive.exe$\""
SectionEnd

; ========== Section Descriptions ==========
!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SecMain} "Install Campus Keepalive main program"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} "Create desktop shortcut"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecStartMenu} "Create start menu program group"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecAutoStart} "Start program automatically with Windows"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; ========== Uninstall Section ==========
Section "Uninstall"
    ; Delete program files
    Delete "$INSTDIR\campus-keepalive.exe"
    Delete "$INSTDIR\icon.ico"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"

    ; Delete desktop shortcut
    Delete "$DESKTOP\Campus Keepalive.lnk"

    ; Delete start menu
    RMDir /r "$SMPROGRAMS\Campus Keepalive"

    ; Delete registry
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\CampusKeepalive"
    DeleteRegKey HKLM "Software\CampusKeepalive"
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "CampusKeepalive"
SectionEnd
