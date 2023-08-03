/** An action that creates a hidden measuring DOM element, which
 * can be used to simulate the style `width: max-content` in an input element.
 */
export function dynamicTextInputWidth(node: HTMLInputElement) {
  const measuringElement = document.createElement("div");
  document.body.appendChild(measuringElement);

  /** check if teh DOM element has box-sizing: border-box. If not, enforce it. */
  // const computedStyles = window.getComputedStyle(node);
  // if (computedStyles.boxSizing !== "border-box") {
  //   node.style.boxSizing = "border-box";
  // }

  /** duplicate the styles of the existing node, but
  remove the measuring element from the viewport. */
  function duplicateAndSet() {
    const styles = window.getComputedStyle(node);
    measuringElement.innerHTML = node.value;
    measuringElement.style.fontSize = styles.fontSize;
    measuringElement.style.fontWeight = styles.fontWeight;
    measuringElement.style.fontFamily = styles.fontFamily;
    measuringElement.style.paddingLeft = styles.paddingLeft;
    measuringElement.style.paddingRight = styles.paddingRight;
    measuringElement.style.border = styles.border;
    measuringElement.style.boxSizing = "border-box";
    measuringElement.style.width = "max-content";
    measuringElement.style.position = "absolute";
    measuringElement.style.top = "0";
    measuringElement.style.left = "-9999px";
    measuringElement.style.overflow = "hidden";
    measuringElement.style.visibility = "hidden";
    measuringElement.style.whiteSpace = "pre";
    measuringElement.style.height = "0";
    node.style.width = `${measuringElement.offsetWidth}px`;
  }

  /** update the measuringElement */
  duplicateAndSet();

  /** listen to any style changes */
  const observer = new MutationObserver(duplicateAndSet);
  observer.observe(node, { attributes: true, childList: true, subtree: true });

  node.addEventListener("input", duplicateAndSet);
  return {
    destroy() {
      /** remove mutation observer */
      observer.disconnect();
      node.removeEventListener("input", duplicateAndSet);
      document.body.removeChild(measuringElement);
    },
  };
}
