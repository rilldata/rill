import { getOffset } from "./offset.js";
const locks = new Set();
function getScrollbarWidth() {
  const documentWidth = document.documentElement.clientWidth;
  return Math.abs(window.innerWidth - documentWidth);
}
export function lockBodyScrolling(lockingEl) {
  locks.add(lockingEl);
  if (!document.body.classList.contains("sl-scroll-lock")) {
    const scrollbarWidth = getScrollbarWidth();
    document.body.classList.add("sl-scroll-lock");
    document.body.style.setProperty(
      "--sl-scroll-lock-size",
      `${scrollbarWidth}px`,
    );
  }
}
export function unlockBodyScrolling(lockingEl) {
  locks.delete(lockingEl);
  if (locks.size === 0) {
    document.body.classList.remove("sl-scroll-lock");
    document.body.style.removeProperty("--sl-scroll-lock-size");
  }
}
export function scrollIntoView(
  element,
  container,
  direction = "vertical",
  behavior = "smooth",
) {
  const offset = getOffset(element, container);
  const offsetTop = offset.top + container.scrollTop;
  const offsetLeft = offset.left + container.scrollLeft;
  const minX = container.scrollLeft;
  const maxX = container.scrollLeft + container.offsetWidth;
  const minY = container.scrollTop;
  const maxY = container.scrollTop + container.offsetHeight;
  if (direction === "horizontal" || direction === "both") {
    if (offsetLeft < minX) {
      container.scrollTo({ left: offsetLeft, behavior });
    } else if (offsetLeft + element.clientWidth > maxX) {
      container.scrollTo({
        left: offsetLeft - container.offsetWidth + element.clientWidth,
        behavior,
      });
    }
  }
  if (direction === "vertical" || direction === "both") {
    if (offsetTop < minY) {
      container.scrollTo({ top: offsetTop, behavior });
    } else if (offsetTop + element.clientHeight > maxY) {
      container.scrollTo({
        top: offsetTop - container.offsetHeight + element.clientHeight,
        behavior,
      });
    }
  }
}
