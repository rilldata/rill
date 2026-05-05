interface DragTableCellOptions {
  onresize?: (size: number) => void;
  onresizeend?: () => void;
}

export function dragTableCell(
  node: HTMLElement,
  options: DragTableCellOptions = {},
) {
  let moving = false;
  let opts = options;

  function mousedown() {
    moving = true;
  }

  function mousemove(e: MouseEvent) {
    if (moving) {
      const rect = node.parentElement!.getBoundingClientRect();
      opts.onresize?.(e.pageX - rect.left);
    }
  }

  function mouseup() {
    if (moving) {
      moving = false;
      opts.onresizeend?.();
    }
  }

  node.addEventListener("mousedown", mousedown);
  window.addEventListener("mousemove", mousemove);
  window.addEventListener("mouseup", mouseup);

  return {
    update(newOptions: DragTableCellOptions) {
      opts = newOptions;
      moving = false;
    },
    destroy() {
      node.removeEventListener("mousedown", mousedown);
      window.removeEventListener("mousemove", mousemove);
      window.removeEventListener("mouseup", mouseup);
    },
  };
}
