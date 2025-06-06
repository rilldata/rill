const HEADER_SIZE_LIMIT = 16 * 1024; // 16KB
const URL_LENGTH_THRESHOLD = HEADER_SIZE_LIMIT * 0.9; // -10% of the actual limit. This is arbitrary
const HEADER_WORD_SIZE = 4; // Word size for headers is 4

export function isUrlTooLong(url: URL) {
  return url.toString().length * HEADER_WORD_SIZE > URL_LENGTH_THRESHOLD;
}
