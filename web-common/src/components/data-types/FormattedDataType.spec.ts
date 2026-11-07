import { render } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import FormattedDataType from "./FormattedDataType.svelte";

// Tailwind emits the `text-fg-*` utilities in alphabetical order, so when an
// element carries two of them the alphabetically later one wins no matter how
// the classes are ordered in the markup. A data-type component must therefore
// never bake a text color of its own into the class list next to the `color`
// it was given: on the dark cell tooltip, `text-fg-primary` beat
// `text-fg-inverse` and timestamp values rendered dark on dark.
describe("FormattedDataType", () => {
  it.each([
    ["TIMESTAMP", "2026-10-16 13:21:00"],
    ["VARCHAR", "hello"],
    ["INTEGER", 42],
  ])("renders a %s value with only the requested text color", (type, value) => {
    const { container } = render(FormattedDataType, {
      props: { type, value, color: "text-fg-inverse" },
    });

    const textColors = [...container.querySelector("span")!.classList].filter(
      (cls) => cls.startsWith("text-fg-"),
    );
    expect(textColors).toEqual(["text-fg-inverse"]);
  });

  it("falls back to the primary text color when no color is given", () => {
    const { container } = render(FormattedDataType, {
      props: { type: "TIMESTAMP", value: "2026-10-16 13:21:00" },
    });

    const textColors = [...container.querySelector("span")!.classList].filter(
      (cls) => cls.startsWith("text-fg-"),
    );
    expect(textColors).toEqual(["text-fg-primary"]);
  });
});
