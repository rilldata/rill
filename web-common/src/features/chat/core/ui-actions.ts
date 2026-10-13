export const AI_ACTION_ATTRIBUTE = "data-ai-action";
export const AI_ACTION_LABEL_ATTRIBUTE = "data-ai-label";

export interface AvailableUIAction {
  id: string;
  label: string;
}

function isAvailable(element: HTMLElement): boolean {
  if (
    element.hidden ||
    element.closest("[data-ai-ignore]") ||
    element.closest('[aria-hidden="true"]') ||
    element.matches(":disabled") ||
    element.getAttribute("aria-disabled") === "true"
  ) {
    return false;
  }

  const style = window.getComputedStyle(element);
  if (style.display === "none" || style.visibility === "hidden") return false;

  const rect = element.getBoundingClientRect();
  return rect.width > 0 && rect.height > 0;
}

function actionLabel(element: HTMLElement, id: string): string {
  const label =
    element.getAttribute(AI_ACTION_LABEL_ATTRIBUTE) ||
    element.getAttribute("aria-label") ||
    element.textContent ||
    id;
  return label.replace(/\s+/g, " ").trim().slice(0, 256) || id;
}

function availableActionElements(root: ParentNode = document): HTMLElement[] {
  return Array.from(
    root.querySelectorAll<HTMLElement>(`[${AI_ACTION_ATTRIBUTE}]`),
  ).filter(isAvailable);
}

/**
 * Returns visible and enabled actions that have a unique ID on the current page.
 * Duplicate IDs are omitted so the agent can never address an ambiguous target.
 */
export function getAvailableUIActions(
  root: ParentNode = document,
): AvailableUIAction[] {
  const elements = availableActionElements(root);
  const counts = new Map<string, number>();
  for (const element of elements) {
    const id = element.getAttribute(AI_ACTION_ATTRIBUTE)?.trim();
    if (id) counts.set(id, (counts.get(id) ?? 0) + 1);
  }

  return elements.flatMap((element) => {
    const id = element.getAttribute(AI_ACTION_ATTRIBUTE)?.trim();
    if (!id || counts.get(id) !== 1) return [];
    return [{ id, label: actionLabel(element, id) }];
  });
}

const actionHandlers = new Map<string, () => Promise<void>>();

/**
 * Registers a handler that runs in place of a click when the AI triggers the action.
 * Use it for actions that finish asynchronously, so the AI waits for the outcome:
 * a rejected promise is reported back to the AI as the action's error.
 * The element carrying the action ID still gates availability. Returns an unregister function.
 */
export function registerUIActionHandler(
  actionId: string,
  handler: () => Promise<void>,
): () => void {
  actionHandlers.set(actionId, handler);
  return () => {
    if (actionHandlers.get(actionId) === handler)
      actionHandlers.delete(actionId);
  };
}

/**
 * Runs a visible, enabled, uniquely-addressable action: its registered handler if it has one,
 * otherwise a click. Throws if the page changed since discovery or the action fails.
 */
export async function runUIAction(
  actionId: string,
  root: ParentNode = document,
): Promise<void> {
  const matches = availableActionElements(root).filter(
    (element) => element.getAttribute(AI_ACTION_ATTRIBUTE)?.trim() === actionId,
  );
  if (matches.length !== 1) {
    throw new Error(
      `Action is no longer uniquely available on the page: ${actionId}`,
    );
  }

  const handler = actionHandlers.get(actionId);
  if (handler) {
    await handler();
  } else {
    matches[0].click();
  }
}
