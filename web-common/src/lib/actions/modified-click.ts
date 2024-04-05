type Description = string;
type Handler = (e: MouseEvent) => unknown;
type Modifier = "command" | "shift" | "shift-command" | "click";

type Params = Partial<Record<Modifier, [Handler | null, Description | null]>>;

export function modifiedClick(node: HTMLElement, params: Params) {
  function update(params: Params) {
    //eslint-disable-next-line
    node.addEventListener("click", async (e) => {
      const { ctrlKey, shiftKey, metaKey } = e;

      let handler: Handler | null = null;
      let modifier: Modifier | null = null;

      if ((ctrlKey || metaKey) && shiftKey && params["shift-command"]) {
        e.preventDefault();
        handler = params["shift-command"][0];
        modifier = "shift-command";
      } else if ((ctrlKey || metaKey) && params.command) {
        e.preventDefault();
        handler = params.command[0];
        modifier = "command";
      } else if (shiftKey && params.shift) {
        e.preventDefault();
        handler = params.shift[0];
        modifier = "shift";
      } else if (params.click) {
        handler = params.click[0];
      }

      if (handler) {
        await handler(e);
      } else {
        node.dispatchEvent(
          new CustomEvent(modifier ? `${modifier}-click` : "click"),
        );
      }
    });

    const actions = Object.entries(params)
      .map(([action, value]) => {
        return value[1] ? `${action}:${value[1]}` : null;
      })
      .filter((s) => s !== null)
      .join(",");

    if (actions.length) node.setAttribute("data-actions", actions);
  }

  function destroy() {
    if (node.parentNode) {
      node.parentNode.removeChild(node);
    }
  }

  void update(params);
  return {
    update,
    destroy,
  };
}
