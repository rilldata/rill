export function extractNameFromSlug(
  slug: string
): [name: string, version: string, state: string] {
  const [name, version, state] = slug.split("--");
  return [name, version, state];
}
