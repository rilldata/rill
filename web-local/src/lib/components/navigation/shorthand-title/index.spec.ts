import { shorthandTitle } from "./";

const data = [
  { input: "Product", output: "Pr" },
  { input: "Product Dashboard", output: "PD" },
  { input: undefined, output: undefined },
  { input: "wonderful things!", output: "WT" },
  { input: "growth and usage", output: "GU" },
  { input: "Rill KPI Dashboards", output: "RK" },
];

describe("shorthand-title", () => {
  it("correctly parses title to shorthand", () => {
    data.forEach(({ input, output }) => {
      expect(shorthandTitle(input)).toEqual(output);
    });
  });
});
