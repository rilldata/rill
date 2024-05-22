import { acceptCompletion } from "@codemirror/autocomplete";
import { indentWithTab, insertNewline } from "@codemirror/commands";
import { EditorState, Prec } from "@codemirror/state";
import { keymap } from "@codemirror/view";
import { lineStatus } from "../line-status";
import { editorTheme } from "../theme";
import { basicSetup } from "codemirror";
import { indentationMarkers } from "@replit/codemirror-indentation-markers";

/** the base extension adds
 * - a theme
 * - the line status extension
 * - the indent guide extension
 * - a bunch of useful code mirror extensions
 *
 * This is the extension you should use for most editors.
 */
export const base = () => [
  basicSetup,
  editorTheme(),
  // lineStatus(),
  indentationMarkers(),

  Prec.high(
    keymap.of([
      {
        key: "Enter",
        run: insertNewline,
      },
    ]),
  ),
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
