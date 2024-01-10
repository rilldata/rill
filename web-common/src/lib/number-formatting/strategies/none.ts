import { shortScaleSuffixIfAvailableForStr } from "../utils/short-scale-suffixes";
import {
  NumberParts,
  Formatter,
  NumberKind,
  FormatterOptionsNoneStrategy,
} from "../humanizer-types";
import { numStrToParts } from "../utils/number-parts-utils";

export class NonFormatter implements Formatter {
  options: FormatterOptionsNoneStrategy;
  initialSample: number[];

  constructor(sample: number[], options: FormatterOptionsNoneStrategy) {
    this.options = options;
    this.initialSample = sample;
  }

  stringFormat(x: number): string {
    return x.toString();
  }

  partsFormat(x: number): NumberParts {
    let numParts;

    const isPercent = this.options.numberKind === NumberKind.PERCENT;

    if (isPercent) x = 100 * x;

    if (x === 0) {
      numParts = { int: "0", dot: "", frac: "", suffix: "" };
    } else {
      const str = new Intl.NumberFormat("en", {
        maximumFractionDigits: 20,
        useGrouping: false,
      }).format(x);
      numParts = numStrToParts(str);
    }

    numParts.suffix = shortScaleSuffixIfAvailableForStr(numParts.suffix);

    if (this.options.upperCaseEForExponent !== true) {
      numParts.suffix = numParts.suffix.replace("E", "e");
    }

    if (this.options.numberKind === NumberKind.DOLLAR) {
      numParts.currencySymbol = "$";
    } else if (this.options.numberKind === NumberKind.EURO) {
      numParts.currencySymbol = "€";
    }

    if (isPercent) {
      numParts.percent = "%";
    }

    return numParts;
  }
}
