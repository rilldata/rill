import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

const IS_SAFARI = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

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

  if (typeof ClipboardItem === "undefined") {
    console.warn("ClipboardItem is not supported in this environment.");
    return false;
  }

  return true;
}

async function copyToClipboardAPI(value: string) {
  try {
    if (IS_SAFARI) {
      const text = new ClipboardItem({
        "text/plain": new Blob([value], { type: "text/plain" }),
      });
      console.log("a");
      await navigator.clipboard.write([text]);
    } else {
      console.log("b");
      await navigator.clipboard.writeText(value);
    }
  } catch (error) {
    console.error("Failed to copy to clipboard using Clipboard API:", error);
  }
}

export async function copyToClipboard(value: string, message?: string) {
  if (isClipboardApiSupported()) {
    await copyToClipboardAPI(value);
    eventBus.emit("notification", {
      message: message || `Copied ${value} to clipboard`,
    });
    return true;
  } else {
    console.warn("Clipboard API not supported.");
    return false;
  }
}

// REVISIT
// // See: https://wolfgangrittner.dev/how-to-use-clipboard-api-in-safari/
// // Related: https://developer.apple.com/forums/thread/691873

// // SAFARI ERROR:
// // Failed to copy to clipboard using Clipboard API:
// // NotAllowedError: The request is not allowed by the user agent or the platform in the current context, possibly because the user denied permission.
// function copyToClipboardSafari(value: string) {
//   const text = new ClipboardItem({
//     "text/plain": fetch(value)
//       .then((response) => response.text())
//       .then((text) => new Blob([text], { type: "text/plain" })),
//   });
//   navigator.clipboard.write([text]);
// }

// // See: https://wolfgangrittner.dev/how-to-use-clipboard-api-in-firefox/
// function copyToClipboardFirefox(value: string) {
//   fetch(value)
//     .then((response) => response.text())
//     .then((text) => navigator.clipboard.writeText(text))
//     .catch(console.error);
// }
