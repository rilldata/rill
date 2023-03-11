// import {
//   FormatterFactoryOptions,
//   NumberKind,
//   NumberParts,
// } from "../humanizer-types";
// import { DefaultHumanizer } from "./default";

// type TestArgs = [number, number, number, boolean, boolean?];

// const baseOptions: FormatterFactoryOptions = {
//   strategy: "default",
//   padWithInsignificantZeros: true,
//   numberKind: NumberKind.ANY,
//   maxDigitsRightSmallNums: 3,
//   maxDigitsRightSuffixNums: 2,
// };

// const testCases: [
//   number,
//   {
//     maxDigitsRightSmallNums?: number;
//     maxDigitsRightSuffixNums?: number;
//     padWithInsignificantZeros?: boolean;
//     numberKind: NumberKind;
//   },
//   string
// ][] = [
//   [999_999.9999, , "1.00M"],

//   [12_345_789, , "12.35M"],
//   [12_345_789, { numberKind: NumberKind.DOLLAR }, "$12.35M"],
//   [12_345_789, { numberKind: NumberKind.PERCENT }, "1.23B%"],

//   [12_345.789012, , "12346."],
//   [12_345.789012, { numberKind: NumberKind.DOLLAR }, "$12346."],
//   [12_345.789012, { numberKind: NumberKind.PERCENT }, "1.23M%"],

//   [12.345789012, , "12.346"],
//   [12.345789012, { numberKind: NumberKind.DOLLAR }, "$12.35"],
//   [12.345789012, { numberKind: NumberKind.PERCENT }, "1235.%"],

//   [0.0012345789012, , "0.001"],
//   [0.0012345789012, { numberKind: NumberKind.DOLLAR }, "$1.23e-3"],
//   [0.0012345789012, { numberKind: NumberKind.PERCENT }, "0.123%"],

//   [0.000_000_2345789012, , "234.58e-9"],
//   [0.000_000_2345789012, { numberKind: NumberKind.DOLLAR }, "$234.58e-9"],
//   [0.000_000_2345789012, { numberKind: NumberKind.PERCENT }, "23.46e-6%"],
// ];

// describe("default formatter, default options `.stringFormat()`", () => {
//   testCases.forEach(([input, options, output]) => {
//     it(`returns the correct split string in case: ${input}`, () => {
//       const formatter = new DefaultHumanizer([input], {
//         ...baseOptions,
//         ...options,
//       });
//       // expect(formatter.stringFormat(input)).toEqual(output);
//     });
//   });
// });

// const currencyTestCases: [
//   number,
//   {
//     maxDigitsRightSmallNums?: number;
//     maxDigitsRightSuffixNums?: number;
//     padWithInsignificantZeros?: boolean;
//   },
//   string
// ][] = [
//   // integers
//   [12_345_789, , "$12.35M"],
//   [2_345_789, , "$2.35M"],
//   [999_999, , "$1.0M"],
//   [345_789, , "$345.8k"],
//   [45_789, , "$45.8k"],
//   [5_789, , "$5.8k"],
//   [789, , "$789.00"],
//   [89, , "$89.00"],
//   [9, , "$9.00"],
//   // non integers[12_345_789, , "$12.35M"],
//   [2_345_789.1234123, , "$2.35M"],
//   [345_789.1234123, , "$345789."],
//   [45_789.1234123, , "$45789."],
//   [5_789.1234123, , "$5789."],
//   [789.1234123, , "$789.12"],
//   [89.1234123, , "$89.12"],
//   [9.1234123, , "$9.12"],
// ];

// describe("default formatter, currency cases `.stringFormat()`", () => {
//   currencyTestCases.forEach(([input, options, output]) => {
//     it(`returns the correct split string in case: ${input}`, () => {
//       const formatter = new DefaultHumanizer([input], {
//         ...baseOptions,
//         ...options,
//         ...{ numberKind: NumberKind.DOLLAR },
//       });
//       // expect(formatter.stringFormat(input)).toEqual(output);
//     });
//   });
// });
