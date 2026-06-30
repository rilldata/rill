// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

// Mock SvelteKit's $app/environment so the module can read `browser`.
vi.mock("$app/environment", () => ({ browser: true }));

// Mock ADMIN_URL so the module can build URLs without an env var.
vi.mock("@rilldata/web-admin/client/http-client", () => ({
  ADMIN_URL: "http://admin.example.com/",
}));

// window.location.href is assigned during assume/unassume. jsdom throws on
// navigation, so replace location with a writable stub.
beforeEach(() => {
  localStorage.clear();
  const locationStub = {
    href: "http://app.example.com/",
    origin: "http://app.example.com",
  };
  Object.defineProperty(window, "location", {
    writable: true,
    value: locationStub,
  });
  vi.resetModules();
});

describe("assumedUser", () => {
  it("initializes from localStorage", async () => {
    localStorage.setItem("rill-representing-user", "stored@example.com");
    const { assumedUser } = await import("./assume-state");
    expect(get(assumedUser)).toBe("stored@example.com");
  });

  it("assume() writes email to localStorage and updates the store", async () => {
    const { assumedUser, STORAGE_KEY } = await import("./assume-state");
    assumedUser.assume("target@example.com");
    expect(localStorage.getItem(STORAGE_KEY)).toBe("target@example.com");
    expect(get(assumedUser)).toBe("target@example.com");
  });

  it("unassume() removes the localStorage entry and clears the store", async () => {
    localStorage.setItem("rill-representing-user", "target@example.com");
    const { assumedUser } = await import("./assume-state");
    assumedUser.unassume();
    expect(localStorage.getItem("rill-representing-user")).toBeNull();
    expect(get(assumedUser)).toBe("");
  });

  it("assume() builds an assume-open URL with representing_user and ttl", async () => {
    const { assumedUser } = await import("./assume-state");
    assumedUser.assume("target@example.com", { ttlMinutes: 30 });
    const href = window.location.href;
    const url = new URL(href);
    expect(url.pathname).toBe("/auth/assume-open");
    expect(url.searchParams.get("representing_user")).toBe(
      "target@example.com",
    );
    expect(url.searchParams.get("ttl_minutes")).toBe("30");
  });

  it("assume() forwards the redirect param when provided", async () => {
    const { assumedUser } = await import("./assume-state");
    assumedUser.assume("target@example.com", { redirect: "/acme/project" });
    const url = new URL(window.location.href);
    expect(url.searchParams.get("redirect")).toBe("/acme/project");
  });

  it("unassume() redirects to /auth/login", async () => {
    const { assumedUser } = await import("./assume-state");
    assumedUser.unassume();
    const url = new URL(window.location.href);
    expect(url.pathname).toBe("/auth/login");
    expect(url.searchParams.get("redirect")).toBe("http://app.example.com");
  });
});
