#define MyAppName "Messenger"
#define MyAppVersion "1.0"
#define MyAppPublisher "LambdaHub"
#define MyAppURL "https://lambdahub.ru"

[Setup]
AppId={{B3F2A1C4-7D9E-4F2A-8B3C-1A2B3C4D5E6F}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
DefaultDirName={autopf}\Messenger
DefaultGroupName={#MyAppName}
OutputDir=output
OutputBaseFilename=MessengerSetup
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
SetupIconFile=..\web\favicon.ico
UninstallDisplayIcon={app}\messenger.ico
MinVersion=10.0
PrivilegesRequired=admin

[Languages]
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "autostart"; Description: "Запускать при старте Windows"; GroupDescription: "Автозапуск:"

[Files]
; Основные файлы приложения
Source: "..\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs; \
  Excludes: ".git\*,.idea\*,installer\*,*.md"

; Скрипты управления
Source: "..\start.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\stop.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\update.bat"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Запустить Messenger"; Filename: "{app}\start.bat"; WorkingDir: "{app}"; IconFilename: "{app}\messenger.ico"
Name: "{group}\Остановить Messenger"; Filename: "{app}\stop.bat"; WorkingDir: "{app}"
Name: "{group}\Открыть в браузере"; Filename: "http://localhost"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{commondesktop}\Messenger"; Filename: "{app}\start.bat"; WorkingDir: "{app}"; IconFilename: "{app}\messenger.ico"; Tasks: desktopicon

[Registry]
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; \
  ValueName: "Messenger"; ValueData: """{app}\start.bat"""; Flags: uninsdeletevalue; Tasks: autostart

[Run]
Filename: "{app}\install_prereqs.bat"; Description: "Проверка и установка Docker Desktop"; \
  Flags: runhidden waituntilterminated; StatusMsg: "Проверяем Docker Desktop..."
Filename: "{app}\start.bat"; Description: "Запустить Messenger сейчас"; \
  Flags: nowait postinstall skipifsilent; StatusMsg: "Запускаем Messenger..."

[UninstallRun]
Filename: "docker"; Parameters: "compose -f ""{app}\docker-compose.local.yml"" down -v"; \
  WorkingDir: "{app}"; Flags: runhidden waituntilterminated

[Code]
function InitializeSetup(): Boolean;
begin
  Result := True;
end;
