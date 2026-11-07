type Debounced<F extends (...args: Parameters<F>) => ReturnType<F>> = ((
  ...args: Parameters<F>
) => void) & { cancel: () => void };

export const debounce = <F extends (...args: Parameters<F>) => ReturnType<F>>(
  fn: F,
  delay: number,
): Debounced<F> => {
  let timeout: ReturnType<typeof setTimeout>;
  const debounced = function (...args: Parameters<F>) {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  } as Debounced<F>;
  debounced.cancel = () => clearTimeout(timeout);
  return debounced;
};
