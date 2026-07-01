import { fireEvent, render, screen } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@rilldata/web-common/components/dropdown-menu", async () => {
  const Passthrough = (await import("./__fixtures__/SlotPassthrough.svelte"))
    .default;
  const Clickable = (await import("./__fixtures__/ClickableItem.svelte"))
    .default;
  return {
    Sub: Passthrough,
    SubTrigger: Passthrough,
    SubContent: Passthrough,
    CheckboxItem: Clickable,
  };
});

vi.mock("@rilldata/web-common/lib/i18n/gen/runtime", async (importOriginal) => {
  const actual =
    await importOriginal<
      typeof import("@rilldata/web-common/lib/i18n/gen/runtime")
    >();
  return {
    ...actual,
    getLocale: vi.fn(() => "en"),
    setLocale: vi.fn(),
  };
});

import { setLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import LanguageSwitcher from "./LanguageSwitcher.svelte";

describe("LanguageSwitcher", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls setLocale when a different locale is selected", async () => {
    render(LanguageSwitcher);

    const esButton = screen.getByText("Español");
    await fireEvent.click(esButton);

    expect(setLocale).toHaveBeenCalledWith("es");
  });

  it("does not call setLocale when current locale is selected", async () => {
    render(LanguageSwitcher);

    const enButton = screen.getByText("English");
    await fireEvent.click(enButton);

    expect(setLocale).not.toHaveBeenCalled();
  });
});
