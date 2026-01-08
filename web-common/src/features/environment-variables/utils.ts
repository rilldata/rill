/**
 * Check if a key is a duplicate (case-insensitive)
 */
export function isDuplicateKey(
  key: string,
  existingKeys: string[],
  currentKey?: string,
): boolean {
  const normalizedKey = key.toLowerCase();
  const normalizedCurrentKey = currentKey?.toLowerCase();

  return existingKeys.some((existingKey) => {
    const normalizedExistingKey = existingKey.toLowerCase();

    // Skip if this is the current key being edited
    if (
      normalizedCurrentKey &&
      normalizedExistingKey === normalizedCurrentKey
    ) {
      return false;
    }

    return normalizedExistingKey === normalizedKey;
  });
}
