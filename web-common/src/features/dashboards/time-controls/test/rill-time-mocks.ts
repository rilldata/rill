/**
 * The time range summary every time filter test runs against.
 * `latest` in a rilltime expression is `max`, which deliberately sits mid hour so that a snap to a
 * grain moves the anchor.
 */
export const TIME_RANGE_SUMMARY = {
  min: "2024-01-01T00:00:00Z",
  max: "2024-03-31T14:30:00Z",
};

/** The range the yaml preset starts the dashboard on. */
export const DEFAULT_TIME_RANGE = "7D as of latest/D+1D";

/** The ranges the yaml offers, which is what the time range dropdown lists. */
export const YAML_TIME_RANGES = [{ range: "24h" }, { range: "4W" }];

/**
 * The intervals the runtime resolves rilltime expressions to, keyed by expression.
 *
 * The runtime does the resolving, so a test cannot derive these; every expression its interactions
 * can produce needs an entry here. The intervals below are `latest` snapped to the grain of the
 * `as of` clause, offset by its padding, and then walked back by the range.
 */
export const RESOLVED_RILL_TIMES: Record<
  string,
  { start: string; end: string }
> = {
  // latest/D+1D is 2024-04-01T00:00:00Z, minus 7 days.
  [DEFAULT_TIME_RANGE]: {
    start: "2024-03-25T00:00:00.000Z",
    end: "2024-04-01T00:00:00.000Z",
  },
  // The same anchor, minus 4 weeks.
  "4W as of latest/D+1D": {
    start: "2024-03-04T00:00:00.000Z",
    end: "2024-04-01T00:00:00.000Z",
  },
  // latest/h+1h is 2024-03-31T15:00:00Z, minus 24 hours.
  "24h as of latest/h+1h": {
    start: "2024-03-30T15:00:00.000Z",
    end: "2024-03-31T15:00:00.000Z",
  },
};

/** The time range a test expects once `timeRange` is applied, resolved interval included. */
export function resolvedTimeRange(timeRange: string) {
  const resolved = RESOLVED_RILL_TIMES[timeRange];
  if (!resolved) {
    throw new Error(`No resolved interval mocked for "${timeRange}"`);
  }
  return { name: timeRange, ...resolved };
}
