import { PerRangeFormatter } from "./per-range";
import { defaultPercentOptions } from "./per-range-default-options";

const defaultGenericNumTestCases: [number, string][] = [
  // integers
  [999_999_999 / 100, "1.0B%"],
  [12_345_789 / 100, "12.3M%"],
  [2_345_789 / 100, "2.3M%"],
  [999_999 / 100, "1.0M%"],
  [345_789 / 100, "345.8k%"],
  [45_789 / 100, "45.8k%"],
  [5_789 / 100, "5.8k%"],
  [999 / 100, "999%"],
  [789 / 100, "789%"],
  [89 / 100, "89%"],
  [9 / 100, "9%"],
  [0 / 100, "0%"],
  [-0 / 100, "0%"],
  [-999_999_999 / 100, "-1.0B%"],
  [-12_345_789 / 100, "-12.3M%"],
  [-2_345_789 / 100, "-2.3M%"],
  [-999_999 / 100, "-1.0M%"],
  [-345_789 / 100, "-345.8k%"],
  [-45_789 / 100, "-45.8k%"],
  [-5_789 / 100, "-5.8k%"],
  [-999 / 100, "-999%"],
  [-789 / 100, "-789%"],
  [-89 / 100, "-89%"],
  [-9 / 100, "-9%"],

  // non integers
  [999_999_999.1234686 / 100, "1.0B%"],
  [12_345_789.1234686 / 100, "12.3M%"],
  [2_345_789.1234686 / 100, "2.3M%"],
  [999_999.4397 / 100, "1.0M%"],
  [345_789.1234686 / 100, "345.8k%"],
  [45_789.1234686 / 100, "45.8k%"],
  [5_789.1234686 / 100, "5.8k%"],
  [999.999 / 100, "1.0k%"],
  [999.995 / 100, "1.0k%"],
  [999.994 / 100, "999.99%"],
  [999.99 / 100, "999.99%"],
  [999.1234686 / 100, "999.12%"],
  [789.1234686 / 100, "789.12%"],
  [89.1234686 / 100, "89.12%"],
  [9.1234686 / 100, "9.12%"],
  [0.1234686 / 100, "0.12%"],

  [-999_999_999.1234686 / 100, "-1.0B%"],
  [-12_345_789.1234686 / 100, "-12.3M%"],
  [-2_345_789.1234686 / 100, "-2.3M%"],
  [-999_999.4397 / 100, "-1.0M%"],
  [-345_789.1234686 / 100, "-345.8k%"],
  [-45_789.1234686 / 100, "-45.8k%"],
  [-5_789.1234686 / 100, "-5.8k%"],
  [-999.999 / 100, "-1.0k%"],
  [-999.1234686 / 100, "-999.12%"],
  [-789.1234686 / 100, "-789.12%"],
  [-89.1234686 / 100, "-89.12%"],
  [-9.1234686 / 100, "-9.12%"],
  [-0.1234686 / 100, "-0.12%"],

  // infinitesimals + making sure there is no padding with insignificant zeros
  [0.008, "0.8%"],
  [0.005, "0.5%"],
  /** NOTE CORNER CASE TO IGNORE
   * ideally, 0.009 would format as "0.9%" (no sero padding).
   * In practice because of weirness around FP representations of
   * numbers with fractional parts ending in a "9", we have
   * 0.009*100 = 0.8999999999999999
   * This means when we multiply by 100 to convert from a plain number
   * to percentage representation, some precision is lost.
   *
   * In practice, this will be a rare edge case that won't really
   * impact users anyway (no one is ever likely to notice this,
   * especially since it is not incorrect to have the extra zero),
   * so putting in a fix is not worth it in terms of the additional
   *  code complexity that would be introduced
   */
  [0.009, "0.90%"],

  // Note: .10 IS significant in this case
  [0.095 / 100, "0.10%"],
  [0.0095 / 100, "0.01%"],
  [0.001 / 100, "1.0e-3%"],
  [0.00095 / 100, "950.0e-6%"],
  [0.000999999 / 100, "1.0e-3%"],
  [0.00012335234 / 100, "123.4e-6%"],
  [0.000_000_999999 / 100, "1.0e-6%"],
  [0.000_000_02341253 / 100, "23.4e-9%"],
  [0.000_000_000_999999 / 100, "1.0e-9%"],
];

describe("range formatter, using default options for NumberKind.PERCENT, `.stringFormat()`", () => {
  defaultGenericNumTestCases.forEach(([input, output]) => {
    it(`returns the correct string in case: ${input}`, () => {
      const formatter = new PerRangeFormatter([input], defaultPercentOptions);
      expect(formatter.stringFormat(input)).toEqual(output);
    });
  });
});
