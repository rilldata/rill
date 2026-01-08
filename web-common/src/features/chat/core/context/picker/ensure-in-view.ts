export function ensureInView(node: HTMLElement, focused: boolean) {
  if (focused) {
    node.scrollIntoView({ block: "nearest" });
  }
  return {
    update(focused: boolean) {
      if (focused) {
        node.scrollIntoView({ block: "nearest" });
      }
    },
  };
}
