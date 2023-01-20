export function isDuplicateName(
  name: string,
  fromName: string,
  names: Array<string>
) {
  if (name.toLowerCase() === fromName.toLowerCase()) return false;
  return names.findIndex((n) => n.toLowerCase() === name.toLowerCase()) >= 0;
}
