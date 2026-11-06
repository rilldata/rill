import { overwriteGetLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import RemovePartitionButton from "./RemovePartitionButton.svelte";

beforeEach(() => overwriteGetLocale(() => "en"));
afterEach(cleanup);

describe("partition removal action", () => {
  it("provides a named icon action and invokes removal when clicked", async () => {
    const onRemove = vi.fn();
    render(RemovePartitionButton, { onRemove });
    const button = screen.getByRole("button", { name: "Remove" });
    expect(button).toHaveAccessibleName("Remove");
    expect(button).not.toHaveTextContent("Remove");
    expect(button.querySelector("svg")).not.toBeNull();
    await fireEvent.click(button);
    expect(onRemove).toHaveBeenCalledOnce();
  });
});
