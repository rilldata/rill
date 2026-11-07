/**
 * Wraps the global ResizeObserver so that error telemetry can say which observer callbacks ran in the frame before
 * the browser reported "ResizeObserver loop completed with undelivered notifications."; the notice itself has no stack.
 * Install it before any component mounts: observers created earlier are not wrapped.
 */

export interface ResizeCallbackRecord {
  // performance.now() when the callback ran
  at: number;
  // Short descriptions of the resized elements, see describeElement
  targets: string[];
}

const MAX_RECORDS = 8;
const MAX_TARGETS_PER_CALLBACK = 3;
const MAX_CLASSES = 3;
const MAX_DESCRIPTION_CHARS = 80;

// Descriptions only, never the elements themselves, so the tracker holds no DOM alive
const recent: ResizeCallbackRecord[] = [];
let installed = false;

export function installResizeObserverTracker(): void {
  if (
    installed ||
    typeof window === "undefined" ||
    typeof window.ResizeObserver !== "function"
  ) {
    return;
  }
  installed = true;

  const NativeResizeObserver = window.ResizeObserver;
  window.ResizeObserver = class extends NativeResizeObserver {
    constructor(callback: ResizeObserverCallback) {
      super((entries, observer) => {
        recent.push({
          at: performance.now(),
          targets: entries
            .slice(0, MAX_TARGETS_PER_CALLBACK)
            .map((entry) => describeElement(entry.target)),
        });
        if (recent.length > MAX_RECORDS) recent.shift();
        callback(entries, observer);
      });
    }
  };
}

/** The callbacks that ran within the last `withinMs` milliseconds, most recent first. */
export function getRecentResizeObserverCallbacks(
  withinMs = 100,
): ResizeCallbackRecord[] {
  const now = performance.now();
  return recent.filter((record) => now - record.at <= withinMs).reverse();
}

/** `tag#id.class.class`, plus the closest ancestor with an id when the element has none, kept short. */
export function describeElement(element: Element): string {
  let description = element.tagName.toLowerCase();
  if (element.id) description += `#${element.id}`;
  const classes = Array.from(element.classList).slice(0, MAX_CLASSES);
  if (classes.length) description += `.${classes.join(".")}`;
  if (!element.id) {
    const ancestor = element.parentElement?.closest("[id]");
    if (ancestor) description += ` in #${ancestor.id}`;
  }
  return description.slice(0, MAX_DESCRIPTION_CHARS);
}
