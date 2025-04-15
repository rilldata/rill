import { HighlightStyle } from "@codemirror/language";
import { EditorView } from "@codemirror/view";
import { tags as t } from "@lezer/highlight";

const tooltipBackground = "var(--color-gray-100)";
const textColor = "var(--color-green-700)";
const background = "var(--color-gray-50)";
const cursor = "var(--color-gray-900)";
const selection = "hsla(214, 95%, 70%, 25%)";

export const editorTheme = () =>
  EditorView.theme(
    {
      "&": {
        color: textColor,
        backgroundColor: background,
        fontWeight: "500",
      },

      "&.cm-editor": {
        overflowX: "hidden",
        width: "100%",
        height: "100%",
        fontWeight: "500",
        background: "var(--surface)",
        "&.cm-focused": {
          outline: "none",
        },
      },

      ".cm-content": {
        caretColor: cursor,
      },

      ".cm-scroller": {
        fontFamily: "var(--monospace)",
      },

      ".cm-cursor, .cm-dropCursor": { borderLeftColor: cursor },
      "&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
        {
          backgroundColor: selection,
          backgroundBlendMode: "hue",
        },

      ".cm-panels": { backgroundColor: background },
      // ".cm-panels.cm-panels-top": { borderBottom: "2px solid black" },
      // ".cm-panels.cm-panels-bottom": { borderTop: "2px solid black" },

      ".cm-searchMatch": {
        backgroundColor: "#72a1ff59",
        outline: "1px solid #457dff",
      },
      ".cm-searchMatch.cm-searchMatch-selected": {
        backgroundColor: "#6199ff2f",
      },

      ".cm-activeLine": {
        backgroundColor: "var(--color-blue-50)",
      },
      "&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
        {
          backgroundColor: "hsla(214, 95%, 70%, 25%)",
          backgroundBlendMode: "hue",
        },
      ".cm-selectionMatch": { backgroundColor: "var(--color-blue-100)" },

      "&.cm-focused .cm-matchingBracket, &.cm-focused .cm-nonmatchingBracket": {
        backgroundColor: "#bad0f847",
      },

      ".cm-gutters": {
        backgroundColor: "var(--surface)",
        color: "var(--color-gray-700)",
        border: "none",
      },

      ".cm-activeLineGutter": {
        backgroundColor: "var(--color-blue-50)",
      },

      ".cm-foldPlaceholder": {
        backgroundColor: "transparent",
        border: "none",
        color: "#ddd",
      },

      ".cm-tooltip": {
        border: "solid 1px var(--color-gray-400)",
        borderRadius: "0.25rem",
        padding: "0.5rem",
        backgroundColor: tooltipBackground,
        color: "var(--color-gray-800)",
      },
      ".cm-tooltip .cm-tooltip-arrow:before": {
        borderTopColor: "transparent",
        borderBottomColor: "transparent",
        // border: "solid 6px black",
      },
      ".cm-tooltip .cm-tooltip-arrow:after": {
        borderTopColor: "transparent",
        borderBottomColor: "transparent",
        // border: "solid 6px red",
      },
      ".cm-tooltip-autocomplete": {
        "& > ul > li[aria-selected]": {
          backgroundColor: tooltipBackground,
          color: "textColor",
        },
      },
    },
    { dark: false },
  );

export const oneDarkHighlightStyle = HighlightStyle.define([
  { tag: t.keyword, color: "var(--color-amber-700)" },
  {
    tag: [t.deleted, t.character, t.propertyName, t.macroName],
    color: "var(--color-violet-700)",
  },

  {
    tag: [t.function(t.variableName), t.labelName],
    color: "var(--color-violet-700)",
  },
  {
    tag: [t.color, t.constant(t.name), t.standard(t.name)],
    color: "var(--color-violet-700)",
  },
  { tag: [t.definition(t.name), t.separator], color: "var(--color-green-700)" },
  {
    tag: [
      t.typeName,
      t.className,
      t.number,
      t.changed,
      t.annotation,
      t.modifier,
      t.self,
      t.namespace,
    ],
    color: "var(--color-blue-800)",
  },

  { tag: [t.meta, t.comment], color: "var(--color-gray-700)" },
  { tag: t.strong, fontWeight: "bold" },
  { tag: t.emphasis, fontStyle: "italic" },
  { tag: t.strikethrough, textDecoration: "line-through" },
  {
    tag: t.link,
    color: "var(--color-violet-700)",
    textDecoration: "underline",
  },
  { tag: t.heading, fontWeight: "bold", color: "var(--color-violet-700)" },
  {
    tag: [t.atom, t.bool, t.special(t.variableName)],
    color: "var(--color-violet-700)",
  },
  {
    tag: [t.processingInstruction, t.string, t.inserted],
    color: "var(--color-blue-800)",
  },
  { tag: t.invalid, color: "var(--color-violet-700)" },
]);
