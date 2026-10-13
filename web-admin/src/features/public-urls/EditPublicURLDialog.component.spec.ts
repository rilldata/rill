import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { writable } from "svelte/store";
import EditPublicURLDialog from "./EditPublicURLDialog.svelte";

const mocks = vi.hoisted(() => ({
  update: vi.fn(),
  invalidate: vi.fn(),
  write: vi.fn(),
}));
vi.mock("$app/state", () => ({
  page: { params: { organization: "org", project: "project" } },
}));
vi.mock("@rilldata/web-admin/client", () => ({
  createAdminServiceUpdateMagicAuthToken: () =>
    writable({ mutateAsync: mocks.update, isPending: false }),
  getAdminServiceListMagicAuthTokensQueryKey: () => ["tokens"],
}));
vi.mock("@tanstack/svelte-query", () => ({
  useQueryClient: () => ({ invalidateQueries: mocks.invalidate }),
}));

beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal("localStorage", { getItem: () => null, setItem: () => {} });
  mocks.update.mockResolvedValue({});
  mocks.invalidate.mockResolvedValue(undefined);
  mocks.write.mockResolvedValue(undefined);
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    value: { writeText: mocks.write },
  });
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

function open(passwordProtected: boolean) {
  render(EditPublicURLDialog, {
    open: true,
    token: { id: "token", displayName: "Label", passwordProtected },
  });
}

async function save() {
  await fireEvent.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() => expect(mocks.update).toHaveBeenCalled());
}

it("keeps the password when only the label changes", async () => {
  open(true);
  await fireEvent.input(screen.getByLabelText("Label"), {
    target: { value: "Renamed" },
  });
  await save();
  expect(mocks.update).toHaveBeenCalledWith({
    tokenId: "token",
    data: { displayName: "Renamed" },
  });
});

it("sets a password and preserves its display after refreshing the list", async () => {
  open(true);
  await fireEvent.input(screen.getByPlaceholderText("Set a new password"), {
    target: { value: "new password" },
  });
  await save();
  expect(mocks.update).toHaveBeenCalledWith({
    tokenId: "token",
    data: { password: "new password" },
  });
  await waitFor(() => expect(screen.getByText("new password")).toBeTruthy());
});

it("removes a password and discards a replacement entered earlier", async () => {
  open(true);
  const input = screen.getByPlaceholderText("Set a new password");
  await fireEvent.input(input, { target: { value: "new password" } });
  await fireEvent.click(
    screen.getByRole("switch", { name: "Remove password" }),
  );
  expect((input as HTMLInputElement).value).toBe("");
  expect(input.hasAttribute("disabled")).toBe(true);
  await save();
  expect(mocks.update).toHaveBeenCalledWith({
    tokenId: "token",
    data: { removePassword: true },
  });
});

it("requires a valid password when adding protection, including generated passwords", async () => {
  open(false);
  await fireEvent.click(
    screen.getByRole("switch", { name: "Require password" }),
  );
  const saveButton = screen.getByRole("button", { name: "Save" });
  expect(saveButton.hasAttribute("disabled")).toBe(true);
  const input = screen.getByPlaceholderText("Password");
  await fireEvent.input(input, { target: { value: "é".repeat(37) } });
  expect(saveButton.hasAttribute("disabled")).toBe(true);
  await fireEvent.click(
    screen.getByRole("button", { name: "Generate and copy secure password" }),
  );
  await waitFor(() => expect(saveButton.hasAttribute("disabled")).toBe(false));
  await waitFor(() =>
    expect(
      screen.getByRole("button", { name: "Copied password" }),
    ).toBeTruthy(),
  );
  const password = (input as HTMLInputElement).value;
  expect(password).toHaveLength(16);
  await save();
  expect(mocks.update).toHaveBeenCalledWith({
    tokenId: "token",
    data: { password },
  });
  expect(mocks.write).toHaveBeenCalledWith(password);
});

it("does not claim a generated password was copied when clipboard access fails", async () => {
  vi.spyOn(console, "error").mockImplementation(() => {});
  mocks.write.mockRejectedValue(new Error("Permission denied"));
  open(false);
  await fireEvent.click(
    screen.getByRole("switch", { name: "Require password" }),
  );
  await fireEvent.click(
    screen.getByRole("button", { name: "Generate and copy secure password" }),
  );
  await waitFor(() => expect(mocks.write).toHaveBeenCalled());
  expect(screen.queryByRole("button", { name: "Copied password" })).toBeNull();
  expect(
    (screen.getByPlaceholderText("Password") as HTMLInputElement).value,
  ).toHaveLength(16);
});
