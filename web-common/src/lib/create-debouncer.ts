export function createDebouncer() {
  let timeout: number;
  const callback = <F extends (...args: Parameters<F>) => ReturnType<F>>(
    callback: F,
    time: number,
  ) => {
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

export const debounce = <F extends (...args: Parameters<F>) => ReturnType<F>>(
  fn: F,
  delay: number,
) => {
  let timeout: ReturnType<typeof setTimeout>;
  return function (...args: Parameters<F>) {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  };
};
