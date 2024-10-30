import grammar from "./expression.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function convertFilterParamToExpression(filter: string) {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(filter);
  return parser.results[0];
}

export function stripParserError(err: Error) {
  return err.message.substring(
    0,
    err.message.indexOf("Instead, I was expecting") - 1,
  );
}
