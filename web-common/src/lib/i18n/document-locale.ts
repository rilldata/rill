import { baseLocale } from "@rilldata/web-common/paraglide/runtime.js";

const RTL_LOCALES = new Set<string>([]);

export function localeDirection(
  locale: string | undefined | null,
): "ltr" | "rtl" {
  const base = (locale || baseLocale).toLowerCase().split(/[-_]/)[0];
  return RTL_LOCALES.has(base) ? "rtl" : "ltr";
}

export function syncDocumentLocale(locale: string | undefined | null): void {
  if (typeof document === "undefined") return;
  const effective = locale || baseLocale;
  const root = document.documentElement;
  root.lang = effective;
  root.dir = localeDirection(effective);
}
