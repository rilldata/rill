import {
  GridStack,
  type GridItemHTMLElement,
  type GridStackOptions,
} from "gridstack";

export class GridStackManager {
  private instance: GridStack | null = null;
  private options: GridStackOptions;

  constructor(readonly: boolean) {
    this.options = {
      // TODO: set this to 1 in smaller screens
      column: 12,
      resizable: {
        handles: "e,se,s,sw,w",
      },
      animate: true,
      float: true,
      staticGrid: readonly,
    };
  }

  initialize(container: HTMLElement) {
    if (!this.instance) {
      this.instance = GridStack.init(this.options, container);

      // See: https://github.com/gridstack/gridstack.js/tree/master/doc#resizestartevent-el
      this.instance.on(
        "resizestart",
        (event: Event, el: GridItemHTMLElement) => {
          console.log("Resize started:", el);
        },
      );

      // See: https://github.com/gridstack/gridstack.js/tree/master/doc#resizeevent-el
      this.instance.on("resize", (event: Event, el: GridItemHTMLElement) => {
        console.log("Resizing:", el);
        const { w, h } = el.gridstackNode || {};
        console.log("Current size:", { w, h });
      });

      // See: https://github.com/gridstack/gridstack.js/tree/master/doc#resizestopevent-el
      this.instance.on(
        "resizestop",
        (event: Event, el: GridItemHTMLElement) => {
          console.log("Resize stopped:", el);
          const { w, h, x, y } = el.gridstackNode || {};
          console.log("Final dimensions:", { x, y, w, h });
        },
      );
    }

    return this.instance;
  }

  destroy() {
    if (this.instance) {
      this.instance.destroy();
      this.instance = null;
    }
  }
}
