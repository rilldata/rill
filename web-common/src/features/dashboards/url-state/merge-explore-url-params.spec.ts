import { protoBase64 } from "@bufbuild/protobuf";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { toEphemeralMeasuresParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param";
import { MAX_EPHEMERAL_MEASURES } from "@rilldata/web-common/features/dashboards/ephemeral-measures/validation";
import { compressUrlParams } from "@rilldata/web-common/features/dashboards/url-state/compression";
import {
  addEphemeralMeasureLibraryToUrlParams,
  type ExploreUrlParamsSources,
  fillExploreUrlParams,
  mergeExploreUrlParams,
  normalizeExploreUrlParams,
} from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params";
import {
  DashboardState,
  DashboardState_ActivePage,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { describe, expect, it } from "vitest";

const RillDefaultParams =
  "view=explore&tr=rill-QTD&tz=UTC&compare_tr=&f=&measures=*&dims=*&sort_by=impressions&sort_type=value&sort_dir=DESC&leaderboard_measures=impressions";
const YAMLParams =
  "view=explore&tr=P7D&tz=Asia/Kathmandu&f=publisher IN ('Contoso')&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";
const MostRecentParams =
  "tz=America/New_York&measures=bid_price&dims=domain&sort_by=bid_price";
const BookmarkParams =
  "view=explore&tr=PT24H&compare_tr=rill-PP&f=&measures=impressions&sort_by=impressions";

describe("mergeExploreUrlParams", () => {
  const TestCases: {
    title: string;
    sources: Partial<Record<keyof ExploreUrlParamsSources, string | boolean>>;
    expected: string;
  }[] = [
    {
      title: "url is filled from rill defaults only",
      sources: { urlParams: "measures=bid_price&f=domain IN ('x')" },
      expected:
        "measures=bid_price&f=domain IN ('x')&view=explore&tr=rill-QTD&tz=UTC&compare_tr=&dims=*&sort_by=impressions&sort_type=value&sort_dir=DESC&leaderboard_measures=impressions",
    },
    {
      title: "empty url after the back button is the rill defaults",
      sources: {
        urlParams: "",
        sessionParams: "view=explore&tr=P14D",
        backButtonUsed: true,
      },
      expected: RillDefaultParams,
    },
    {
      title: "pivot url is filled with shared params only",
      sources: { urlParams: "view=pivot&rows=publisher&sort_by=publisher" },
      expected:
        "view=pivot&rows=publisher&sort_by=publisher&tr=rill-QTD&tz=UTC&compare_tr=&f=",
    },
    {
      title: "session entry of a visited view wins",
      sources: {
        urlParams: "",
        sessionParams:
          "view=explore&tr=P14D&tz=UTC&compare_tr=rill-PW&f=&measures=impressions&dims=publisher&sort_by=impressions&sort_type=delta_abs&sort_dir=DESC&leaderboard_measures=impressions",
      },
      expected:
        "view=explore&tr=P14D&tz=UTC&compare_tr=rill-PW&f=&measures=impressions&dims=publisher&sort_by=impressions&sort_type=delta_abs&sort_dir=DESC&leaderboard_measures=impressions",
    },
    {
      title:
        "session entry of a view not visited yet is filled from most recent, bookmark, yaml then rill defaults",
      sources: {
        urlParams: "",
        sessionParams: "view=explore&tr=P14D&f=",
      },
      expected:
        "view=explore&tr=P14D&f=&tz=America/New_York&measures=bid_price&dims=domain&sort_by=bid_price&compare_tr=rill-PP&sort_type=percent&sort_dir=ASC&leaderboard_measures=impressions",
    },
    {
      title:
        "fresh visit is the bookmark with most recent over it, filled from yaml then rill defaults",
      sources: { urlParams: "" },
      expected:
        "view=explore&tr=PT24H&compare_tr=rill-PP&f=&measures=bid_price&sort_by=bid_price&tz=America/New_York&dims=domain&sort_type=percent&sort_dir=ASC&leaderboard_measures=impressions",
    },
    {
      title: "fresh visit without a bookmark starts from most recent",
      sources: { urlParams: "", bookmarkParams: false },
      expected:
        "tz=America/New_York&measures=bid_price&dims=domain&sort_by=bid_price&view=explore&tr=P7D&f=publisher IN ('Contoso')&sort_type=percent&sort_dir=ASC&compare_tr=&leaderboard_measures=impressions",
    },
    {
      title: "most recent does not apply to a pivot bookmark's view params",
      sources: {
        urlParams: "",
        bookmarkParams: "view=pivot&rows=publisher&sort_by=publisher&f=",
      },
      expected:
        "view=pivot&rows=publisher&sort_by=publisher&f=&tz=America/New_York&tr=P7D&compare_tr=",
    },
  ];

  for (const { title, sources, expected } of TestCases) {
    it(title, () => {
      const paramsOf = (
        key: keyof ExploreUrlParamsSources,
        fallback: string | undefined,
      ) => {
        const value = key in sources ? sources[key] : fallback;
        return typeof value === "string"
          ? new URLSearchParams(value)
          : undefined;
      };
      const params = mergeExploreUrlParams({
        urlParams: paramsOf("urlParams", "")!,
        sessionParams: paramsOf("sessionParams", undefined),
        mostRecentParams: paramsOf("mostRecentParams", MostRecentParams),
        bookmarkParams: paramsOf("bookmarkParams", BookmarkParams),
        yamlParams: paramsOf("yamlParams", YAMLParams)!,
        rillDefaultParams: paramsOf("rillDefaultParams", RillDefaultParams)!,
        backButtonUsed: sources.backButtonUsed === true,
      });
      expect(sortedParams(params)).toEqual(sortedParams(expected));
    });
  }
});

describe("fillExploreUrlParams", () => {
  const TestCases: {
    title: string;
    target: string;
    source: string;
    overwrite?: boolean;
    expected: string;
  }[] = [
    {
      title: "explore sort is not copied to pivot",
      target: "view=pivot&rows=publisher",
      source: "tz=UTC&measures=*&sort_by=impressions&sort_dir=ASC",
      expected: "view=pivot&rows=publisher&tz=UTC",
    },
    {
      title: "params shared by explore and tdd are copied between them",
      target: "view=tdd&measure=impressions",
      source: "view=explore&grain=day&chart_type=bar&measures=*",
      expected: "view=tdd&measure=impressions&grain=day&chart_type=bar",
    },
    {
      title: "a missing view is explore",
      target: "tr=P7D",
      source: "view=explore&measures=impressions",
      expected: "tr=P7D&view=explore&measures=impressions",
    },
    {
      title: "empty params of the target hold",
      target: "compare_tr=&f=",
      source: "compare_tr=rill-PP&f=publisher IN ('Contoso')&tr=P7D",
      expected: "compare_tr=&f=&tr=P7D",
    },
    {
      title: "overwrite replaces params of the target",
      target: "tr=PT24H&measures=impressions",
      source: "measures=bid_price&tz=UTC",
      overwrite: true,
      expected: "tr=PT24H&measures=bid_price&tz=UTC",
    },
  ];

  for (const { title, target, source, overwrite, expected } of TestCases) {
    it(title, () => {
      const targetParams = new URLSearchParams(target);
      fillExploreUrlParams(targetParams, new URLSearchParams(source), {
        overwrite: !!overwrite,
      });
      expect(sortedParams(targetParams)).toEqual(sortedParams(expected));
    });
  }
});

describe("normalizeExploreUrlParams", () => {
  it("expands compressed params", () => {
    const params = new URLSearchParams({
      gzipped_state: compressUrlParams("tr=P7D&measures=impressions"),
    });
    expect(normalizeExploreUrlParams(params).toString()).toEqual(
      "tr=P7D&measures=impressions",
    );
  });

  it("expands the legacy state with the url's other params over it", () => {
    const state = protoBase64.enc(
      new DashboardState({
        activePage: DashboardState_ActivePage.DEFAULT,
        visibleMeasures: ["impressions"],
        allMeasuresVisible: false,
        leaderboardMeasure: "impressions",
      }).toBinary(),
    );
    const params = normalizeExploreUrlParams(
      new URLSearchParams({ state, view: "pivot" }),
    );
    expect(params.has("state")).toBe(false);
    expect(params.get("view")).toEqual("pivot");
    expect(params.get("measures")).toEqual("impressions");
  });

  it("keeps a legacy state that does not decode", () => {
    expect(
      normalizeExploreUrlParams(new URLSearchParams("state=abc")).toString(),
    ).toEqual("state=abc");
  });
});

describe("addEphemeralMeasureLibraryToUrlParams", () => {
  const def = (name: string, expression = "impressions * 2") =>
    ({ name, displayName: name, expression }) as EphemeralMeasureDef;

  it("keeps definitions in the params over library copies", () => {
    const params = new URLSearchParams({
      adhoc_m: toEphemeralMeasuresParam([def("a", "impressions * 3")]),
    });
    addEphemeralMeasureLibraryToUrlParams(params, [def("a"), def("b")]);
    expect(params.get("adhoc_m")).toEqual(
      toEphemeralMeasuresParam([def("a", "impressions * 3"), def("b")]),
    );
  });

  it("keeps definitions the params reference when capping", () => {
    const library = Array.from({ length: MAX_EPHEMERAL_MEASURES + 1 }, (_, i) =>
      def(`m${i}`),
    );
    const lastName = `m${MAX_EPHEMERAL_MEASURES}`;
    const params = new URLSearchParams({ measures: `impressions,${lastName}` });
    addEphemeralMeasureLibraryToUrlParams(params, library);
    expect(params.get("adhoc_m")).toEqual(
      toEphemeralMeasuresParam([
        ...library.slice(0, MAX_EPHEMERAL_MEASURES - 1),
        def(lastName),
      ]),
    );
  });
});

function sortedParams(params: URLSearchParams | string) {
  return [...new URLSearchParams(params).entries()].sort(([a], [b]) =>
    a.localeCompare(b),
  );
}
