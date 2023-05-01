/**
 * Get the total size of an element, including margins
 * @param element
 */
export function getEltSize(element: HTMLElement, direction: "x" | "y") {
  if (!["x", "y"].includes(direction)) {
    throw new Error("direction must be 'x' or 'y'");
  }
  if (!element) return 0;
  // Get the computed style of the element
  const style = window.getComputedStyle(element);
  if (direction === "y") {
    // Get the element's height (including padding and border)
    const height = element.getBoundingClientRect().height;
    // Get the margin values
    const marginTop = parseFloat(style.marginTop);
    const marginBottom = parseFloat(style.marginBottom);
    // Calculate the total height including margin
    return height + marginTop + marginBottom;
  } else {
    const width = element.getBoundingClientRect().width;
    const marginLeft = parseFloat(style.marginLeft);
    const marginRight = parseFloat(style.marginRight);
    return width + marginLeft + marginRight;
  }
}
