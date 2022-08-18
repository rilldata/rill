export function dragTableCell(node) {
  let moving = false;
  function mousedown() {
    moving = true;
  }

  function mousemove(e) {
    if (moving) {
      const rect = node.parentNode.getBoundingClientRect();
      const left = rect.left;
      // do we set the size here?

      node.dispatchEvent(
        new CustomEvent("resize", {
          detail: {
            size: e.pageX - left,
          },
        })
      );
    }
  }

  function mouseup() {
    moving = false;
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
