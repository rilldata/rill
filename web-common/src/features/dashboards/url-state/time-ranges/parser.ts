import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function parseRillTime(rillTimeRange: string): RillTime {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(rillTimeRange);
  const rt = parser.results[0] as RillTime;
  if (!rt) throw new Error("Unknown error");
  return rt;
}

export function validateRillTime(rillTime: string): Error | undefined {
  try {
    const parser = parseRillTime(rillTime);
    if (!parser) return new Error("Unknown error");
  } catch (err) {
    return err;
  }
  return undefined;
}

/**
 * Convenience method to parse and rill time and return its label.
 */
export function getRillTimeLabel(rillTime: string): string {
  try {
    const rt = parseRillTime(rillTime);
    return rt.getLabel();
  } catch {
    return rillTime;
  }
}

/**
 * Overrides the ref part of a rill time range.
 * @param rt RillTime instance to override
 * @param refOverride Ref to override with, should be in the format of `watermark` or `watermark/Y` or `watermark/Y+1Y` etc
 */
export function overrideRillTimeRef(rt: RillTime, refOverride: string) {
  const overriddenRillTime = parseRillTime(`7D as of ${refOverride}`);
  const overriddenPoint = overriddenRillTime.anchorOverrides[0];
  if (!overriddenPoint) throw new Error("No anchor overrides found");
  rt.overrideRef(overriddenPoint);
}
