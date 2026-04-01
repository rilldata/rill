import { writable } from "svelte/store";

/**
 * When true, custom themes are disabled and dashboards fall back to the default (no theme).
 * Set by web-admin based on the organization's billing plan.
 * web-local never sets this, so local dev themes are unaffected.
 */
export const customThemesDisabled = writable(false);
