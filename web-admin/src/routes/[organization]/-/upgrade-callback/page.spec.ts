import { describe, expect, it, vi } from "vitest";
import {
  createUpgradeCallbackAction,
  type UpgradeCallbackDependencies,
  type UpgradeCallbackInput,
  type UpgradeCallbackState,
} from "./upgrade-callback-action";

type PaymentIssue = { reason: string };

function input(
  overrides: Partial<UpgradeCallbackInput<PaymentIssue>> = {},
): UpgradeCallbackInput<PaymentIssue> {
  return {
    organization: "acme",
    planName: "growth",
    cancelled: false,
    paymentIssues: [],
    redirect: null,
    currentUrl: "https://app.example/acme/-/upgrade-callback",
    allowedRedirectOrigins: ["https://developer.example"],
    ...overrides,
  };
}

function setup(
  overrides: Partial<UpgradeCallbackDependencies<PaymentIssue>> = {},
) {
  const states: UpgradeCallbackState[] = [];
  const dependencies: UpgradeCallbackDependencies<PaymentIssue> = {
    getPaymentPortalUrl: vi
      .fn()
      .mockResolvedValue("https://payments.example/session"),
    notifyPaymentIssue: vi.fn(),
    fetchPlan: vi.fn().mockResolvedValue({ displayName: "Growth" }),
    renew: vi.fn().mockResolvedValue(undefined),
    update: vi.fn().mockResolvedValue(undefined),
    notifyRenewed: vi.fn(),
    showWelcome: vi.fn(),
    invalidate: vi.fn(),
    navigate: vi.fn(),
    open: vi.fn(),
    formatError: (error) =>
      error instanceof Error ? error.message : "Unknown error",
    onState: (state) => states.push(state),
    ...overrides,
  };
  return {
    dependencies,
    states,
    run: createUpgradeCallbackAction(dependencies),
  };
}

describe("upgrade callback action", () => {
  it("returns payment issues to billing without mutating a plan", async () => {
    // Payment issues return to billing through a fresh portal URL.
    const { dependencies, run, states } = setup();
    const request = input({
      paymentIssues: [{ reason: "payment-required" }],
    });

    await run(request);

    expect(dependencies.getPaymentPortalUrl).toHaveBeenCalledWith(request);
    expect(dependencies.notifyPaymentIssue).toHaveBeenCalledWith(
      request,
      "https://payments.example/session",
    );
    expect(dependencies.navigate).toHaveBeenCalledWith(
      "/acme/-/settings/billing",
    );
    expect(dependencies.update).not.toHaveBeenCalled();
    expect(dependencies.invalidate).not.toHaveBeenCalled();
    expect(states.at(-1)).toMatchObject({ completed: true, errorMessage: "" });
  });

  it("displays a portal request failure and stays on the callback", async () => {
    // Portal failures remain retryable and do not navigate away.
    const { dependencies, run, states } = setup({
      getPaymentPortalUrl: vi
        .fn()
        .mockRejectedValue(new Error("portal unavailable")),
    });

    await run(input({ paymentIssues: [{ reason: "payment-required" }] }));

    expect(states.at(-1)).toEqual({
      inFlight: false,
      completed: false,
      errorMessage: "portal unavailable",
    });
    expect(dependencies.notifyPaymentIssue).not.toHaveBeenCalled();
    expect(dependencies.navigate).not.toHaveBeenCalled();
  });

  it("updates, invalidates, and navigates internally after success", async () => {
    // A successful update commits cache and default navigation effects.
    const { dependencies, run, states } = setup();

    await run(input());

    expect(dependencies.update).toHaveBeenCalledWith("acme", "growth");
    expect(dependencies.showWelcome).toHaveBeenCalledWith("growth");
    expect(dependencies.invalidate).toHaveBeenCalledWith("acme");
    expect(dependencies.navigate).toHaveBeenCalledWith("/acme");
    expect(states[0]).toMatchObject({ inFlight: true });
    expect(states.at(-1)).toMatchObject({ inFlight: false, completed: true });
  });

  it("renews and opens an approved configured redirect", async () => {
    // Cancelled subscriptions renew before following an approved redirect.
    const { dependencies, run } = setup();

    await run(
      input({
        cancelled: true,
        redirect: "https://developer.example/project",
      }),
    );

    expect(dependencies.renew).toHaveBeenCalledWith("acme", "growth");
    expect(dependencies.update).not.toHaveBeenCalled();
    expect(dependencies.notifyRenewed).toHaveBeenCalledWith("Growth");
    expect(dependencies.open).toHaveBeenCalledWith(
      "https://developer.example/project",
    );
    expect(dependencies.navigate).not.toHaveBeenCalled();
  });

  it.each([
    ["update", false],
    ["renew", true],
  ] as const)(
    "handles a failed %s without invalidating or navigating",
    async (method, cancelled) => {
      // Either billing mutation may fail without committing later effects.
      const failure = vi.fn().mockRejectedValue(new Error(`${method} failed`));
      const { dependencies, run, states } = setup({ [method]: failure });

      await run(
        input({
          cancelled,
          redirect: "https://app.example/return",
        }),
      );

      expect(states.at(-1)).toMatchObject({
        completed: false,
        errorMessage: `${method} failed`,
      });
      expect(dependencies.invalidate).not.toHaveBeenCalled();
      expect(dependencies.open).not.toHaveBeenCalled();
      expect(dependencies.navigate).not.toHaveBeenCalled();
    },
  );

  it("displays an invalid-plan error before any mutation", async () => {
    // Invalid plan input must fail before renew or update is attempted.
    const { dependencies, run, states } = setup({
      fetchPlan: vi.fn().mockResolvedValue(null),
    });

    await run(input({ planName: "bogus" }));

    expect(states.at(-1)?.errorMessage).toBe("Plan bogus not found");
    expect(dependencies.update).not.toHaveBeenCalled();
    expect(dependencies.renew).not.toHaveBeenCalled();
    expect(dependencies.navigate).not.toHaveBeenCalled();
  });

  it.each(["https://app.example/return", "https://developer.example/return"])(
    "opens an approved redirect: %s",
    async (redirect) => {
      // Exact same-origin and configured-origin destinations may open.
      const { dependencies, run } = setup();

      await run(input({ redirect }));

      expect(dependencies.open).toHaveBeenCalledWith(redirect);
      expect(dependencies.navigate).not.toHaveBeenCalled();
    },
  );

  it("rejects an external lookalike and falls back to the organization", async () => {
    // A lookalike hostname falls back to the organization landing page.
    const { dependencies, run } = setup();

    await run(
      input({ redirect: "https://developer.example.attacker.test/return" }),
    );

    expect(dependencies.open).not.toHaveBeenCalled();
    expect(dependencies.navigate).toHaveBeenCalledWith("/acme");
    expect(dependencies.showWelcome).toHaveBeenCalledWith("growth");
  });

  it("coalesces repeated callback execution into one mutation", async () => {
    // Repeated lifecycle execution must share one billing mutation.
    let resolveUpdate!: () => void;
    const update = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveUpdate = resolve;
        }),
    );
    const { dependencies, run } = setup({ update });
    const request = input();

    const first = run(request);
    const second = run(request);
    await vi.waitFor(() => expect(update).toHaveBeenCalledOnce());

    expect(second).toBe(first);
    resolveUpdate();
    await Promise.all([first, second]);
    expect(dependencies.invalidate).toHaveBeenCalledOnce();
  });
});
