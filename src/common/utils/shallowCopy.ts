export function shallowCopy(
  source: Record<string, any>,
  target: Record<string, any>
): void {
  Object.keys(source).forEach((k) => {
    target[k] = source[k];
  });
}
