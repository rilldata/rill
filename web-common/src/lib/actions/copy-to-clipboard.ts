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

function copyToClipboardAPI(value: string) {
  navigator.clipboard.writeText(value).catch(console.error);
}

export function copyToClipboard(
  value: string,
  message?: string,
  notify: boolean = true,
) {
  if (isClipboardApiSupported()) {
    copyToClipboardAPI(value);
    if (notify) {
      eventBus.emit("notification", {
        message: message || `Copied ${value} to clipboard`,
      });
    }
  } else {
    console.warn("Clipboard API not supported.");
  }
}
