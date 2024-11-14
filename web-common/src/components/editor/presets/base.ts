import { acceptCompletion } from "@codemirror/autocomplete";
import { indentWithTab } from "@codemirror/commands";
import { lineStatus } from "../line-status";
import { editorTheme } from "../theme";
import { indentationMarkers } from "@replit/codemirror-indentation-markers";
import {
  keymap,
  highlightSpecialChars,
  drawSelection,
  highlightActiveLine,
  dropCursor,
  rectangularSelection,
  crosshairCursor,
  highlightActiveLineGutter,
} from "@codemirror/view";
import { EditorState, Prec } from "@codemirror/state";
import {
  defaultHighlightStyle,
  syntaxHighlighting,
  indentOnInput,
  bracketMatching,
  foldKeymap,
} from "@codemirror/language";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { searchKeymap, highlightSelectionMatches } from "@codemirror/search";
import {
  autocompletion,
  completionKeymap,
  closeBrackets,
  closeBracketsKeymap,
} from "@codemirror/autocomplete";
import { lintKeymap } from "@codemirror/lint";

/**
 * https://github.com/codemirror/basic-setup/blob/main/src/codemirror.ts
 * https://codemirror.net/docs/ref/#codemirror.basicSetup
 *
 */

export const base = () => [
  // BASE EXTENSIONS
  // lineNumbers(),
  highlightActiveLineGutter(),
  highlightSpecialChars(),
  history(),
  // foldGutter(),
  drawSelection(),
  dropCursor(),
  EditorState.allowMultipleSelections.of(true),
  indentOnInput(),
  syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
  bracketMatching(),
  closeBrackets(),
  autocompletion(),
  rectangularSelection(),
  crosshairCursor(),
  highlightActiveLine(),
  highlightSelectionMatches(),
  keymap.of([
    ...closeBracketsKeymap,
    ...defaultKeymap,
    ...searchKeymap,
    ...historyKeymap,
    ...foldKeymap,
    ...completionKeymap,
    ...lintKeymap,
  ]),

  // ADDITIONAL EXTENSIONS
  editorTheme(),
  lineStatus(),
  indentationMarkers(),
  Prec.highest(
    keymap.of([
      {
        key: "Tab",
        run: acceptCompletion,
      },
    ]),
  ),
  keymap.of([indentWithTab]),
];
