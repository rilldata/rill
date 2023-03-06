import type {
  FormatterFactory,
  NumberFormatter,
  NumberParts,
  NumberKind,
  FormatterWidths,
  PxWidthLookupFn,
} from "../humanizer-types";

function maxWidthsForSplitStrs(
  numPartsArr: NumberParts[],
  widthLookupFn: PxWidthLookupFn
): FormatterWidths {
  let widths: FormatterWidths = { left: 0, dot: 0, frac: 0, suffix: 0 };
  numPartsArr.forEach((ss) => {
    widths.left = Math.max(
      widths.left,
      widthLookupFn(ss.neg) + widthLookupFn(ss.dollar) + widthLookupFn(ss.int)
    );

    widths.dot = Math.max(widths.dot, widthLookupFn(ss.dot));

    widths.frac = Math.max(widths.left, widthLookupFn(ss.frac));

    widths.suffix = Math.max(
      widths.left,
      widthLookupFn(ss.suffix) + widthLookupFn(ss.percent)
    );
  });

  return widths;
}

export const maxPxWidthsForSplitStrs = (
  numPartsArr: NumberParts[],
  pxWidthLookupFn: PxWidthLookupFn
): FormatterWidths => maxWidthsForSplitStrs(numPartsArr, pxWidthLookupFn);

export const maxCharWidthsForSplitStrs = (numPartsArr: NumberParts[]) =>
  maxWidthsForSplitStrs(
    numPartsArr,
    (str: string | undefined): number => str?.length || 0
  );

// export function maxCharWidthsForSplitStrs(
//   numPartsArr: NumberStringParts[]
// ): FormatterWidths {
//   let widths: FormatterWidths = { left: 0, dot: 0, frac: 0, suffix: 0 };
//   numPartsArr.forEach((ss) => {
//     widths.left = Math.max(
//       widths.left,
//       ss.neg.length + (ss?.dollar.length || 0) + ss.int.length
//     );

//     widths.dot = Math.max(widths.dot, ss.dot.length);

//     widths.frac = Math.max(widths.left, ss.frac.length);

//     widths.suffix = Math.max(
//       widths.left,
//       ss.suffix.length + (ss?.percent.length || 0)
//     );
//   });

//   return widths;
// }

// export function maxPxWidthsForSplitStrs(
//   numPartsArr: NumberStringParts[],
//   pxWidthLookupFn: PxWidthLookupFn
// ): FormatterWidths {
//   let widths: FormatterWidths = { left: 0, dot: 0, frac: 0, suffix: 0 };
//   numPartsArr.forEach((ss) => {
//     widths.left = Math.max(
//       widths.left,
//       pxWidthLookupFn(ss.neg) +
//         pxWidthLookupFn(ss.dollar) +
//         pxWidthLookupFn(ss.int)
//     );

//     widths.dot = Math.max(widths.dot, pxWidthLookupFn(ss.dot));

//     widths.frac = Math.max(widths.left, pxWidthLookupFn(ss.frac));

//     widths.suffix = Math.max(
//       widths.left,
//       pxWidthLookupFn(ss.suffix) + pxWidthLookupFn(ss.percent)
//     );
//   });

//   return widths;
// }
