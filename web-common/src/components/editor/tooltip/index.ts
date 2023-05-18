import { syntaxTree } from "@codemirror/language";
import { hoverTooltip } from "@codemirror/view";
export const tooltipPlugin = hoverTooltip((view, pos) => {
  // const path = [];
  // const line = view.lineAt(e.clientY);
  // const pos = view.posAtCoords({ x: e.clientX, y: e.clientY });
  const { from, to } = view.state.doc.lineAt(pos);
  // const parent = null;
  syntaxTree(view.state).iterate({
    from,
    to,
    enter(node) {
      console.log(
        node.name,
        node.name === "atom",
        node,
        node.from,
        node.to,
        view.state.doc.sliceString(node.from, node.to).slice(0, 30),
        node.node.parent
      );
    },
  });
  return {
    pos: from - 3,
    end: to + 3,
    above: true,
    create(v) {
      const dom = document.createElement("div");
      dom.textContent = "sup!!!";
      return { dom };
    },
  };
});

export function createTooltipPlugin() {
  return [tooltipPlugin];
}
