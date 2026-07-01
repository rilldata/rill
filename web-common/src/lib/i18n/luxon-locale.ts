import { Settings } from "luxon";
import { baseLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";

export function syncLuxonLocale(locale: string | undefined | null): void {
  Settings.defaultLocale = locale || baseLocale;
}
