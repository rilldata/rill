export interface ThemeModeColors {
  primary?: string;
  secondary?: string;
  variables?: Record<string, string>;
  [key: string]: string | undefined | Record<string, string>;
}

export interface V1ThemeSpec {
  light?: ThemeModeColors;
  dark?: ThemeModeColors;
  primaryColorRaw?: string;
  secondaryColorRaw?: string;
}

