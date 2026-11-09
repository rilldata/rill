import { TimeBoundedBuffer } from "@rilldata/web-common/lib/time-bounded-buffer";

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

// The loop notice arrives in the same frame as the callbacks that caused it
const RECORD_WINDOW_MS = 100;
// A hard limit for frames in which many observers fire, such as a canvas with many charts
const MAX_RECORDS = 50;
const MAX_TARGETS_PER_CALLBACK = 3;
const MAX_CLASSES = 3;
const MAX_DESCRIPTION_CHARS = 80;

// Weak references to the resized elements, described only when a notice is reported, so the tracker holds no DOM alive
const recent = new TimeBoundedBuffer<Array<WeakRef<Element>>>(
  RECORD_WINDOW_MS,
  MAX_RECORDS,
);
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
        recent.push(
          entries
            .slice(0, MAX_TARGETS_PER_CALLBACK)
            .map((entry) => new WeakRef(entry.target)),
        );
        callback.call(observer, entries, observer);
      });
    }
  };
}

/** The callbacks that ran within the last 100 milliseconds, most recent first. */
export function getRecentResizeObserverCallbacks(): ResizeCallbackRecord[] {
  return recent.recent().map(({ at, item }) => ({
    at,
    targets: item
      .map((target) => target.deref())
      .filter((element) => element !== undefined)
      .map(describeElement),
  }));
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
