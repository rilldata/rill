export function filterItemsSortFunction<T extends { name: string }>(
  a: T,
  b: T
) {
  return a.name > b.name ? 1 : -1;
}
