import { Settings } from "luxon";
import { baseLocale } from "@rilldata/web-common/paraglide/runtime.js";

export function syncLuxonLocale(locale: string | undefined | null): void {
  Settings.defaultLocale = locale || baseLocale;
}
