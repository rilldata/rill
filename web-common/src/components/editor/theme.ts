import { EditorView } from "@codemirror/view";
export const editorTheme = () =>
  EditorView.theme({
    "&.cm-editor": {
      overflowX: "hidden",
      width: "100%",
      height: "100%",
      fontWeight: "500",
      "&.cm-focused": {
        outline: "none",
      },
    },
    ".cm-line.cm-line-error": {
      // this is tailwind bg-red-50
      backgroundColor: "#FEF2F2",
      // this appears to be the best option for interaction with selections.
      mixBlendMode: "hue",
      color: "var(--ui-editor-error-foreground)",
    },

    // hovered line error styling on selections need this
    ".cm-selectionLayer .cm-selectionBackground": {
      // This is unfortunate but it does not seem possible to otherwise select
      // the right element when the selection is focused.
      backgroundColor: "red",
    },

    ".cm-line-level.cm-activeLine": {
      backgroundColor: "red",
    },

    ".cm-activeLine": {
      backgroundColor: "var(--gray-2)",
    },

    ".cm-line.cm-line-error.cm-activeLine": {
      // tailwind bg-red-200
      backgroundColor: "red",
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
      {
        backgroundColor: "red",
        backgroundBlendMode: "hue",
      },

    // the color of the selectionMatch background
    ".cm-selectionMatch": {
      backgroundBlendMode: "multiply",
    },
    ".cm-gutter": {
      backgroundColor: "hsl(var(--background))",
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
      backgroundColor: "hsl(var(--background))",
    },
    ".cm-gutterElement.cm-activeLineGutter": {
      // backgroundColor: "hsla(214, 95%, 70%, 25%)",
      // backgroundBlendMode: "hue",
      backgroundColor: "hsl(var(--background))",
    },
    ".cm-tooltip": {
      borderRadius: "0.25rem",
      padding: "0.5rem",
      color: "var(--gray-accent)",
      backgroundColor: "var(--gray-2)",
      borderColor: "var(--gray-3)",
    },
    ".cm-tooltip-autocomplete": {
      backgroundColor: "red",
      border: "none",
      "& > ul > li[aria-selected]": {
        border: "none",
        borderRadius: "0.25rem",
        backgroundColor: "red",
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
      // backgroundColor: "rgb(254 240 138)",
    },

    // key
    ".ͼc": {
      color: "var(--code-key)",
    },
    // strings
    // note: we have to code .cm-line as well since CodeMirror does not seem to always wrap
    // strings in a classed span.
    ".cm-line": {
      color: "var(--code-green)",
    },
    ".ͼe": {
      color: "var(--code-blue)",
    },
    ".ͼl": {
      color: "var(--code-purple)",
    },
    // decimal / number
    ".ͼd": {
      color: "hsl(224, 76%, 48%)",
    },
    // boolean
    ".ͼb": {
      color: "var(--code-orange)",
    },
    ".ͼ5": {
      color: "var(--gray-12)",
    },
    // comment
    ".ͼm": {
      color: "var(--code-subtle)",
    },

    // make sure all the above colors are plain red when there is an error.
    ".cm-line-error .ͼc, .cm-line-error .ͼ5, .cm-line-error .ͼb, .cm-line-error .ͼd, .cm-line-error .ͼe, .cm-line-error .ͼm":
      {
        color: "var(--ui-editor-error-foreground)",
      },
  });
