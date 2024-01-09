/**
 * Note that this file has been vendored from Shoelace
 * in order to fix a bug with storybook.
 *
 * Linting is disabled for this file. If you make changes to this file,
 * you should merge it with the corresponding d.ts file, and change this
 * to a .ts file.
 */

export function getOffset(element, parent) {
  return {
    top: Math.round(
      element.getBoundingClientRect().top - parent.getBoundingClientRect().top,
    ),
    left: Math.round(
      element.getBoundingClientRect().left -
        parent.getBoundingClientRect().left,
    ),
  };
}
