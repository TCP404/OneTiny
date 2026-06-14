export interface ConfigDTO {
  rootPath: string;
  port: number;
  maxLevel: number;
  isAllowUpload: boolean;
  isSecure: boolean;
}

export type TabKey = "panel" | "security" | "logs" | "about";

export type AppMode = "browser-preview" | "wails-desktop";

export interface StatusDTO {
  running: boolean;
  stateLabel: string;
  address: string;
  config: ConfigDTO;
  hasCredentials: boolean;
  configPath: string;
  accessLogPath: string;
  portRestartRequired: boolean;
  lastError: string;
}

export interface ConfigPatchDTO {
  rootPath?: string;
  port?: number;
  maxLevel?: number;
  isAllowUpload?: boolean;
  isSecure?: boolean;
  restartPort?: boolean;
}

export interface CredentialPatchDTO {
  username: string;
  password: string;
  confirmPassword: string;
  enableSecure: boolean;
}

export interface LogFilterDTO {
  event?: string;
  since?: string | null;
  until?: string | null;
}

export interface LogEntryDTO {
  time: string;
  clientIP: string;
  method: string;
  event: string;
  path: string;
  status: number;
  result: string;
}

export interface OneTinyService {
  GetStatus(): Promise<StatusDTO>;
  StartSharing(): Promise<StatusDTO>;
  StopSharing(): Promise<StatusDTO>;
  UpdateConfig(patch: ConfigPatchDTO): Promise<StatusDTO>;
  SetCredentials(patch: CredentialPatchDTO): Promise<StatusDTO>;
  GetLogs(filter: LogFilterDTO): Promise<LogEntryDTO[]>;
  ClearLogs(): Promise<void>;
  ChooseDirectory(current: string): Promise<string>;
  ExportLogs(filter: LogFilterDTO): Promise<string>;
  OpenConfigDir(): Promise<void>;
}
