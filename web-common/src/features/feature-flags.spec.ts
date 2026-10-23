import { get } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";
import {
  alertsEnabled,
  featureFlags,
  reportsEnabled,
} from "@rilldata/web-common/features/feature-flags";

describe("alertsEnabled and reportsEnabled", () => {
  afterEach(() => {
    featureFlags.adminServer.resetToDefault();
    featureFlags.embedAdminActions.resetToDefault();
    featureFlags.alerts.resetToDefault();
    featureFlags.reports.resetToDefault();
  });

  const cases = [
    {
      name: "Rill Developer",
      adminServer: false,
      embedAdminActions: false,
      expected: false,
    },
    {
      name: "Rill Cloud",
      adminServer: true,
      embedAdminActions: false,
      expected: true,
    },
    {
      name: "embed with admin token",
      adminServer: false,
      embedAdminActions: true,
      expected: true,
    },
  ];

  for (const { name, adminServer, embedAdminActions, expected } of cases) {
    it(`is ${expected} in ${name}`, () => {
      featureFlags.set(adminServer, "adminServer");
      featureFlags.set(embedAdminActions, "embedAdminActions");
      expect(get(alertsEnabled)).toBe(expected);
      expect(get(reportsEnabled)).toBe(expected);
    });
  }

  it("is false when the project disables the feature", () => {
    featureFlags.set(true, "adminServer", "embedAdminActions");
    featureFlags.set(false, "alerts", "reports");
    expect(get(alertsEnabled)).toBe(false);
    expect(get(reportsEnabled)).toBe(false);
  });
});
