export function snakeToCamel(self: string): string {
  let str = self[0];
  for (let i = 1; i < self.length; i++) {
    const isUnderscore = self[i] === "_";
    if (!isUnderscore) {
      str += self[i];
      continue;
    }

    i++;
    if (i >= self.length) break;

    str += self[i].toUpperCase();
  }
  return str;
}
