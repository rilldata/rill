import {
  baseLocale,
  isLocale,
  locales,
  overwriteGetLocale,
  type Locale,
} from "@rilldata/web-common/lib/i18n/gen/runtime";

let currentLocale: Locale = baseLocale;

// The host tells the widget which locale to use; paraglide's own strategies (cookie,
// global variable) have nothing to read inside a sandboxed iframe.
overwriteGetLocale(() => currentLocale);

/** Maps a BCP 47 host locale such as "es-MX" onto a supported locale, else the base locale. */
export function resolveLocale(hostLocale: string | undefined): Locale {
  if (!hostLocale) return baseLocale;
  const normalized = hostLocale.toLowerCase();
  if (isLocale(normalized)) return normalized;
  const language = normalized.split(/[-_]/)[0];
  if (isLocale(language)) return language;
  return baseLocale;
}

export function setLocaleFromHost(hostLocale: string | undefined): Locale {
  currentLocale = resolveLocale(hostLocale);
  if (typeof document !== "undefined") {
    document.documentElement.lang = currentLocale;
  }
  return currentLocale;
}

export function getCurrentLocale(): Locale {
  return currentLocale;
}

export const supportedLocales: readonly string[] = locales;
