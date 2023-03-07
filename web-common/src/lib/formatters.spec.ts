import { justEnoughPrecision } from "./formatters";

interface TestCase {
  input: number;
  output: string;
}

const onlyIntegers = [
  { input: 0, output: "0" },
  { input: 1, output: "1" },
  { input: 12, output: "12" },
  { input: 123, output: "123" },
  { input: 1234, output: "1234" },
  { input: 12345, output: "12345" },
  { input: 123456, output: "123456" },
  { input: 1234567, output: "1234567" },
  { input: 12345678, output: "12345678" },
  { input: 123456789, output: "123456789" },
  { input: 12345678900000, output: "12345678900000" },
];

const zeros = [
  { input: 0.1, output: "0.1" },
  { input: 0.12, output: "0.12" },
  { input: 0.123, output: "0.123" },
  { input: 0.1234, output: "0.1234" },
  { input: 0.12345, output: "0.1234" },
];

const ones = [
  { input: 1.0, output: "1" },
  { input: 1.1, output: "1.1" },
  { input: 1.12, output: "1.12" },
  { input: 1.123, output: "1.123" },
  { input: 1.1234, output: "1.1234" },
  { input: 1.12345, output: "1.1234" },
];

const twos = [
  { input: 12.0, output: "12" },
  { input: 12.1, output: "12.1" },
  { input: 12.12, output: "12.12" },
  { input: 12.123, output: "12.123" },
  { input: 12.1234, output: "12.123" },
];

const threes = [
  { input: 123.0, output: "123" },
  { input: 123.1, output: "123.1" },
  { input: 123.12, output: "123.12" },
  { input: 123.123, output: "123.12" },
];

const fours = [
  { input: 1234.0, output: "1234" },
  { input: 1234.1, output: "1234.1" },
  { input: 1234.12, output: "1234.1" },
];

// add negative versions of all test cases
const withNegativesToo = (testCases: TestCase[]) => [
  ...testCases,
  ...testCases.map(({ input, output }) => ({
    input: -input,
    output: input === 0 ? `${input}` : `-${output}`,
  })),
];

const allTestCases = withNegativesToo([
  ...onlyIntegers,
  ...zeros,
  ...ones,
  ...twos,
  ...threes,
  ...fours,
]);

describe("justEnoughPrecision", () => {
  it("throws on non-numbers", () => {
    expect(() => justEnoughPrecision("foo" as any)).toThrow();
    expect(() => justEnoughPrecision(undefined as any)).toThrow();
  });
  it("returns formatted values, removing floating points depending on number of significant digits", () => {
    allTestCases.forEach(({ input, output }) => {
      expect(justEnoughPrecision(input)).toBe(output);
    });
  });
});
