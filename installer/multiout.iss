; Inno Setup Script for Multiout (Windows 64-bit)
; To compile: iscc installer\multiout.iss

#define MyAppName "Multiout"
#define MyAppVersion "1.1.0"
#define MyAppPublisher "codesbysagar"
#define MyAppURL "https://github.com/codesbysagar/multiout"
#define MyAppExeName "multiout.exe"

[Setup]
AppId={{9F571243-7D4C-4D2A-98C3-25A128B237CE}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
OutputDir=..\bin
OutputBaseFilename=multiout_setup_x64
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
ArchitecturesInstallIn64BitMode=x64compatible
ArchitecturesAllowed=x64compatible
UninstallDisplayIcon={app}\{#MyAppExeName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "..\bin\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\internal\ui\multiout-logo-rounded.png"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Messages]
WelcomeLabel2=This will install Multiout on your computer.%n%nSimultaneous Multi-Device Audio Output with Real-Time Drift Compensation.%n%nTip: For seamless Windows audio capture without duplicate playback on your default speakers, a Virtual Audio Cable (e.g. VB-Cable from https://vb-audio.com/Cable/) is recommended.
