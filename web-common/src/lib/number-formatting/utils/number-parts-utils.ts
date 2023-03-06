import type {
  FormatterFactory,
  NumberFormatter,
  NumberParts,
  NumberKind,
  FormatterWidths,
  PxWidthLookupFn,
} from "../humanizer-types";

export const numberPartsToString = (parts: NumberParts): string =>
  (parts.neg || "") +
  (parts.dollar || "") +
  parts.int +
  parts.dot +
  parts.frac +
  parts.suffix +
  (parts.percent || "");
