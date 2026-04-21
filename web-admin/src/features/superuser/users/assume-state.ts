// Manages assume/unassume user state for the superuser console.
// Uses localStorage so the banner syncs across all browser tabs;
// auth cookies are browser-wide, so every tab must show the warning.
import { writable } from "svelte/store";
import { browser } from "$app/environment";
import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";

export const STORAGE_KEY = "rill-representing-user";

// Store tracks the currently assumed user email
const initial = browser ? (localStorage.getItem(STORAGE_KEY) ?? "") : "";
const { subscribe, set } = writable(initial);

export const assumedUser = {
  subscribe,

  /**
   * Navigates to Rill Cloud as the given user in the current tab.
   * Writes the email to localStorage so every open tab shows the
   * "browsing as" banner (auth cookies are browser-wide).
   * Optionally redirects to a specific path (e.g. "/<orgName>") after assuming.
   */
  assume(email: string, opts?: { ttlMinutes?: number; redirect?: string }) {
    const { ttlMinutes = 60, redirect } = opts ?? {};
    set(email);
    if (browser) localStorage.setItem(STORAGE_KEY, email);

    const u = new URL("auth/assume-open", ADMIN_URL);
    u.searchParams.set("representing_user", email);
    u.searchParams.set("ttl_minutes", String(ttlMinutes));
    if (redirect) u.searchParams.set("redirect", redirect);
    window.location.href = u.toString();
  },

  /**
   * Reverts to the original superuser session.
   * Redirects to /auth/login which re-authenticates through the auth provider.
   * Since the superuser's auth provider session is still active, this auto-completes
   * and issues a fresh superuser cookie, effectively "unassuming".
   */
  unassume() {
    set("");
    if (browser) localStorage.removeItem(STORAGE_KEY);

    const u = new URL("auth/login", ADMIN_URL);
    u.searchParams.set("redirect", window.location.origin);
    window.location.href = u.toString();
  },
};
