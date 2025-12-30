import { writable } from "svelte/store";
import type { Theme } from "./theme";

/**
 * Shared store for the currently active dashboard's resolved theme.
 * This allows components outside the Dashboard (like the chat in the layout)
 * to inherit the active dashboard's theme.
 *
 * - Dashboard.svelte publishes its resolved theme to this store
 * - Layout components can subscribe to use it for their ThemeProvider
 * - Set to undefined when no dashboard is active
 */
export const activeDashboardTheme = writable<Theme | undefined>(undefined);
