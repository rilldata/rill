/** this action appends another text DOM element
 * that gives an outlined / punched-out look to whatever
 * svg text node it is applied to. It will then listen to
 * any of the relevant attributes / the content itself
 * and update accordingly via a basic MutationObserver.
 */

interface OutlineAction {
  destroy: () => void;
}

export function outline(
  node: SVGElement,
  args = { color: "white" }
): OutlineAction {
  const enclosingSVG = node.ownerSVGElement;

  // create a clone of the element.
  const clonedElement = node.cloneNode(true) as SVGElement;
  node.parentElement.insertBefore(clonedElement, node);
  clonedElement.setAttribute("fill", args.color);
  clonedElement.style.fill = args.color;
  clonedElement.setAttribute("filter", "url(#outline-filter)");
  // apply the filter to this svg element.
  let outlineFilter = enclosingSVG.querySelector("#outline-filter");
  if (outlineFilter === null) {
    outlineFilter = document.createElementNS(
      "http://www.w3.org/2000/svg",
      "filter"
    );
    outlineFilter.id = "outline-filter";

    const morph = document.createElementNS(
      "http://www.w3.org/2000/svg",
      "feMorphology"
    );
    morph.setAttribute("operator", "dilate");
    morph.setAttribute("radius", "2");
    morph.setAttribute("in", "SourceGraphic");
    morph.setAttribute("result", "THICKNESS");

    const composite = document.createElementNS(
      "http://www.w3.org/2000/svg",
      "feComposite"
    );
    composite.setAttribute("operator", "out");
    composite.setAttribute("in", "THICKNESS");
    composite.setAttribute("in2", "SourceGraphic");

    outlineFilter.appendChild(morph);
    outlineFilter.appendChild(composite);
    enclosingSVG.prepend(outlineFilter);
  }

  const config = {
    attributes: true,
    childList: true,
    subtree: true,
    characterData: true,
  };
  const observer = new MutationObserver(() => {
    clonedElement.setAttribute("x", node.getAttribute("x"));
    clonedElement.setAttribute("y", node.getAttribute("y"));
    if (node.getAttribute("text-anchor")) {
      clonedElement.setAttribute(
        "text-anchor",
        node.getAttribute("text-anchor")
      );
    }

    if (node.getAttribute("dx")) {
      clonedElement.setAttribute("dx", node.getAttribute("dx"));
    }
    if (node.getAttribute("dy")) {
      clonedElement.setAttribute("dy", node.getAttribute("dy"));
    }

    // clone any animations that may be applied via svelte transitions.
    clonedElement.style.animation = node.style.animation;
    // copy the contents of the node.
    clonedElement.innerHTML = node.innerHTML;
  });
  observer.observe(node, config);

  return {
    destroy() {
      clonedElement.remove();
    },
  };
}
