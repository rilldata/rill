import { GridStack, type GridStackOptions } from "gridstack";

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

  initialize() {
    if (!this.instance) {
      this.instance = GridStack.init(this.options);
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
