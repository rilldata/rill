import { LanguageSupport, StreamLanguage } from "@codemirror/language";
import * as yamlMode from "@codemirror/legacy-modes/mode/yaml";

export const yaml = () => [
  new LanguageSupport(StreamLanguage.define(yamlMode.yaml)),
];
