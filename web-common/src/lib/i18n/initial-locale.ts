import { normalizeLocale } from "@rilldata/web-common/lib/i18n/normalize-locale";
import type { Locale } from "@rilldata/web-common/lib/i18n/gen/runtime";

export type InitialLocaleAction =
  | { type: "set-and-reload"; locale: Locale }
  | { type: "keep" };

export function resolveInitialLocale(
  storedPreference: string | undefined | null,
  currentLocale: string,
): InitialLocaleAction {
  const preferred = normalizeLocale(storedPreference);
  if (preferred) {
    return preferred === currentLocale
      ? { type: "keep" }
      : { type: "set-and-reload", locale: preferred };
  }
  return { type: "keep" };
}
