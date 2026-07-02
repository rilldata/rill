import { Settings } from "luxon";
import { baseLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";

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

export function syncLuxonLocale(locale: string | undefined | null): void {
  Settings.defaultLocale = locale || baseLocale;
}

export function escapeHtml(value: string | undefined | null): string {
  if (!value) return "";
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}
