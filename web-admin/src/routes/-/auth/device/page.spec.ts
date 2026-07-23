import { describe, expect, it, vi } from "vitest";
import {
  buildDeviceAuthorizationRequestUrl,
  createDeviceAuthorizationAction,
  type DeviceAuthorizationDependencies,
  type DeviceAuthorizationInput,
  type DeviceAuthorizationState,
} from "./device-authorization-action";

const messages = {
  confirmed: "confirmed",
  rejected: "rejected",
  confirmationFailed: "confirmation failed",
  rejectionFailed: "rejection failed",
  invalidCode: "invalid code",
  unsafeRedirect: "unsafe redirect",
};

function input(
  overrides: Partial<DeviceAuthorizationInput> = {},
): DeviceAuthorizationInput {
  return {
    userCode: "ABCD-1234",
    confirmed: true,
    redirect: null,
    currentUrl: "https://app.example/-/auth/device",
    ...overrides,
  };
}

function response(ok: boolean, body = "") {
  return {
    ok,
    text: vi.fn().mockResolvedValue(body),
  };
}

function setup(overrides: Partial<DeviceAuthorizationDependencies> = {}) {
  const states: DeviceAuthorizationState[] = [];
  const dependencies: DeviceAuthorizationDependencies = {
    adminUrl: "https://admin.example/api",
    fetch: vi.fn().mockResolvedValue(response(true)),
    navigate: vi.fn(),
    messages,
    onState: (state) => states.push(state),
    ...overrides,
  };
  return {
    dependencies,
    states,
    authorize: createDeviceAuthorizationAction(dependencies),
  };
}

describe("device authorization action", () => {
  it("confirms a code with one credentialed POST", async () => {
    // Confirmation sends one encoded request and records completed state.
    const { authorize, dependencies, states } = setup();

    await authorize(input());

    expect(dependencies.fetch).toHaveBeenCalledOnce();
    const [requestUrl, init] = vi.mocked(dependencies.fetch).mock.calls[0];
    const parsed = new URL(requestUrl);
    expect(parsed.pathname).toBe("/api/auth/oauth/device");
    expect(parsed.searchParams.get("user_code")).toBe("ABCD-1234");
    expect(parsed.searchParams.get("code_confirmed")).toBe("true");
    expect(init).toEqual({ method: "POST", credentials: "include" });
    expect(states.at(-1)).toEqual({
      inFlight: false,
      completed: true,
      successMessage: "confirmed",
      errorMessage: "",
    });
  });

  it("rejects a code successfully without following a redirect", async () => {
    // Rejection records success but never follows the supplied callback.
    const { authorize, dependencies, states } = setup();

    await authorize(
      input({
        confirmed: false,
        redirect: "https://app.example/should-not-open",
      }),
    );

    const [requestUrl] = vi.mocked(dependencies.fetch).mock.calls[0];
    expect(new URL(requestUrl).searchParams.get("code_confirmed")).toBe(
      "false",
    );
    expect(dependencies.navigate).not.toHaveBeenCalled();
    expect(states.at(-1)?.successMessage).toBe("rejected");
  });

  it.each([
    ["confirmation", true, "confirmation failed"],
    ["rejection", false, "rejection failed"],
  ])(
    "displays a failed %s response and enables retry",
    async (_name, confirmed, text) => {
      // HTTP failures show decision-specific detail and unlock a retry.
      const fetch = vi
        .fn()
        .mockResolvedValueOnce(response(false, "expired code"))
        .mockResolvedValueOnce(response(true));
      const { authorize, dependencies, states } = setup({ fetch });
      const request = input({ confirmed });

      await authorize(request);

      expect(states.at(-1)).toEqual({
        inFlight: false,
        completed: false,
        successMessage: "",
        errorMessage: `${text}: expired code`,
      });
      expect(dependencies.navigate).not.toHaveBeenCalled();

      await authorize(request);
      expect(fetch).toHaveBeenCalledTimes(2);
      expect(states.at(-1)?.completed).toBe(true);
    },
  );

  it("displays a network error without completing the action", async () => {
    // Transport failures leave the code incomplete and surface their detail.
    const { authorize, states } = setup({
      fetch: vi.fn().mockRejectedValue(new Error("offline")),
    });

    await authorize(input());

    expect(states.at(-1)).toMatchObject({
      completed: false,
      errorMessage: "confirmation failed: offline",
    });
  });

  it("rejects a malformed code before issuing a request", async () => {
    // Invalid user codes are rejected locally before URL construction.
    const { authorize, dependencies, states } = setup();

    await authorize(input({ userCode: "ABC&code_confirmed=false" }));

    expect(dependencies.fetch).not.toHaveBeenCalled();
    expect(states.at(-1)).toMatchObject({
      inFlight: false,
      completed: false,
      errorMessage: "invalid code",
    });
  });

  it("encodes parameters without allowing user input to replace them", () => {
    // URLSearchParams must preserve attacker-like text as one code value.
    const requestUrl = buildDeviceAuthorizationRequestUrl(
      "https://admin.example/api?old=value",
      "ABC&code_confirmed=false",
      true,
    );
    const parsed = new URL(requestUrl);

    expect(parsed.searchParams.get("user_code")).toBe(
      "ABC&code_confirmed=false",
    );
    expect(parsed.searchParams.getAll("code_confirmed")).toEqual(["true"]);
    expect(parsed.searchParams.has("old")).toBe(false);
  });

  it("keeps both buttons disabled while the request is in flight", async () => {
    // Concurrent decisions share the active request until it settles.
    let resolveRequest!: (value: ReturnType<typeof response>) => void;
    const fetch = vi.fn(
      () =>
        new Promise<ReturnType<typeof response>>((resolve) => {
          resolveRequest = resolve;
        }),
    );
    const { authorize, states } = setup({ fetch });
    const request = input();

    const first = authorize(request);
    const second = authorize(request);

    expect(states.at(-1)).toMatchObject({
      inFlight: true,
      completed: false,
    });
    expect(second).toBe(first);
    expect(fetch).toHaveBeenCalledOnce();

    resolveRequest(response(true));
    await Promise.all([first, second]);
    expect(states.at(-1)).toMatchObject({
      inFlight: false,
      completed: true,
    });
  });

  it.each([
    "https://app.example/complete",
    "http://localhost:43123/callback",
    "http://127.0.0.1:43123/callback?code=a%20b",
    "http://[::1]:43123/callback",
  ])(
    "allows an approved same-origin or loopback redirect: %s",
    async (redirect) => {
      // Browser-local callbacks are allowed only by the explicit policy.
      const { authorize, dependencies } = setup();

      await authorize(input({ redirect }));

      expect(dependencies.navigate).toHaveBeenCalledWith(
        new URL(redirect).toString(),
      );
    },
  );

  it.each([
    "https://attacker.example/callback",
    "http://localhost.attacker.example/callback",
    "javascript:alert(1)",
  ])("rejects an external redirect after confirming: %s", async (redirect) => {
    // Unsafe callbacks do not undo confirmation, but they never navigate.
    const { authorize, dependencies, states } = setup();

    await authorize(input({ redirect }));

    expect(dependencies.fetch).toHaveBeenCalledOnce();
    expect(dependencies.navigate).not.toHaveBeenCalled();
    expect(states.at(-1)).toEqual({
      inFlight: false,
      completed: true,
      successMessage: "confirmed",
      errorMessage: "unsafe redirect",
    });
  });
});
