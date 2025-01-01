import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function parseRillTime(rillTime: string): RillTime {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(rillTime);
  return parser.results[0];
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
