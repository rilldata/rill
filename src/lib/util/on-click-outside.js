// https://stackoverflow.com/a/3028037/4297741
export function onClickOutside(callback, ...elements) {
  const outsideClickListener = (event) => {
    if (
      !elements.every((element) => Boolean(element)) ||
      elements.every((element) => !element.contains(event.target))
    ) {
      callback();
      /* eslint-disable no-use-before-define */
      removeClickListener();
      /* eslint-enable no-use-before-define */
    }
  };
  const removeClickListener = () => {
    document.removeEventListener("click", outsideClickListener);
  };
  setTimeout(() => {
    document.addEventListener("click", outsideClickListener);
  });
  return removeClickListener;
}
