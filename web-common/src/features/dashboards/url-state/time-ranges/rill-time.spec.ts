import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [
      rillTime: string,
      label: string,
      rangeGrain: V1TimeGrain,
      bucketGrain: V1TimeGrain,
    ][] = [
      [
        "m by |s|",
        "Minute to date, incomplete",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_SECOND,
      ],
      [
        "-5m by |m|",
        "Last 5 minutes, incomplete",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE,
      ],
      [
        "-5m to 0m by |m|",
        "Last 5 minutes",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE,
      ],
      [
        "-7d to 0d by |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d to now/d by |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d to now by |h|",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d to now by h",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "d by h",
        "Today, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      // TODO: correct label for the below
      [
        "-7d to -5d by h",
        "-7d to -5d by h",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-2d to now/d by h @ -5d",
        "-2d to now/d by h @ -5d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-2d to now/d @ -5d",
        "-2d to now/d @ -5d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],

      [
        "-7d to now/d by h @ {Asia/Kathmandu}",
        "-7d to now/d by h @ {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d to now/d by |h| @ {Asia/Kathmandu}",
        "-7d to now/d by |h| @ {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d to now/d by |h| @ -5d {Asia/Kathmandu}",
        "-7d to now/d by |h| @ -5d {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      // TODO: should these be something different when end is latest vs now?
      [
        "-7d to latest/d by |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d to latest by |h|",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d to latest by h",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      [
        "2024-01-01 to 2024-03-31",
        "2024-01-01 to 2024-03-31",
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "2024-01-01 10:00 to 2024-03-31 18:00",
        "2024-01-01 10:00 to 2024-03-31 18:00",
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],

      [
        "-7W+2d",
        "-7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W to latest+2d",
        "-7W to latest+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W to -7W+2d",
        "-7W to -7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W+2d",
        "-7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W + 2d",
        "-7W + 2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "2024-01-01-1W to 2024-01-01",
        "2024-01-01-1W to 2024-01-01",
        V1TimeGrain.TIME_GRAIN_WEEK,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],

      [
        "-24h",
        "Last 24 hours, incomplete",
        V1TimeGrain.TIME_GRAIN_HOUR,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label, rangeGrain, bucketGrain] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(rt.getLabel()).toEqual(label);
        expect(rt.getRangeGrain()).toEqual(rangeGrain);
        expect(rt.getBucketGrain()).toEqual(bucketGrain);

        const convertedRillTime = rt.toString();
        const convertedRillTimeParsed = parseRillTime(convertedRillTime);
        expect(convertedRillTimeParsed.toString()).toEqual(convertedRillTime);
      });
    }
  });

  describe("Update timezone", () => {
    const Cases: [rillTime: string, timezone: string, replaced: string][] = [
      ["-6d to latest by |h|", "UTC", "-6d to latest by |h|@{UTC}"],
      ["-6d to latest by |h| @ {IST}", "UTC", "-6d to latest by |h|@{UTC}"],
      ["-6d to latest by |h| @ now", "UTC", "-6d to latest by |h|@now {UTC}"],
      [
        "-6d to latest by |h| @ now {IST}",
        "UTC",
        "-6d to latest by |h|@now {UTC}",
      ],
    ];

    for (const [rillTime, timezone, replaced] of Cases) {
      it(`'${rillTime}'.addTimezone(${timezone})=${replaced}`, () => {
        const rt = parseRillTime(rillTime);
        rt.addTimezone(timezone);
        expect(rt.toString()).toEqual(replaced);
      });
    }
  });
});
