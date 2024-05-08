import { describe, it, expect } from "vitest";
import { formatMsInterval } from "./intervals";

const nonNumericTestCases = [
  null,
  undefined,
  "foo",
  false,
  [1, 2, true],
  { foo: 6, bar: 7 },
  new Date("1999-9-9"),
  BigInt(1e300),
  new Map(),
  new Set(),
  new WeakMap(),
  new WeakSet(),
  Symbol("foo"),
  () => "blah",
];
describe("formatMsInterval - non numeric inputs", () => {
  nonNumericTestCases.forEach((input) => {
    let inputString;
    try {
      inputString = JSON.stringify(input);
    } catch (error) {
      inputString = input?.toString();
    }

    it(`returns the empty string for non numeric input: ${inputString}`, () => {
      expect(formatMsInterval(input as unknown as number)).toEqual("");
    });
  });
});

const MS = 1;
const SEC = 1000 * MS;
const MIN = 60 * SEC;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const MONTH = 30 * DAY; //eslint-disable-line
const YEAR = 365 * DAY; //eslint-disable-line

const time_formula_normal_cases = [
  ["1.797", "1.8 ms"],
  ["123.7989", "0.12 s"],
  ["793.987", "0.79 s"],
  ["100.9797", "0.1 s"],
  ["1 * SEC", "1 s"],
  ["1.4709879 * SEC", "1.5 s"],
  ["9.49797 * SEC", "9.5 s"],
  ["10 * SEC", "10 s"],
  ["59 * SEC", "59 s"],
  ["1 * MIN", "60 s"],
  ["99.9 * SEC", "1.7 m"],
  ["100 * SEC", "1.7 m"],
  ["59.23451 * MIN", "59 m"],
  ["89.411 * MIN", "89 m"],
  ["89.94353 * MIN", "90 m"],
  ["99 * MIN", "1.7 h"],
  ["99.9 * MIN", "1.7 h"],
  ["100 * MIN", "1.7 h"],
  ["71.936 * HOUR", "72 h"],
  ["72 * HOUR", "3 d"],
  ["99 * HOUR", "4.1 d"],
  ["89.9 * DAY", "90 d"],
  ["90 * DAY", "3 mon"],
  ["99 * DAY", "3.3 mon"],
  ["7.87978 * MONTH", "7.9 mon"],
  ["17.923 * MONTH", "18 mon"],
  ["18 * MONTH", "1.5 y"],
  ["18.0234234 * MONTH", "1.5 y"],
  ["36 * MONTH", "3 y"],
  ["3247 * DAY", "8.9 y"],
  ["43.34523 * YEAR", "43 y"],
  ["99 * YEAR", "99 y"],
  ["99 * YEAR + 6 * SEC", "99 y"],
  ["99 * YEAR + 6.0004 * SEC", "99 y"],
  ["99 * YEAR + 6.99999 * SEC", "99 y"],
  ["99.9 * YEAR", "100 y"],
];

describe("formatMsInterval - normal cases", () => {
  time_formula_normal_cases.forEach(([input, output]) => {
    const ms = eval(input) as number;
    it(`return "${output}" for input: ${ms.toString()}ms (${input})`, () => {
      expect(formatMsInterval(ms)).toEqual(output);
    });
  });
});

describe("formatMsInterval - normal cases, negative", () => {
  time_formula_normal_cases.forEach(([input, output]) => {
    const ms = -eval(input);
    it(`return "${output}" for input: ${ms.toString()}ms (${input})`, () => {
      expect(formatMsInterval(ms)).toEqual("-" + output);
    });
  });
});

const time_formula_special_cases = [
  ["0", "0 s"],
  ["0.0011797", "~0 s"],
  ["0.01231", "~0 s"],
  ["100.234 * YEAR", ">100 y"],
  ["123797.239797 * YEAR", ">100 y"],
  ["123797.239797 * YEAR", ">100 y"],
  ["123797.239797 * YEAR", ">100 y"],

  // infinitesimals
  [0.9, "~0 s"],
  [0.095, "~0 s"],
  [0.0095, "~0 s"],
  [0.001, "~0 s"],
  [0.00095, "~0 s"],
  [0.000999999, "~0 s"],
  [0.00012335234, "~0 s"],
  [0.000_000_999999, "~0 s"],
  [0.000_000_02341253, "~0 s"],
  [0.000_000_000_999999, "~0 s"],

  // negative infinitesimals
  [-0.9, "~0 s"],
  [-0.095, "~0 s"],
  [-0.0095, "~0 s"],
  [-0.001, "~0 s"],
  [-0.00095, "~0 s"],
  [-0.000999999, "~0 s"],
  [-0.00012335234, "~0 s"],
  [-0.000_000_999999, "~0 s"],
  [-0.000_000_02341253, "~0 s"],
  [-0.000_000_000_999999, "~0 s"],

  // huge numbers
  [1e19, ">100 y"],
  [3.2e12, ">100 y"],

  // hugely negative numbers
  [-1e19, "< -100 y"],
  [-3.2e12, "< -100 y"],
];

describe("formatMsInterval - special cases", () => {
  time_formula_special_cases.forEach(([input, output]) => {
    const ms = eval(input.toString()) as number;
    it(`return "${output}" for input: ${ms.toString()}ms (${input})`, () => {
      expect(formatMsInterval(ms)).toEqual(output);
    });
  });
});
