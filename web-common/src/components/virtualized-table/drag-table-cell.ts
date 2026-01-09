export function dragTableCell(node) {
  let moving = false;

  function mousedown() {
    moving = true;
  }

  function mousemove(e: MouseEvent) {
    if (moving) {
      const rect = node.parentNode.getBoundingClientRect();
      const left = rect.left;

      node.dispatchEvent(
        new CustomEvent("resize", {
          detail: e.pageX - left,
        }),
      );
    }
  }

  function mouseup() {
    if (moving) {
      moving = false;
      node.dispatchEvent(new CustomEvent("resizeend"));
    }
  }

  node.addEventListener("mousedown", mousedown);
  window.addEventListener("mousemove", mousemove);
  window.addEventListener("mouseup", mouseup);

  return {
    update() {
      moving = false;
    },
    destroy() {
      node.removeEventListener("mousedown", mousedown);
      window.removeEventListener("mousemove", mousemove);
      window.removeEventListener("mouseup", mouseup);
    },
  };
}
