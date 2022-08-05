export function clickOutside(node: Node, [elementsToIgnore, cb]) {
  function handleClick(event) {
    if (
      !node.contains(document.activeElement) &&
      !node.contains(event.target) &&
      node !== event.target &&
      (!elementsToIgnore.every((element) => Boolean(element)) ||
        elementsToIgnore.every((element) => !element.contains(event.target)))
    ) {
      cb(event);
    }
  }
  document.addEventListener("click", handleClick);
  document.addEventListener("focusin", handleClick);
  return {
    destroy() {
      document.removeEventListener("click", handleClick);
      document.removeEventListener("focusin", handleClick);
    },
  };
}
