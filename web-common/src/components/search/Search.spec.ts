import { fireEvent, render, screen } from "@testing-library/svelte";
import { tabbable } from "tabbable";
import { describe, expect, it, vi } from "vitest";
import Search from "./Search.svelte";

describe("Search keyboard focus", () => {
  it.each(["Enter", "NumpadEnter"])(
    "submits %s without native form navigation",
    async (code) => {
      const onSubmit = vi.fn();
      render(Search, { props: { value: "", onSubmit, autofocus: false } });
      expect(
        await fireEvent.keyDown(screen.getByRole("textbox"), {
          key: "Enter",
          code,
        }),
      ).toBe(false);
      expect(onSubmit).toHaveBeenCalledOnce();
    },
  );
  it("leaves the search field on Escape without clearing the query", async () => {
    render(Search, {
      props: { value: "Device", autofocus: false, retainValueOnMount: true },
    });
    const search = screen.getByRole("textbox");
    search.focus();
    expect(search).toHaveFocus();
    await fireEvent.keyDown(search, { key: "Escape", code: "Escape" });
    expect(search).not.toHaveFocus();
    expect(search).toHaveValue("Device");
  });
  it("makes the text box the first keyboard focus target", () => {
    const { container } = render(Search, {
      props: { value: "", autofocus: false },
    });

    expect(tabbable(container, { displayCheck: "none" })[0]).toBe(
      screen.getByRole("textbox"),
    );
  });
});
