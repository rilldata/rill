import { jsonLanguage } from "@codemirror/lang-json";
import { yamlLanguage } from "@codemirror/lang-yaml";
import { LRLanguage } from "@codemirror/language";
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

const customYAMLandSQLParser = yamlLanguage.parser.configure({
  wrap,
});

export const customYAMLwithJSONandSQL = LRLanguage.define({
  parser: customYAMLandSQLParser,
});
