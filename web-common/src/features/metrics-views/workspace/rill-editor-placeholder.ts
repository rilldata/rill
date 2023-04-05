import { Extension } from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  EditorView,
  ViewPlugin,
  WidgetType,
} from "@codemirror/view";
import EditorPlaceholderText from "./EditorPlaceholderText.svelte";

class Placeholder extends WidgetType {
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
export function rillEditorPlaceholder(
  content: string | HTMLElement,
  showOnEmpty = false
): Extension {
  return ViewPlugin.fromClass(
    class {
      placeholder: DecorationSet;

      constructor(readonly view: EditorView) {
        this.placeholder = Decoration.set([
          Decoration.widget({
            widget: new Placeholder(content),
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

export function createPlaceholderElement(content) {
  const DOMElement = document.createElement("span");
  // create placeholder text and attach it to the DOM element.
  const component = new EditorPlaceholderText({
    target: DOMElement,
    props: {
      content,
    },
  });
  return {
    DOMElement,
    set(content) {
      component.$set({ content });
    },
    on(event, callback) {
      component.$on(event, callback);
    },
  };
}
