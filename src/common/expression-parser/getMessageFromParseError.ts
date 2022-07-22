import type { ParseExpressionError } from "$common/expression-parser/parseExpression";
import { AllowedAggregates } from "$common/expression-parser/validateAggregate";
import { TypesMap } from "$lib/duckdb-data-types";

export function getMessageFromParseError(
  expression: string,
  parseExpressionError: ParseExpressionError
) {
  const messages = new Array<string>();

  if (parseExpressionError.disallowedSyntax) {
    messages.push(getDisallowedSyntaxError(parseExpressionError));
  }

  if (parseExpressionError.missingColumns?.length) {
    messages.push(getMissingColumnsError(parseExpressionError));
  }

  if (parseExpressionError.unexpectedToken) {
    messages.push(getInvalidTokenError(expression, parseExpressionError));
  }

  if (parseExpressionError.invalidAggregates?.length) {
    messages.push(...getInvalidAggregatedError(parseExpressionError));
  }

  if (messages.length === 0) {
    messages.push(parseExpressionError.message);
  }

  return messages.join("\n");
}

function getDisallowedSyntaxError(
  parseExpressionError: ParseExpressionError
): string {
  return (
    `Disallowed syntax: "${parseExpressionError.disallowedSyntax}". ` +
    "Measure aggregation expression may only contain aggregation functions and arithmetic operations"
  );
}

function getMissingColumnsError(
  parseExpressionError: ParseExpressionError
): string {
  return (
    (parseExpressionError.missingColumns.length > 1
      ? `Columns "${parseExpressionError.missingColumns
          .slice(0, parseExpressionError.missingColumns.length - 1)
          .join('", "')}" ` +
        `and "${
          parseExpressionError.missingColumns[
            parseExpressionError.missingColumns.length - 1
          ]
        }"` +
        ` are `
      : `Column "${parseExpressionError.missingColumns[0]}" is `) +
    `missing from model ${parseExpressionError.missingFrom}`
  );
}

function getInvalidTokenError(
  expression: string,
  parseExpressionError: ParseExpressionError
): string {
  return `Token "${
    parseExpressionError.unexpectedToken
  }" is not allowed at "${expression.substring(
    0,
    parseExpressionError.location.start
  )}\`==>${expression.substring(
    parseExpressionError.location.start,
    parseExpressionError.location.end
  )}<==\`${expression.substring(parseExpressionError.location.end)}".`;
}

function getInvalidAggregatedError(
  parseExpressionError: ParseExpressionError
): Array<string> {
  return parseExpressionError.invalidAggregates.map((invalidAggregate) => {
    if (invalidAggregate.aggregateNotAllowed) {
      return `"${invalidAggregate.name}" is not a valid aggregate function`;
    }
    return (
      "Invalid aggregate arguments. " +
      `Expected: ${invalidAggregate.name}(${AllowedAggregates[
        invalidAggregate.name
      ]
        .map((argTypeSet) =>
          !argTypeSet ? "UNKNOWN" : TypesMap.get(argTypeSet)
        )
        .join(", ")}) ` +
      `Actual: ${invalidAggregate.name}(${invalidAggregate.invalidArgs.join(
        ", "
      )})`
    );
  });
}
