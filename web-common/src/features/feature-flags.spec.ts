import { get } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";
import {
  featureFlags,
  reportsEnabled,
} from "@rilldata/web-common/features/feature-flags";

describe("reportsEnabled", () => {
  afterEach(() => {
    featureFlags.adminServer.resetToDefault();
    featureFlags.embedAdminActions.resetToDefault();
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
      expect(get(reportsEnabled)).toBe(expected);
    });
  }

  it("is false when the project disables the feature", () => {
    featureFlags.set(true, "adminServer", "embedAdminActions");
    featureFlags.set(false, "reports");
    expect(get(reportsEnabled)).toBe(false);
  });
});
