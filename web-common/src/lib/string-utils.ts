export function hasSpaces(str: string) {
  return /\s/.test(str);
}

export function slugify(str: string) {
  return str.toLowerCase().replace(/\s+/g, "-");
}
