import "@testing-library/jest-dom/vitest";
import AlertFilterCriteria from "@rilldata/web-admin/features/alerts/metadata/AlertFilterCriteria.svelte";
import {
  type V1Expression,
  type V1TimeRange,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";

// A percent change criterion, as the alert dialog saves it.
const filters: V1Expression = {
  cond: {
    op: V1Operation.OPERATION_AND,
    exprs: [
      {
        cond: {
          op: V1Operation.OPERATION_LT,
          exprs: [{ ident: "revenue_delta_perc" }, { val: -0.1 }],
        },
      },
    ],
  },
};

describe("AlertFilterCriteria", () => {
  const cases: { comparisonTimeRange: V1TimeRange; label: RegExp }[] = [
    {
      comparisonTimeRange: { expression: "MTD as of latest/D offset -1M" },
      label: /from previous month/i,
    },
    {
      comparisonTimeRange: { isoDuration: "P7D", isoOffset: "P7D" },
      label: /from previous period/i,
    },
  ];
  cases.forEach(({ comparisonTimeRange, label }) => {
    it(`names the comparison of ${JSON.stringify(comparisonTimeRange)}`, () => {
      render(AlertFilterCriteria, { props: { filters, comparisonTimeRange } });
      expect(screen.getByText(label)).toBeInTheDocument();
    });
  });
});
