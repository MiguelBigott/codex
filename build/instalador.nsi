Unicode true
SetCompressor /SOLID lzma
!include "MUI2.nsh"
!include "FileFunc.nsh"
!include "x64.nsh"
!include "LogicLib.nsh"

!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP "Codex"
!define EXE "Codex.exe"
!define REGAPP "Software\Codex"
!define UNINST "Software\Microsoft\Windows\CurrentVersion\Uninstall\Codex"

Name "${APP}"
OutFile "..\dist\Codex-Setup-${VERSION}.exe"
; Todo vive en C:\Program Files\Codex: el launcher aquí y cada app en su subcarpeta.
InstallDir "$PROGRAMFILES64\Codex"
RequestExecutionLevel admin
ManifestDPIAware true
BrandingText "${APP} ${VERSION}"
VIProductVersion "${VERSION}.0"
VIAddVersionKey /LANG=1034 "ProductName" "${APP}"
VIAddVersionKey /LANG=1034 "FileDescription" "Instalador de ${APP}"
VIAddVersionKey /LANG=1034 "FileVersion" "${VERSION}"
VIAddVersionKey /LANG=1034 "ProductVersion" "${VERSION}"
VIAddVersionKey /LANG=1034 "LegalCopyright" "© 2026 Miguel Bigott"
VIAddVersionKey /LANG=1034 "CompanyName" "Miguel Bigott"

!define MUI_ICON "app.ico"
!define MUI_UNICON "app.ico"
!define MUI_WELCOMEFINISHPAGE_BITMAP "lateral.bmp"
!define MUI_WELCOMEFINISHPAGE_BITMAP_STRETCH AspectFitHeight
!define MUI_UNWELCOMEFINISHPAGE_BITMAP "lateral.bmp"
!define MUI_UNWELCOMEFINISHPAGE_BITMAP_STRETCH AspectFitHeight
!define MUI_ABORTWARNING

!define MUI_WELCOMEPAGE_TITLE "Instalar ${APP}"
!define MUI_WELCOMEPAGE_TEXT "El launcher instala, actualiza y abre las apps de Codex.$\r$\n$\r$\nSe instala en C:\Program Files\Codex, donde también quedarán las apps que instales desde él."
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "Abrir ${APP}"
!define MUI_FINISHPAGE_RUN_FUNCTION AbrirLauncher
!define MUI_FINISHPAGE_SHOWREADME
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Crear acceso directo en el escritorio"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION CrearAccesoEscritorio
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "Spanish"

Function .onInit
  ${If} ${RunningX64}
    SetRegView 64
  ${EndIf}
  SetShellVarContext all
FunctionEnd

Function un.onInit
  ${If} ${RunningX64}
    SetRegView 64
  ${EndIf}
  SetShellVarContext all
FunctionEnd

; El instalador corre como administrador: el launcher se abre a través del Explorador
; para que arranque con los permisos normales del usuario.
Function AbrirLauncher
  Exec '"$WINDIR\explorer.exe" "$INSTDIR\${EXE}"'
FunctionEnd

; En el escritorio del usuario (no en el público), para que pueda borrarlo sin permisos.
Function CrearAccesoEscritorio
  SetShellVarContext current
  CreateShortcut "$DESKTOP\${APP}.lnk" "$INSTDIR\${EXE}"
  SetShellVarContext all
FunctionEnd

Section "Instalar"
  ; Cierra el launcher si está abierto para poder reemplazarlo.
  nsExec::Exec 'taskkill /IM "${EXE}" /F'
  Sleep 500

  SetOutPath "$INSTDIR"
  File "..\dist\${EXE}"
  ; Licencia MIT de la app y las de terceros que van dentro del .exe.
  File "..\dist\LICENSE.txt"
  File "..\dist\LICENCIAS-TERCEROS.txt"
  WriteUninstaller "$INSTDIR\Desinstalar Codex.exe"
  CreateShortcut "$SMPROGRAMS\${APP}.lnk" "$INSTDIR\${EXE}"

  WriteRegStr HKLM "${REGAPP}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "${UNINST}" "DisplayName" "${APP}"
  WriteRegStr HKLM "${UNINST}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "${UNINST}" "Publisher" "Miguel Bigott"
  WriteRegStr HKLM "${UNINST}" "DisplayIcon" "$INSTDIR\${EXE}"
  WriteRegStr HKLM "${UNINST}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "${UNINST}" "UninstallString" '"$INSTDIR\Desinstalar Codex.exe"'
  WriteRegStr HKLM "${UNINST}" "QuietUninstallString" '"$INSTDIR\Desinstalar Codex.exe" /S'
  WriteRegDWORD HKLM "${UNINST}" "NoModify" 1
  WriteRegDWORD HKLM "${UNINST}" "NoRepair" 1
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKLM "${UNINST}" "EstimatedSize" "$0"

  ; Actualización desde el propio launcher: vuelve a abrirlo al terminar.
  ${GetParameters} $R0
  ClearErrors
  ${GetOptions} $R0 "/reiniciar" $R1
  ${IfNot} ${Errors}
    Exec '"$WINDIR\explorer.exe" "$INSTDIR\${EXE}"'
  ${EndIf}
SectionEnd

Section "Uninstall"
  nsExec::Exec 'taskkill /IM "${EXE}" /F'
  Sleep 500
  Delete "$INSTDIR\${EXE}"
  Delete "$INSTDIR\LICENSE.txt"
  Delete "$INSTDIR\LICENCIAS-TERCEROS.txt"
  Delete "$INSTDIR\Desinstalar Codex.exe"
  ; Solo se borra la carpeta si quedó vacía: las apps instaladas tienen su propio desinstalador.
  RMDir "$INSTDIR"
  Delete "$SMPROGRAMS\${APP}.lnk"
  DeleteRegKey HKLM "${UNINST}"
  DeleteRegKey HKLM "${REGAPP}"
  SetShellVarContext current
  Delete "$DESKTOP\${APP}.lnk"
  ; «Abrir Codex al iniciar sesión».
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${APP}"
  ; Configuración y caché del launcher.
  RMDir /r "$LOCALAPPDATA\Codex"
SectionEnd
