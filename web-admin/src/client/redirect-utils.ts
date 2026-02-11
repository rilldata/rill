import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
import { redirect } from "@sveltejs/kit";

/**
 * Redirects to the login page by throwing a SvelteKit redirect.
 * Use this in SvelteKit load functions (+page.ts, +layout.ts, etc.)
 */
export function redirectToLogin() {
  throw redirect(307, buildLoginUrl());
}

/**
 * Redirects to the login page using window.location.href.
 * Use this in Svelte component event handlers (onClick, etc.)
 */
export function redirectToLoginFromComponent() {
  window.location.href = buildLoginUrl();
}

export function redirectToLogout() {
  window.location.href = buildLogoutUrl();
}

export function redirectToGithubLogin(
  remote: string,
  redirect: string | null,
  mode: "auth" | "connect",
) {
  window.location.href = buildGithubLoginUrl(remote, redirect, mode);
}

function buildLoginUrl() {
  const u = new URL(ADMIN_URL);
  u.pathname = appendPath(u.pathname, "auth/login");
  u.searchParams.set("redirect", window.location.href);
  return u.toString();
}

function buildLogoutUrl() {
  const u = new URL(ADMIN_URL);
  u.pathname = appendPath(u.pathname, "auth/logout");
  u.searchParams.set("redirect", buildLoginUrl());
  return u.toString();
}

function buildGithubLoginUrl(
  remote: string,
  redirect: string | null,
  mode: "auth" | "connect",
) {
  const u = new URL(ADMIN_URL);
  switch (mode) {
    case "auth":
      u.pathname = appendPath(u.pathname, "github/auth/login");
      break;
    case "connect":
      u.pathname = appendPath(u.pathname, "github/connect");
  }
  u.searchParams.set("remote", remote);
  if (redirect) {
    u.searchParams.set("redirect", redirect);
  }
  return u.toString();
}

function appendPath(path: string, suffix: string) {
  return `${path.replace(/\/$/, "")}/${suffix}`;
}
