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
        : this.content,
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

/* Extension that enables a simple placeholder for the metrics editor. **/
export function createPlaceholder(metricsViewName: string, showOnEmpty = true) {
  const component = createPlaceholderElement(metricsViewName);
  const elem = component.DOMElement;
  const extension = ViewPlugin.fromClass(
    class {
      placeholder: DecorationSet;

      constructor(readonly view: EditorView) {
        this.placeholder = Decoration.set([
          Decoration.widget({
            widget: new PlaceholderWidget(elem),
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
    { decorations: (v) => v.decorations },
  );
  return { component, extension };
}

/** creates a set of callbacks that enables updating the placeholder,
 * which itself is a svelte component.
 */
export function createPlaceholderElement(metricsName: string) {
  const DOMElement = document.createElement("span");
  // create placeholder text and attach it to the DOM element.
  const component = new Placeholder({
    target: DOMElement,
    props: {
      metricsName,
      view: undefined,
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
