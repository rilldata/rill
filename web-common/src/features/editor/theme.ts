import { EditorView } from "@codemirror/view";

// TODO: All these hardcoded colors ain't good. Try to use Tailwind colors.
// Might have to navigated CodeMirror generated classes.

const highlightBackground = "#f3f9ff";

export const rillTheme = EditorView.theme({
  "&.cm-editor": {
    overflowX: "hidden",
    width: "100%",
    fontSize: "13px",
    height: "100%",
    "&.cm-focused": {
      outline: "none",
    },
  },
  ".cm-scroller": {
    fontFamily: "var(--monospace)",
  },
  "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
    { backgroundColor: "rgb(65 99 255 / 25%)" },
  ".cm-selectionMatch": { backgroundColor: "rgb(189 233 255)" },
  ".cm-activeLine": { backgroundColor: highlightBackground },

  ".cm-activeLineGutter": {
    backgroundColor: highlightBackground,
  },
  ".cm-gutters": {
    backgroundColor: "white",
    borderRight: "none",
  },
  ".cm-lineNumbers .cm-gutterElement": {
    paddingLeft: "5px",
    paddingRight: "10px",
    minWidth: "32px",
    backgroundColor: "white",
  },
  ".cm-breakpoint-gutter .cm-gutterElement": {
    color: "red",
    paddingLeft: "24px",
    paddingRight: "24px",
    cursor: "default",
  },
  ".cm-tooltip": {
    border: "none",
    borderRadius: "0.25rem",
    backgroundColor: "rgb(243 249 255)",
    color: "black",
  },
  ".cm-tooltip-autocomplete": {
    "& > ul > li[aria-selected]": {
      border: "none",
      borderRadius: "0.25rem",
      backgroundColor: "rgb(15 119 204 / .25)",
      color: "black",
    },
  },
  ".cm-completionLabel": {
    fontSize: "13px",
    fontFamily: "var(--monospace)",
  },
  ".cm-completionMatchedText": {
    textDecoration: "none",
    color: "rgb(15 119 204)",
  },
  ".cm-underline": {
    backgroundColor: "rgb(254 240 138)",
  },
  ".ͼb": {
    fontWeight: "700",
  },
  ".ͼe": {
    fontStyle: "italic",
    fontWeight: "600",
    color: "hsl(200, 70%, 50%)",
  },
});
