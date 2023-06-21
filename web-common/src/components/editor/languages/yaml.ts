import { StreamLanguage } from "@codemirror/language";
import * as yamlMode from "@codemirror/legacy-modes/mode/yaml";

const yaml = StreamLanguage.define(yamlMode.yaml);

export const yamlEditor = () => [yaml];
