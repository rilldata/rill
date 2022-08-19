export function removeIfExists<T>(array: Array<T>, checker: (e: T) => boolean) {
  const index = array.findIndex(checker);
  if (index >= 0) {
    array.splice(index, 1);
  }
}
