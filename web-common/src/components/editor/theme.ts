import { EditorView } from "@codemirror/view";
export const editorTheme = () =>
  EditorView.theme({
    "&.cm-editor": {
      overflowX: "hidden",
      width: "100%",
      height: "100%",
      "&.cm-focused": {
        outline: "none",
      },
    },
    ".cm-line.cm-line-error": {
      // this is tailwind bg-red-50
      backgroundColor: "#FEF2F2",
    },
    ".cm-line-error .ͼc, .cm-line-error .ͼe, ": {
      // this is tailwind text-red-900
      color: "#7F1D1D",
    },
    ".cm-line-level.cm-activeLine": {
      backgroundColor: "hsl(1,90%,80%)",
      color: "blue",
    },
    ".cm-line.cm-line-error.cm-activeLine": {
      // tailwind bg-red-200
      backgroundColor: "#FEE2E2",
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
      { backgroundColor: "rgb(65 99 255 / 25%)" },
    ".cm-selectionMatch": { backgroundColor: "rgb(189 233 255)" },
    ".cm-gutter": {
      backgroundColor: "white",
    },
    ".cm-gutters": {
      borderRight: "none",
    },
    ".cm-scroller": {
      fontFamily: "var(--monospace)",
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
    },
    ".cm-completionMatchedText": {
      textDecoration: "none",
      color: "rgb(15 119 204)",
    },
    ".cm-underline": {
      backgroundColor: "rgb(254 240 138)",
    },
  });
