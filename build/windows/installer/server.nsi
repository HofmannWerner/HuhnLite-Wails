Unicode true

# Include the wails tools for shared defines
!include "wails_tools.nsh"

# The version information
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support.
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
ShowInstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext

    SetOutPath $INSTDIR

    !insertmacro wails.files

    # Create user folders
    CreateDirectory "$INSTDIR\images"
    CreateDirectory "$INSTDIR\backups"

    # Copy help HTML files (overwrite to ensure latest version)
    File "..\..\..\HuhnLite-de.html"
    File "..\..\..\HuhnLite-en.html"
    File "..\..\..\HuhnLite-it.html"

    # Copy help images recursively
    SetOutPath "$INSTDIR\images"
    File /r "..\..\..\images\*"
    SetOutPath $INSTDIR

    # Safely copy database if SQLite version and it doesn't exist
    !ifdef INSTALL_SQLITE_DB
        IfFileExists "$INSTDIR\HuhnLite.db" db_exists
            File "/nonfatal" "..\..\..\HuhnLite.db"
        db_exists:
        IfFileExists "$INSTDIR\HuhnLite_test.db" db_test_exists
            File "/nonfatal" "..\..\..\HuhnLite_test.db"
        db_test_exists:
    !endif

    # Safely copy server settings file
    !ifdef SETTINGS_FILE
        IfFileExists "$INSTDIR\${SETTINGS_FILE}" settings_exists
            File "/oname=${SETTINGS_FILE}" "..\..\..\${SETTINGS_FILE}"
        settings_exists:
    !endif

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    # Delete shortcuts
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    # Delete installed files
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    !ifdef SETTINGS_FILE
        Delete "$INSTDIR\${SETTINGS_FILE}"
    !endif
    Delete "$INSTDIR\HuhnLite-de.html"
    Delete "$INSTDIR\HuhnLite-en.html"
    Delete "$INSTDIR\HuhnLite-it.html"
    Delete "$INSTDIR\images\*"

    # We do NOT delete HuhnLite.db or backups by default here
    # to protect user data from accidental uninstalls.
    # We only remove directories if they are completely empty.
    !ifdef INSTALL_SQLITE_DB
        # Delete "$INSTDIR\HuhnLite.db" # Commented out to prevent database loss
    !endif

    RMDir "$INSTDIR\images"
    RMDir "$INSTDIR\backups"
    RMDir "$INSTDIR"

    !insertmacro wails.deleteUninstaller
SectionEnd
