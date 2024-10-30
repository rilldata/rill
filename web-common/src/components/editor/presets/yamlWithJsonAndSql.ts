import { jsonLanguage } from "@codemirror/lang-json";
import { yamlLanguage } from "@codemirror/lang-yaml";
import { parseMixed } from "@lezer/common";
import { DuckDBSQL } from "./duckDBDialect";

const activateOnNodes = new Set(["BlockLiteralContent"]);

let blockLiteralCount = 0;

const wrap = parseMixed((node) => {
  if (node?.name === "Stream") {
    blockLiteralCount = 0;
  }
  if (!node || !activateOnNodes.has(node.name)) {
    return null;
  }
  blockLiteralCount++;

  if (blockLiteralCount === 1) {
    return {
      parser: DuckDBSQL.language.parser,
    };
  } else if (blockLiteralCount === 2) {
    return {
      parser: jsonLanguage.parser,
    };
  }
  return null;
});

const names = new Set(["Literal", ":"]);
let foundExpression = false;
let foundColon = false;

const metricsParsing = parseMixed(({ name, from, to }, input) => {
  if (!names.has(name)) return null;

  if (
    !foundExpression &&
    name === "Literal" &&
    input.read(from, to) === "expression"
  ) {
    foundExpression = true;
    return null;
  }

  if (name === ":") {
    foundColon = true;
    return null;
  }

  if (foundExpression && foundColon && name === "Literal") {
    foundExpression = false;
    foundColon = false;
    return {
      parser: DuckDBSQL.language.parser,
    };
  }
  return null;
});

export const customYAMLwithJSONandSQL = yamlLanguage.configure({
  wrap,
});

export const metricsPlusSQL = yamlLanguage.configure({
  wrap: metricsParsing,
});
