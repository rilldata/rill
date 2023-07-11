/** FIXME: there is another Debounce class in this codebase, but
 * it appears to run the first callback, not the last.
 */
export function createDebouncer() {
  let timeout: ReturnType<typeof setTimeout>;
  const callback = (callback: () => void, time: number) => {
    if (timeout) {
      clearTimeout(timeout);
    }
    timeout = setTimeout(callback, time);
  };
  callback.clear = () => {
    clearTimeout(timeout);
  };
  return callback;
}
