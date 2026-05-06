import { writable } from "svelte/store";

// ── App-wide flags ─────────────────────────────────────────────────────
//
// These flags depend on which frontend app is running (web-admin, web-local,
// or an embed iframe), not on which Rill runtime is connected. Each app's
// root layout seeds them at startup; child layouts (e.g. embed) override
// where needed.

export const adminServer = writable(false);
export const readOnly = writable(false);
export const legacyArchiveDeploy = writable(
  !!import.meta.env.VITE_PLAYWRIGHT_TEST,
);
