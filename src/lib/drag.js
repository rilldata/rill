import { layout } from "$lib/application-state-stores/layout-store";
export function drag(node, params) {
  let minSize_ = params?.minSize || 300;
  let maxSize_ = params?.maxSize || 800;
  let reverse_ = params?.reverse || false;
  let orientation_ = params?.orientation || "horizontal";

  let side_ = params?.side || "right";
  let property = `--${side_}-sidebar-width`; //params?.property || '--left-sidebar-width';
  let moving = false;
  let space = minSize_;

  node.style.cursor = "move";
  node.style.userSelect = "none";

  function mousedown() {
    moving = true;
  }

  function mousemove(e) {
    if (moving) {
      let size;
      if (orientation_ === "horizontal") {
        size = reverse_ ? innerWidth - e.pageX : e.pageX;
      } else if (orientation_ === "vertical") {
        size = reverse_ ? innerHeight - e.pageY : e.pageY;
      }
      if (size > minSize_ && size < maxSize_) {
        space = size;
      }
      layout.update((l) => {
        l[side_] = space;
        return l;
      });
      //document.body.style.setProperty(property, `${xSpace}px`)
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
  };
}
