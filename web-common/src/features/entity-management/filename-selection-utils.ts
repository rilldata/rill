/**
 * Calculates the selection range for a filename, excluding the extension.
 * This allows users to edit just the filename without having to adjust the selection.
 *
 * @param fileName - The full filename including extension
 * @returns An object with selectionStart and selectionEnd, or undefined if no selection should be made
 */
export function getFilenameSelectionRange(
  fileName: string | undefined,
): { selectionStart: number; selectionEnd: number } | undefined {
  if (!fileName) return undefined;

  const lastDotIndex = fileName.lastIndexOf(".");
  
  // Only create a selection if there's an extension (dot not at the start)
  if (lastDotIndex > 0) {
    return {
      selectionStart: 0,
      selectionEnd: lastDotIndex,
    };
  }

  return undefined;
}

