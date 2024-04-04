type Description = string;
type Handler = (e: MouseEvent) => Promise<void> | void;
type Modifier = "command" | "shift" | "shift-command" | "click";

type Params = Partial<Record<Modifier, [Handler, Description]>>;

export function modifiedClick(node: HTMLElement, params: Params) {
  function update(params: Params) {
    //eslint-disable-next-line
    node.addEventListener("click", async (e) => {
      const { ctrlKey, shiftKey, metaKey } = e;

      let handler: Handler | undefined = undefined;

      if ((ctrlKey || metaKey) && shiftKey && params["shift-command"]) {
        e.preventDefault();
        handler = params["shift-command"][0];
      } else if ((ctrlKey || metaKey) && params.command) {
        e.preventDefault();
        handler = params.command[0];
      } else if (shiftKey && params.shift) {
        e.preventDefault();
        handler = params.shift[0];
      } else if (params.click) {
        handler = params.click[0];
      }

      if (handler) {
        console.log("yeah");

        await handler(e);
      }
    });

    const actions = Object.entries(params)
      .map(([key, value]) => `${key}:${value[1]}`)
      .join(",");

    node.setAttribute("data-actions", actions);
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
