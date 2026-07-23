import { describe, expect, it, vi } from "vitest";
import type { CategorisedOrganizationBillingIssues } from "../selectors";
import {
  createUpgradeToPlanAction,
  type UpgradeToPlanDependencies,
} from "./upgrade-to-plan";

const CURRENT_URL = "https://app.example/acme/settings";

function issues(
  overrides: Partial<CategorisedOrganizationBillingIssues> = {},
): CategorisedOrganizationBillingIssues {
  return {
    payment: [],
    needsPaymentSetup: false,
    ...overrides,
  };
}

function dependencies(
  overrides: Partial<UpgradeToPlanDependencies> = {},
): UpgradeToPlanDependencies {
  return {
    getPage: vi.fn(
      () =>
        ({ url: new URL(CURRENT_URL) }) as ReturnType<
          UpgradeToPlanDependencies["getPage"]
        >,
    ),
    getCurrentUrl: vi.fn(() => CURRENT_URL),
    getAllowedRedirectOrigins: vi.fn(() => ["https://developer.example"]),
    fetchPaymentPortalUrl: vi
      .fn()
      .mockResolvedValue("https://payments.example/session"),
    fetchPlan: vi.fn().mockResolvedValue({ displayName: "Growth" }),
    renew: vi.fn().mockResolvedValue(undefined),
    update: vi.fn().mockResolvedValue(undefined),
    notifyRenewed: vi.fn(),
    showWelcome: vi.fn(),
    invalidate: vi.fn(),
    open: vi.fn(),
    ...overrides,
  };
}

describe("upgradeToPlan", () => {
  it("requests one short-lived portal URL when payment setup is required", async () => {
    // Payment issues must take the portal path and skip plan mutations.
    const deps = dependencies();
    const upgrade = createUpgradeToPlanAction(deps);

    await upgrade(
      "acme",
      "growth",
      issues({ payment: [{}] as never[], needsPaymentSetup: true }),
      null,
    );

    expect(deps.fetchPaymentPortalUrl).toHaveBeenCalledOnce();
    expect(deps.fetchPaymentPortalUrl).toHaveBeenCalledWith(
      "acme",
      "https://app.example/acme/-/upgrade-callback?plan=growth",
      true,
    );
    expect(deps.open).toHaveBeenCalledWith("https://payments.example/session");
    expect(deps.fetchPlan).not.toHaveBeenCalled();
    expect(deps.invalidate).not.toHaveBeenCalled();
  });

  it("does not navigate when the payment portal request fails", async () => {
    // A failed portal lookup must not trigger any post-request effects.
    const deps = dependencies({
      fetchPaymentPortalUrl: vi.fn().mockRejectedValue(new Error("offline")),
    });
    const upgrade = createUpgradeToPlanAction(deps);

    await expect(
      upgrade("acme", "growth", issues({ payment: [{}] as never[] }), null),
    ).rejects.toThrow("offline");

    expect(deps.open).not.toHaveBeenCalled();
    expect(deps.invalidate).not.toHaveBeenCalled();
  });

  it("renews a cancelled plan and reports success", async () => {
    // Cancelled subscriptions renew instead of issuing an update.
    const deps = dependencies();
    const upgrade = createUpgradeToPlanAction(deps);

    await upgrade("acme", "growth", issues({ cancelled: {} as never }), null);

    expect(deps.renew).toHaveBeenCalledOnce();
    expect(deps.renew).toHaveBeenCalledWith("acme", "growth");
    expect(deps.update).not.toHaveBeenCalled();
    expect(deps.notifyRenewed).toHaveBeenCalledWith("Growth");
    expect(deps.invalidate).toHaveBeenCalledWith("acme");
  });

  it("leaves post-mutation effects untouched when renewal fails", async () => {
    // Renewal errors must stop notification, invalidation, and navigation.
    const deps = dependencies({
      renew: vi.fn().mockRejectedValue(new Error("renew failed")),
    });
    const upgrade = createUpgradeToPlanAction(deps);

    await expect(
      upgrade(
        "acme",
        "growth",
        issues({ cancelled: {} as never }),
        "https://app.example/return",
      ),
    ).rejects.toThrow("renew failed");

    expect(deps.notifyRenewed).not.toHaveBeenCalled();
    expect(deps.invalidate).not.toHaveBeenCalled();
    expect(deps.open).not.toHaveBeenCalled();
  });

  it("updates a valid plan and shows the welcome state", async () => {
    // A standard upgrade updates the plan and shows the default success UI.
    const deps = dependencies();
    const upgrade = createUpgradeToPlanAction(deps);

    await upgrade("acme", "growth", issues(), null);

    expect(deps.update).toHaveBeenCalledWith("acme", "growth");
    expect(deps.showWelcome).toHaveBeenCalledWith("growth");
    expect(deps.invalidate).toHaveBeenCalledWith("acme");
  });

  it("surfaces update failure and permits a clean retry", async () => {
    // Failed mutations release the in-flight guard for a later retry.
    const update = vi
      .fn()
      .mockRejectedValueOnce(new Error("update failed"))
      .mockResolvedValueOnce(undefined);
    const deps = dependencies({ update });
    const upgrade = createUpgradeToPlanAction(deps);

    await expect(
      upgrade("acme", "growth", issues(), "https://app.example/return"),
    ).rejects.toThrow("update failed");
    expect(deps.invalidate).not.toHaveBeenCalled();
    expect(deps.open).not.toHaveBeenCalled();

    await upgrade("acme", "growth", issues(), null);
    expect(update).toHaveBeenCalledTimes(2);
    expect(deps.invalidate).toHaveBeenCalledOnce();
  });

  it("rejects an invalid plan before mutation", async () => {
    // Missing plan metadata must fail before any billing write.
    const deps = dependencies({
      fetchPlan: vi.fn().mockResolvedValue(undefined),
    });
    const upgrade = createUpgradeToPlanAction(deps);

    await expect(upgrade("acme", "bogus", issues(), null)).rejects.toThrow(
      "Plan bogus not found",
    );

    expect(deps.renew).not.toHaveBeenCalled();
    expect(deps.update).not.toHaveBeenCalled();
    expect(deps.invalidate).not.toHaveBeenCalled();
  });

  it.each([
    ["same-origin", "https://app.example/return"],
    ["configured origin", "https://developer.example/return"],
  ])("opens an approved %s redirect", async (_name, redirect) => {
    // Exact same-origin and configured-origin redirects are allowed.
    const deps = dependencies();
    const upgrade = createUpgradeToPlanAction(deps);

    await upgrade("acme", "growth", issues(), redirect);

    expect(deps.open).toHaveBeenCalledWith(redirect);
    expect(deps.showWelcome).not.toHaveBeenCalled();
  });

  it("rejects a lookalike external redirect", async () => {
    // Hostname lookalikes must fall back to the normal welcome state.
    const deps = dependencies();
    const upgrade = createUpgradeToPlanAction(deps);

    await upgrade(
      "acme",
      "growth",
      issues(),
      "https://developer.example.evil.test/return",
    );

    expect(deps.open).not.toHaveBeenCalled();
    expect(deps.showWelcome).toHaveBeenCalledWith("growth");
  });

  it("coalesces repeated clicks into one mutation", async () => {
    // Concurrent clicks must share one request and one invalidation.
    let resolveUpdate!: () => void;
    const update = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveUpdate = resolve;
        }),
    );
    const deps = dependencies({ update });
    const upgrade = createUpgradeToPlanAction(deps);

    const first = upgrade("acme", "growth", issues(), null);
    const second = upgrade("acme", "growth", issues(), null);
    await vi.waitFor(() => expect(update).toHaveBeenCalledOnce());

    expect(second).toBe(first);
    resolveUpdate();
    await Promise.all([first, second]);
    expect(deps.invalidate).toHaveBeenCalledOnce();
  });
});
