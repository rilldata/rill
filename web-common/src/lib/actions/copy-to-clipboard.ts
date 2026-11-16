import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

/**
 * The Clipboard API is only available in secure contexts.
 * So, a self-hosted Rill Developer instance served over HTTP (not HTTPS) will not have access to the Clipboard API.
 * See: https://developer.mozilla.org/en-US/docs/Web/API/Clipboard
 */
export function isClipboardApiSupported(): boolean {
  if (!navigator.clipboard) {
    console.warn(
      "Clipboard API is not supported in this environment. Ensure HTTPS or localhost.",
    );
    return false;
  }

  return true;
}

export function copyToClipboard(
  value: string,
  message?: string,
  notify: boolean = true,
): void {
  void tryCopyToClipboard(value).then((copied) => {
    if (copied && notify) {
      eventBus.emit("notification", {
        message: message || `Copied ${value} to clipboard`,
      });
    }
  });
}

export async function tryCopyToClipboard(value: string): Promise<boolean> {
  if (!isClipboardApiSupported()) return false;

  try {
    await navigator.clipboard.writeText(value);
  } catch (error) {
    console.error(error);
    return false;
  }

  return true;
}
