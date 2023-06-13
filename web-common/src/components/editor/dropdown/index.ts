import type { EditorState } from "@codemirror/basic-setup";
import { StateEffect, StateField } from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  EditorView,
  ViewPlugin,
  ViewUpdate,
  WidgetType,
} from "@codemirror/view";

export interface Option {
  label: string;
  key: string;
}

export const updateOptions = StateEffect.define<{
  options: Array<Option>;
}>({
  map: (value) => value,
});

export const optionsStateField = StateField.define({
  create: () => [],
  update: (lines, tr) => {
    // Handle transactions with the updateLineState effect
    for (const effect of tr.effects) {
      if (effect.is(updateOptions)) {
        // Clear the existing errors and set the new errors
        return effect.value.options.slice();
      }
    }

    return lines;
  },
  compare: (a, b) => a === b,
});

export const fieldDropdown = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet;

    constructor(view: EditorView) {
      this.decorations = this.getDecorations(view.state);
    }

    update(update: ViewUpdate) {
      // FIXME: add check for state update.
      if (update.docChanged || update.selectionSet || update.viewportChanged) {
        this.decorations = this.getDecorations(update.state);
      }
    }

    getDecorations(state: EditorState) {
      // retrieve the current options here.
      const options = state.field(optionsStateField);

      const decorations = [];
      const regex = /property: .*/g; // Adjust this regex according to your needs
      let match;
      // Find all the lines containing the 'property: ' keyword
      while ((match = regex.exec(state.doc.toString())) != null) {
        // Add a widget decoration (button) at the end of each line
        const { number } = state.doc.lineAt(match.index);

        // Then find the position at the end of that line to place the widget
        const endPos = state.doc.line(number).to;
        decorations.push(
          Decoration.widget({
            widget: new Dropdown(options),
            side: 1,
          }).range(endPos)
        );
      }
      return Decoration.set(decorations);
    }
  },
  {
    decorations: (v) => v.decorations,
  }
);

class Dropdown extends WidgetType {
  options: Array<Option>;
  constructor(options) {
    super();
    this.options = options;
  }

  toDOM() {
    const wrap = document.createElement("span");
    wrap.setAttribute("aria-hidden", "true");
    wrap.className = "cm-boolean-toggle";
    const box = wrap.appendChild(document.createElement("input"));
    box.type = "checkbox";
    return wrap;
  }
}

export function createDropdownPlugin() {
  return {
    /** creates an updater function that closes over
     * the options and provides them.
     */
    createUpdater(options) {
      return (view) => {
        const transaction = updateOptions.of({
          options,
        });
        view.dispatch({
          effects: [transaction],
        });
      };
    },
    extension: [optionsStateField, fieldDropdown],
  };
}
