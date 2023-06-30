import { StreamLanguage } from "@codemirror/language";
import * as yamlMode from "@codemirror/legacy-modes/mode/yaml";

// Note: typescript support for legacy modes is not great,
// so this will have to suffice.
//
// CodeMirror does not have modern support for YAML code highlighting,
// but this legacy approach appears to work fairly well. It doesn't have
// a robust way to actually parse the yaml, however, so we can't do much here
// unless we provide our own AST through another library.

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
export const yaml = () => [StreamLanguage.define(yamlMode.yaml)];
