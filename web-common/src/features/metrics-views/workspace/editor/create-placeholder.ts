import type { Extension } from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  EditorView,
  ViewPlugin,
  WidgetType,
} from "@codemirror/view";
import Placeholder from "./Placeholder.svelte";

class PlaceholderWidget extends WidgetType {
  constructor(readonly content: string | HTMLElement) {
    super();
  }

  toDOM() {
    const wrap = document.createElement("span");
    wrap.className = "cm-placeholder";
    //wrap.style.pointerEvents = "none";
    wrap.appendChild(
      typeof this.content == "string"
        ? document.createTextNode(this.content)
        : this.content
    );
    if (typeof this.content == "string")
      wrap.setAttribute("aria-label", "placeholder " + this.content);
    else wrap.setAttribute("aria-hidden", "true");
    return wrap;
  }

  ignoreEvent() {
    return false;
  }
}

/// Extension that enables a placeholder—a piece of example content
/// to show when the editor is empty.
export function createPlaceholder(
  content: string | HTMLElement,
  showOnEmpty = true
): Extension {
  return ViewPlugin.fromClass(
    class {
      placeholder: DecorationSet;

      constructor(readonly view: EditorView) {
        this.placeholder = Decoration.set([
          Decoration.widget({
            widget: new PlaceholderWidget(content),
            side: 1,
          }).range(0),
        ]);
      }

      update!: () => void; // Kludge to convince TypeScript that this is a plugin value

      get decorations() {
        return showOnEmpty
          ? this.view.state.doc.length
            ? Decoration.none
            : this.placeholder
          : this.placeholder;
      }
    },
    { decorations: (v) => v.decorations }
  );
}

export function createPlaceholderElement(metricsName: string) {
  const DOMElement = document.createElement("span");
  // create placeholder text and attach it to the DOM element.
  const component = new Placeholder({
    target: DOMElement,
    props: {
      metricsName,
    },
  });
  return {
    DOMElement,
    setEditorView(view: EditorView) {
      component.$set({ metricsName, view });
    },
    on(event, callback) {
      component.$on(event, callback);
    },
  };
}
