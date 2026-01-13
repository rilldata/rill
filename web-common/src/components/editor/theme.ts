import { EditorView } from "@codemirror/view";
import type { Extension } from "@codemirror/state";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags as t } from "@lezer/highlight";

const blue = "var(--color-blue-800)";
const purple = "var(--color-purple-700)";
const invalid = "var(--color-red-600)";
const emerald = "var(--color-emerald-700)";
const gray = "var(--muted-foreground)";
const amber = "var(--color-amber-600)";
const highlightBackground = "var(--line-highlight)";
const background = "var(--surface)";
const tooltipBackground = "var(--popover)";
const selection = "var(--editor-selection)";
const cursor = "var(--color-gray-800)";
const orange = "var(--color-orange-700)";

export const editorTheme = EditorView.theme(
  {
    "&": {
      color: emerald,
      backgroundColor: background,
    },

    ".cm-content": {
      caretColor: cursor,
    },
    "&.cm-editor": {
      overflowX: "hidden",
      width: "100%",
      height: "100%",
      fontWeight: "500",
      "&.cm-focused": {
        outline: "none",
      },
    },

    ".cm-cursor, .cm-dropCursor": { borderLeftColor: cursor },
    "&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection":
      {
        background: selection,
      },

    ".cm-panels": { backgroundColor: background, color: emerald },
    ".cm-panels.cm-panels-top": { borderBottom: "2px solid black" },
    ".cm-panels.cm-panels-bottom": { borderTop: "2px solid black" },

    ".cm-searchMatch": {
      backgroundColor: "#72a1ff59",
      outline: "1px solid #457dff",
    },
    ".cm-searchMatch.cm-searchMatch-selected": {
      backgroundColor: selection,
    },
    "&.cm-editor .cm-scroller": {
      fontFamily: "var(--monospace)",
    },
    ".cm-activeLine": {
      backgroundColor: highlightBackground,
    },
    ".cm-selectionMatch": { backgroundColor: selection },

    "&.cm-focused .cm-matchingBracket, &.cm-focused .cm-nonmatchingBracket": {
      backgroundColor: "#bad0f847",
    },

    ".cm-gutters": {
      backgroundColor: background,
      color: gray,
      border: "none",
    },

    ".cm-indent-markers::before": {
      backgroundImage:
        "repeating-linear-gradient(to right, var(--color-gray-300) 0px, var(--color-gray-300) 1px, transparent 1px, transparent 2ch)",
    },

    ".cm-activeLineGutter": {
      backgroundColor: highlightBackground,
    },

    ".cm-foldPlaceholder": {
      backgroundColor: "transparent",
      border: "none",
      color: "#ddd",
    },

    // ".cm-tooltip": {
    //   border: "var(--surface)",
    //   color: emerald,
    //   backgroundColor: tooltipBackground,
    // },

    ".cm-tooltip": {
      border: "solid 1px var(--color-gray-400)",
      borderRadius: "0.25rem",
      padding: "0.5rem",
      color: "var(--color-gray-800)",
      backgroundColor: tooltipBackground,
    },
    ".cm-tooltip .cm-tooltip-arrow:before": {
      borderTopColor: "transparent",
      borderBottomColor: "transparent",
    },
    ".cm-tooltip .cm-tooltip-arrow:after": {
      borderTopColor: tooltipBackground,
      borderBottomColor: tooltipBackground,
    },
    ".cm-tooltip-autocomplete": {
      "& > ul > li[aria-selected]": {
        backgroundColor: highlightBackground,
        color: emerald,
      },
    },
  },
  { dark: true },
);

/// The highlighting style for code in the One Dark theme.
export const highlightStyle = HighlightStyle.define([
  { tag: t.keyword, color: amber },
  {
    tag: [t.deleted, t.character, t.propertyName, t.macroName],
    color: purple,
  },
  { tag: [t.function(t.variableName), t.labelName], color: purple },
  { tag: [t.color, t.constant(t.name), t.standard(t.name)], color: purple },
  { tag: [t.definition(t.name), t.separator], color: emerald },
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
    color: blue,
  },
  {
    tag: [
      // t.operator,
      t.operatorKeyword,
      t.url,
      t.escape,
      t.regexp,
      t.link,
      t.special(t.string),
    ],
    color: orange,
  },
  { tag: [t.meta, t.comment], color: gray },
  { tag: t.strong, fontWeight: "bold" },
  { tag: t.emphasis, fontStyle: "italic" },
  { tag: t.strikethrough, textDecoration: "line-through" },
  { tag: t.link, color: purple, textDecoration: "underline" },
  { tag: t.heading, fontWeight: "bold", color: purple },
  { tag: [t.atom, t.bool, t.special(t.variableName)], color: purple },
  { tag: [t.processingInstruction, t.string, t.inserted], color: blue },
  { tag: t.invalid, color: invalid },
]);

export const theme: Extension = [
  editorTheme,
  syntaxHighlighting(highlightStyle),
];
