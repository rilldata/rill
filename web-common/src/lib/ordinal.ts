function indicator(num: number): string {
  num = Math.abs(num);
  const cent = num % 100;
  if (cent >= 10 && cent <= 20) return "th";
  const dec = num % 10;
  if (dec === 1) return "st";
  if (dec === 2) return "nd";
  if (dec === 3) return "rd";
  return "th";
}

export function ordinal(num: number): string {
  return `${num}${indicator(num)}`;
}
