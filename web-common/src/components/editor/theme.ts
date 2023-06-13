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
      // this appears to be the best option for interaction with selections.
      mixBlendMode: "hue",
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
      {
        backgroundColor: "hsla(214, 95%, 70%, 25%)",
        backgroundBlendMode: "hue",
      },
    // the color of the selectionMatch background
    ".cm-selectionMatch": {
      backgroundBlendMode: "multiply",
    },
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

    // key
    ".ͼc": {
      color: "hsl(263, 70%, 50%)",
    },
    // strings
    // note: we have to code .cm-line as well since CodeMirror does not seem to always wrap
    // strings in a classed span.
    ".ͼe, .cm-line": {
      color: "hsl(175, 77%, 26%)",
    },
    // decimal / number
    ".ͼd": {
      color: "hsl(224, 76%, 48%)",
    },
    // boolean
    ".ͼb": {
      color: "hsl(35, 92%, 33%)",
    },
    ".ͼ5": {
      color: "black",
    },
    // comment
    ".ͼm": {
      color: "hsl(215, 25%, 27%)",
    },
  });
