import { setRuntimeEditEnvironment } from "@rilldata/web-common/features/entity-management/edit-environment.ts";
import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { DeleteTarget } from "./delete-target";
import DeleteConfirmationDialog from "./DeleteConfirmationDialog.svelte";

const canvas: DeleteTarget = {
  type: "file",
  path: "/dashboards/orders_overview.yaml",
  resource: { kind: "canvas", name: "orders_overview" },
  dependents: [],
};

function renderDialog(target: DeleteTarget) {
  const onConfirm = vi.fn();
  render(DeleteConfirmationDialog, {
    props: { open: true, target, onConfirm },
  });
  return onConfirm;
}

describe("DeleteConfirmationDialog", () => {
  afterEach(() => setRuntimeEditEnvironment("local"));

  it("names the dashboard and its file, and says the delete can't be undone", () => {
    renderDialog(canvas);

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("Delete canvas dashboard?");
    expect(dialog).toHaveTextContent(
      "The canvas dashboard orders_overview and its file /dashboards/orders_overview.yaml will be deleted from the project.",
    );
    expect(dialog).toHaveTextContent("This can't be undone in Rill.");
  });

  it("deletes only after the user confirms", async () => {
    const onConfirm = renderDialog(canvas);
    expect(onConfirm).not.toHaveBeenCalled();

    await fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(onConfirm).toHaveBeenCalledExactlyOnceWith(canvas);
    await waitFor(() =>
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument(),
    );
  });

  it("deletes nothing when the user cancels", async () => {
    const onConfirm = renderDialog(canvas);

    await fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    await waitFor(() =>
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument(),
    );
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("uses the wording of the resource kind", () => {
    renderDialog({
      type: "file",
      path: "/dashboards/orders_explore.yaml",
      resource: { kind: "explore", name: "orders_explore" },
      dependents: [],
    });

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("Delete explore dashboard?");
    expect(dialog).toHaveTextContent(
      "The explore dashboard orders_explore and its file /dashboards/orders_explore.yaml will be deleted from the project.",
    );
  });

  it("describes a file without a named resource by its path", () => {
    renderDialog({ type: "file", path: "/notes/README.md", dependents: [] });

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("Delete file?");
    expect(dialog).toHaveTextContent(
      "The file /notes/README.md will be deleted from the project.",
    );
  });

  it("lists the files a folder delete removes, relative to the folder", () => {
    renderDialog({
      type: "folder",
      path: "/dashboards",
      files: [
        "/dashboards/orders.yaml",
        "/dashboards/archive/orders_2024.yaml",
      ],
    });

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("Delete folder?");
    expect(dialog).toHaveTextContent(
      "The folder /dashboards and everything in it will be deleted from the project, including files the file explorer doesn't show, such as paths ignored in rill.yaml.",
    );
    expect(
      screen.getAllByRole("listitem").map((item) => item.textContent),
    ).toEqual(["orders.yaml", "archive/orders_2024.yaml"]);
  });

  it("lists the files that depend on the resource", () => {
    renderDialog({
      type: "file",
      path: "/metrics/orders_metrics.yaml",
      resource: { kind: "metrics_view", name: "orders_metrics" },
      dependents: [
        "/dashboards/orders_canvas.yaml",
        "/dashboards/orders_explore.yaml",
      ],
    });

    expect(screen.getByRole("alertdialog")).toHaveTextContent(
      "These 2 files depend on it and will have errors until you update them:",
    );
    expect(
      screen.getAllByRole("listitem").map((item) => item.textContent),
    ).toEqual([
      "/dashboards/orders_canvas.yaml",
      "/dashboards/orders_explore.yaml",
    ]);
  });

  it("says when a single file depends on the resource", () => {
    renderDialog({ ...canvas, dependents: ["/alerts/orders_drop.yaml"] });

    expect(screen.getByRole("alertdialog")).toHaveTextContent(
      "This file depends on it and will have errors until you update it:",
    );
  });

  it("explains how to restore a file in the Cloud editor", () => {
    setRuntimeEditEnvironment("cloud");
    renderDialog(canvas);

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent(
      "Until you publish, a file that existed when your branch last synced with production can be restored from the list of changed files, as it was at that sync; later edits to it are lost. A file added on this branch can't be restored in Rill",
    );
    expect(dialog).not.toHaveTextContent("This can't be undone in Rill.");
  });

  it("renders markup in a file name as text", () => {
    const path = "/notes/<img src=x onerror=alert(1)>.md";
    renderDialog({ type: "file", path, dependents: [] });

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent(path);
    expect(dialog.querySelector("img")).toBeNull();
  });
});
