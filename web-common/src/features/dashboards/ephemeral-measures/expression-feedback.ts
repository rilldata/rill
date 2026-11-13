import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import {
  EPHEMERAL_MEASURE_FUNCTIONS,
  parseMeasureExpression,
} from "./expression-parser";

/**
 * Describes what is wrong with an ephemeral measure expression, in the terms
 * the editor shows: measures by display name and plain-language fixes for the
 * common mistakes. The parser's own messages name internal field names and
 * tokens, so they are only used for rare errors that have no description here.
 * Returns undefined for an empty or valid expression.
 *
 * `measures` are the measures the expression may reference.
 */
export function describeExpressionError(
  expression: string,
  measures: MetricsViewSpecMeasure[],
): string | undefined {
  if (expression.trim() === "") return undefined;

  const parsed = parseMeasureExpression(expression);
  const { error } = parsed;
  if (!error) {
    const unknownRef = parsed.refs.find(
      (ref) => !measures.some((mes) => mes.name === ref),
    );
    return unknownRef
      ? m.dashboard_pivot_ephemeral_error_unknown_measure({ name: unknownRef })
      : undefined;
  }

  switch (error.kind) {
    case "unexpected_token": {
      // The grammar only rejects an operand where the previous operand has
      // just ended, so an operand here is always missing an operator before it.
      const operand = operandAt(expression, error.position);
      if (operand === undefined) {
        return m.dashboard_pivot_ephemeral_error_invalid();
      }
      const measure = measures.find((mes) => mes.name === operand);
      return m.dashboard_pivot_ephemeral_error_missing_operator({
        operand: measure?.displayName || operand,
      });
    }
    case "incomplete":
      return m.dashboard_pivot_ephemeral_error_incomplete();
    case "unclosed_parenthesis":
      return m.dashboard_pivot_ephemeral_error_unclosed_parenthesis();
    case "no_measure":
      return m.dashboard_pivot_ephemeral_error_no_measure();
    case "unsupported_function": {
      const name = operandAt(expression, error.position) ?? "";
      // A measure followed by "(" parses as a call, e.g. a chip and then "(b)".
      if (measures.some((mes) => mes.name === name)) {
        return m.dashboard_pivot_ephemeral_error_missing_operator({
          operand: "(",
        });
      }
      return m.dashboard_pivot_ephemeral_error_unsupported_function({
        name,
        functions: Object.keys(EPHEMERAL_MEASURE_FUNCTIONS).join(", "),
      });
    }
    default:
      return error.message.charAt(0).toUpperCase() + error.message.slice(1);
  }
}

/**
 * Returns the operand that starts at `position`: an identifier, a quoted
 * identifier (unquoted), a number or an opening parenthesis.
 */
function operandAt(expression: string, position: number): string | undefined {
  const rest = expression.slice(position);
  const quoted = /^"((?:[^"]|"")*)"/.exec(rest);
  if (quoted) return quoted[1].replace(/""/g, '"');
  return (
    /^[A-Za-z_][A-Za-z0-9_]*/.exec(rest)?.[0] ??
    /^(?:\d+\.?\d*|\.\d+)/.exec(rest)?.[0] ??
    (rest.startsWith("(") ? "(" : undefined)
  );
}
