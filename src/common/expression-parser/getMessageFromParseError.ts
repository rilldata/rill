import type { ParseExpressionError } from "$common/expression-parser/parseExpression";

export function getMessageFromParseError(
  expression: string,
  parseExpressionError: ParseExpressionError
) {
  const messages = new Array<string>();
  if (parseExpressionError.disallowedSyntax) {
    messages.push(
      `Disallowed syntax : ${parseExpressionError.disallowedSyntax}. ` +
        "Measure aggregation expression may only contain aggregation functions and arithmetic operations."
    );
  }
  if (parseExpressionError.missingColumns?.length) {
    messages.push(
      parseExpressionError.missingColumns.length > 1
        ? `Columns ${parseExpressionError.missingColumns
            .slice(0, parseExpressionError.missingColumns.length - 1)
            .join(", ")} ` +
            `and ${
              parseExpressionError.missingColumns[
                parseExpressionError.missingColumns.length - 1
              ]
            }` +
            ` are missing from model ${parseExpressionError.missingFrom}.`
        : `Column ${parseExpressionError.missingColumns[0]} is missing from model ${parseExpressionError.missingFrom}.`
    );
  }
  if (parseExpressionError.unexpectedToken) {
    messages.push(
      `Token "${
        parseExpressionError.unexpectedToken
      }" is not allowed at "${expression.substring(
        0,
        parseExpressionError.location.start
      )}\`==>${expression.substring(
        parseExpressionError.location.start,
        parseExpressionError.location.end
      )}<==\`${expression.substring(parseExpressionError.location.end)}".`
    );
  }
  if (messages.length === 0) {
    messages.push(parseExpressionError.message);
  }

  return messages.join(" ");
}
