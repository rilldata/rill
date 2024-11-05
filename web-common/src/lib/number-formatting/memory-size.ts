/**
 * Temporary adhoc util to format memory size
 * TODO: fit this into our formatter and support it in measure formatting
 */

const Suffixes = ["", ..."KMGTP".split("")];
const Degree = 1024;

export function formatMemorySize(size: number): string {
  if (!size) return "0";
  const positive = size >= 0;
  size = Math.abs(size);
  let i = 0;
  for (; i < Suffixes.length - 1; i++) {
    if (Degree > size) break;
    size = size / Degree;
  }
  return `${positive ? "" : "-"}${size.toFixed(2)}${Suffixes[i]}B`;
}
