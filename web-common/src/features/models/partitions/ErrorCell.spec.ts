import { overwriteGetLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import { beforeEach, afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import ErrorCell from "./ErrorCell.svelte";

beforeEach(() => overwriteGetLocale(() => "en"));
afterEach(cleanup);

describe("partition error details", () => {
  it("starts collapsed and lets the user expand the full error", async () => {
    const error = `Connection failed\n${"Long error detail ".repeat(100)}`;
    const { container } = render(ErrorCell, { error });
    const details = container.querySelector("details");
    expect(details).not.toBeNull();
    expect(details?.open).toBe(false);
    const summary = container.querySelector("summary");
    expect(summary).not.toBeNull();
    if (!summary) throw new Error("Missing error disclosure");
    await fireEvent.click(summary);
    expect(details?.open).toBe(true);
    expect(container.querySelector("pre")?.textContent).toBe(error);
    await fireEvent.click(summary);
    expect(details?.open).toBe(false);
  });

  it("shows a dash when no error exists", () => {
    const { container } = render(ErrorCell, { error: undefined });
    expect(screen.getByText("-")).toBeInTheDocument();
    expect(container.querySelector("details")).toBeNull();
  });
});
