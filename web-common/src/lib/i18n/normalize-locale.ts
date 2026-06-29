import {
  locales,
  type Locale,
} from "@rilldata/web-common/paraglide/runtime.js";

export function normalizeLocale(
  locale: string | undefined | null,
): Locale | undefined {
  if (!locale) return undefined;
  const base = locale.toLowerCase().split(/[-_]/)[0];
  return (locales as readonly string[]).includes(base)
    ? (base as Locale)
    : undefined;
}
