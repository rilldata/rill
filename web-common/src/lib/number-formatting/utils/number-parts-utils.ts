import type { NumberParts } from "../humanizer-types";

/**
 * Applies thousand separators to an integer string based on locale configuration
 */
function applyThousandSeparators(
  intStr: string,
  thousands: string = ",",
  grouping: number[] = [3],
): string {
  if (!thousands || intStr.length <= grouping[0]) {
    return intStr;
  }

  const parts: string[] = [];
  let remaining = intStr;
  let groupIndex = 0;

  // Process from right to left
  while (remaining.length > 0) {
    const groupSize = grouping[Math.min(groupIndex, grouping.length - 1)];
    const start = Math.max(0, remaining.length - groupSize);
    const group = remaining.slice(start);
    parts.unshift(group);
    remaining = remaining.slice(0, start);
    groupIndex++;
  }

  return parts.join(thousands);
}

export const numberPartsToString = (parts: NumberParts): string => {
  const locale = parts.locale;
  const decimal = locale?.decimal || ".";
  const thousands = locale?.thousands || ",";
  const grouping = locale?.grouping || [3];

  // Apply thousand separators to the integer part if locale is specified
  const formattedInt =
    locale && thousands
      ? applyThousandSeparators(parts.int, thousands, grouping)
      : parts.int;

  // Use locale-specific decimal separator if available
  const dot = parts.dot && locale ? decimal : parts.dot;

  return (
    (parts.prefix || "") +
    (parts.neg || "") +
    (parts.currencySymbol || "") +
    formattedInt +
    dot +
    parts.frac +
    parts.suffix +
    (parts.percent || "")
  );
};

export function numStrToParts(numStr: string): NumberParts {
  const nonNumReMatch = numStr.match(/[a-zA-Z ]/);
  let int = "";
  const dot: "" | "." = numStr.includes(".") ? "." : "";
  let frac = "";
  let suffix = "";
  if (nonNumReMatch) {
    const suffixIndex = nonNumReMatch.index;
    const numPart = numStr.slice(0, suffixIndex);
    suffix = numStr.slice(suffixIndex);

    if (numPart.split(".").length == 1) {
      int = numPart;
    } else {
      int = numPart.split(".")[0];
      frac = numPart.split(".")[1] ?? "";
    }
  } else {
    int = numStr.split(".")[0];
    frac = numStr.split(".")[1] ?? "";
  }
  if (suffix === undefined || int === undefined || frac === undefined) {
    console.error({ numStr, int, frac, suffix });
  }
  return { int, dot, frac, suffix };
}
