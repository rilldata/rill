export function isAnythingSelected(filters): boolean {
  if (!filters) return false;
  return Object.keys(filters).some((key) => {
    return filters[key]?.length;
  });
}
