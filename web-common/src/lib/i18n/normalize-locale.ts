import {
  locales,
  type Locale,
} from "@rilldata/web-common/lib/i18n/gen/runtime";

export function normalizeLocale(
  locale: string | undefined | null,
): Locale | undefined {
  if (!locale) return undefined;
  const base = locale.toLowerCase().split(/[-_]/)[0];
  return (locales as readonly string[]).includes(base)
    ? (base as Locale)
    : undefined;
}
