const convertToPixelNumber = (px: string) => {
  if (!px.length) return 0;
  return +px.slice(0, -2);
};

export function truncateMiddleText(node: Element) {
  /** cache originalText. NOTE: if the text of the node updates, we'll have to  */
  const originalText = node.innerHTML;
  const parent = node.parentElement;
  /** set ARIA label to original text if not specified by node. */
  if (node.ariaLabel === null) node.ariaLabel = originalText;

  /** compares the scrollWidth of the node against hte parent's width.
   * If the node is too wide, truncate the text from the middle and try again.
   */
  function truncate() {
    const leftPad = convertToPixelNumber(
      getComputedStyle(parent).getPropertyValue("padding-left")
    );
    const rightPad = convertToPixelNumber(
      getComputedStyle(parent).getPropertyValue("padding-right")
    );
    const parentWidth = parent.offsetWidth - leftPad - rightPad;

    for (let i = 0; i < originalText.length; i++) {
      if (node.scrollWidth > parentWidth) {
        const leftSide = originalText.slice(0, originalText.length / 2 - i);
        const rightSide = originalText.slice(originalText.length / 2 + i);
        node.innerHTML = `${leftSide}...${rightSide}`;
      } else {
        /** The node is smaller than the available space. Show the entire string. */
        node.innerHTML = originalText;
      }

      /** end this loop. */
      if (node.scrollWidth <= parentWidth) {
        break;
      }
    }
  }

  /** perform initial truncation */
  truncate();

  /** Observe the parent element for changes in size, and truncate accordingly. */
  const r = new ResizeObserver(truncate);
  r.observe(node.parentElement);

  return {
    destroy() {
      r.unobserve(node.parentElement);
    },
  };
}
