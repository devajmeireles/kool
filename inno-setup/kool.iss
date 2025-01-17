#define ApplicationId "{{B26A0699-CADB-4927-82DB-82842ADB2271}"
#define ApplicationGroup "Kool.dev"
#define ApplicationName "Kool CLI"

#include "environment.iss"

[Setup]
AppId={#ApplicationId}
AppName={#ApplicationName}
AppVersion={#ApplicationVersion}
AppVerName={#ApplicationName} {#ApplicationVersion}
AppPublisher=Kool Dev Sistemas de Informacao LTDA
AppComments=From development to production, our tools bring speed and security to software development teams from different stacks, making their development environments reproducible and easy to set up.
AppCopyright=© 2022 kool.dev - All rights reserved.
AppPublisherURL=https://kool.dev
AppSupportURL=https://kool.dev
AppUpdatesURL=https://kool.dev
VersionInfoVersion={#ApplicationVersion}
DefaultDirName={autopf}\{#ApplicationGroup}
DisableWelcomePage=No
DisableDirPage=Yes
DisableProgramGroupPage=Yes
DisableReadyPage=Yes
DefaultGroupName={#ApplicationGroup}
LicenseFile=../LICENSE.md
MinVersion=10.0.10240
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin
SetupIconFile=kool.ico
UninstallDisplayIcon={autopf}\{#ApplicationGroup}\kool.ico
UninstallDisplayName={#ApplicationName}
WizardImageStretch=No
WizardImageFile=kool-setup-small-icon.bmp
WizardSmallImageFile=kool-setup-icon.bmp
ArchitecturesInstallIn64BitMode=x64
ChangesEnvironment=true

[Languages]
Name: en; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\dist\kool.exe"; DestDir: "{autopf}\{#ApplicationGroup}\bin"; Flags: ignoreversion
Source: "kool.ico"; DestDir: "{autopf}\{#ApplicationGroup}"; Flags: ignoreversion

[Code]
procedure CurStepChanged(CurStep: TSetupStep);
begin
    if CurStep = ssPostInstall
     then EnvAddPath(ExpandConstant('{autopf}') + '\' + ExpandConstant('{#ApplicationGroup}') + '\bin');
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
    if CurUninstallStep = usPostUninstall
    then EnvRemovePath(ExpandConstant('{autopf}') + '\' + ExpandConstant('{#ApplicationGroup}') + '\bin');
end;
