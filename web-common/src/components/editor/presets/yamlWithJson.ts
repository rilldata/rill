import { jsonLanguage } from "@codemirror/lang-json";
import { yamlLanguage } from "@codemirror/lang-yaml";
import { LRLanguage } from "@codemirror/language";
import { parseMixed } from "@lezer/common";

const yamlWithJSONParser = yamlLanguage.parser.configure({
  wrap: parseMixed((node) => {
    console.log(node.name, node.type.name, node);
    if (node.name === "BlockLiteralContent") {
      return { parser: jsonLanguage.parser };
    }
    return null;
  }),
});

export const customYAMLWithJSON = LRLanguage.define({
  parser: yamlWithJSONParser,
});
