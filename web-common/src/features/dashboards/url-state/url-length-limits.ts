const HEADER_SIZE_LIMIT = 16 * 1024; // 16KB
const URL_LENGTH_THRESHOLD = HEADER_SIZE_LIMIT * 0.9; // -10% of the actual limit. This is arbitrary

export function isUrlTooLong(url: URL) {
  const length = new TextEncoder().encode(url.toString()).length;
  return length > URL_LENGTH_THRESHOLD;
}
