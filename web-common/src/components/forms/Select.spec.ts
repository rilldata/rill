import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import Select from "./Select.svelte";

const OPTIONS = [
  { value: "opt1", label: "Option 1" },
  { value: "opt2", label: "Option 2" },
  { value: "opt3", label: "Option 3" },
];

describe("Select – clearable", () => {
  mockAnimationsForComponentTesting();

  it("shows clear button when clearable and value is set", () => {
    render(Select, {
      props: {
        id: "test-select",
        options: OPTIONS,
        value: "opt1",
        clearable: true,
      },
    });

    expect(screen.getByLabelText("Clear selection")).toBeInTheDocument();
  });

  it("hides clear button when clearable but value is empty", () => {
    render(Select, {
      props: {
        id: "test-select",
        options: OPTIONS,
        value: "",
        clearable: true,
      },
    });

    expect(screen.queryByLabelText("Clear selection")).not.toBeInTheDocument();
  });

  it("hides clear button when not clearable", () => {
    render(Select, {
      props: {
        id: "test-select",
        options: OPTIONS,
        value: "opt1",
        clearable: false,
      },
    });

    expect(screen.queryByLabelText("Clear selection")).not.toBeInTheDocument();
  });

  it("calls onChange with empty string when clear button is clicked", async () => {
    const onChange = vi.fn();

    render(Select, {
      props: {
        id: "test-select",
        options: OPTIONS,
        value: "opt1",
        clearable: true,
        onChange,
      },
    });

    const clearBtn = screen.getByLabelText("Clear selection");
    await clearBtn.click();

    expect(onChange).toHaveBeenCalledWith("");
  });

  it("removes clear button after clearing", async () => {
    const { component } = render(Select, {
      props: {
        id: "test-select",
        options: OPTIONS,
        value: "opt1",
        clearable: true,
        onChange: () => {},
      },
    });

    const clearBtn = screen.getByLabelText("Clear selection");
    await clearBtn.click();

    // After clearing, the component remounts via selectKey.
    // The clear button should be gone since value is now "".
    expect(screen.queryByLabelText("Clear selection")).not.toBeInTheDocument();
  });
});
