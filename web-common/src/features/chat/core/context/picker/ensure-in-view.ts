export function ensureInView(node: HTMLElement, active: boolean) {
  if (active) {
    node.scrollIntoView({ block: "nearest" });
  }
  return {
    update(active: boolean) {
      if (active) {
        node.scrollIntoView({ block: "nearest" });
      }
    },
  };
}
