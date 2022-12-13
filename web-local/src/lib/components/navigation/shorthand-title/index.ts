const filterStrings =
  (...what: string[]) =>
  (str: string) =>
    !what.some((w) => w === str);

const replacePunctuationWithSpace = (str: string) => {
  return str.replace(/[^a-zA-Z0-9 ]/g, " ");
};

export function shorthandTitle(str: string) {
  if (!str) return;
  const out = replacePunctuationWithSpace(str)
    .toUpperCase()
    .split(" ")
    .filter(filterStrings("AND", "OR", "THE"))
    .filter((word: string) => word !== "") as string[];

  if (out.length === 1) {
    // take first two letters
    const first = out[0].slice(0, 2);
    if (first.length === 2) {
      const chars = first.split("");
      chars[1] = chars[1].toLowerCase();
      return chars.join("");
    }
  }
  // take first letter of first two words
  else
    return out
      .map((word) => word?.[0])
      .join("")
      .slice(0, 2);
}
