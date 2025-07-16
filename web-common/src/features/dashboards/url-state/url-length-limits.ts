const HEADER_SIZE_LIMIT = 16 * 1024; // 16KB
const URL_LENGTH_THRESHOLD = HEADER_SIZE_LIMIT * 0.8; // -20% of the actual limit. This is arbitrary to include cookies and other metadata filled in by backend for tokens.

export function isUrlTooLong(url: URL) {
  // Use TextEncoder to estimate the size of the url in bytes.
  // `encode` here returns bytes array, so its length is what we want.
  const length = new TextEncoder().encode(url.toString()).length;
  return length > URL_LENGTH_THRESHOLD;
}
