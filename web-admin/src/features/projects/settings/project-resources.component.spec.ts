// @vitest-environment jsdom
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";
import HibernateProject from "./HibernateProject.svelte";
import ProjectSlotsSection from "@rilldata/web-admin/features/branches/ProjectSlotsSection.svelte";
import { featureFlags } from "@rilldata/web-common/features/feature-flags";

const mocks = vi.hoisted(() => ({
  update: vi.fn(),
  hibernate: vi.fn(),
  invalidate: vi.fn(),
  notify: vi.fn(),
}));

const project = writable({
  data: {
    project: { prodSlots: "2", devSlots: "3" },
    deployment: { id: "deployment" },
    projectPermissions: { manageProject: true },
  },
});

vi.mock("@rilldata/web-admin/client", () => ({
  createAdminServiceGetProject: () => project,
  createAdminServiceUpdateProject: () =>
    writable({ mutateAsync: mocks.update, isPending: false }),
  createAdminServiceHibernateProject: () =>
    writable({ mutateAsync: mocks.hibernate, isPending: false }),
  getAdminServiceGetProjectQueryKey: () => ["project"],
  getAdminServiceListDeploymentsQueryKey: () => ["deployments"],
  getAdminServiceListProjectsForOrganizationQueryKey: () => ["projects"],
}));
vi.mock("@rilldata/web-common/lib/svelte-query/globalQueryClient", () => ({
  queryClient: { invalidateQueries: mocks.invalidate },
}));
vi.mock("@rilldata/web-common/lib/event-bus/event-bus", () => ({
  eventBus: { emit: mocks.notify },
}));
vi.mock("@rilldata/web-common/features/feature-flags", async () => {
  const { writable } = await import("svelte/store");
  return { featureFlags: { cloudEditing: writable(false) } };
});

beforeEach(() => {
  vi.clearAllMocks();
  featureFlags.cloudEditing.set(false);
  mocks.update.mockResolvedValue({});
  mocks.hibernate.mockResolvedValue({});
  project.set({
    data: {
      project: { prodSlots: "2", devSlots: "3" },
      deployment: { id: "deployment" },
      projectPermissions: { manageProject: true },
    },
  });
});
afterEach(cleanup);

const props = { organization: "test-org", project: "test-project" };

it("shows development slot controls only when cloud editing is enabled", async () => {
  render(ProjectSlotsSection, props);
  expect(
    screen.getByRole("spinbutton", { name: "Production slots" }),
  ).toBeTruthy();
  expect(
    screen.queryByRole("spinbutton", { name: "Development slots" }),
  ).toBeNull();

  featureFlags.cloudEditing.set(true);
  await waitFor(() =>
    expect(
      screen.getByRole("spinbutton", { name: "Development slots" }),
    ).toBeTruthy(),
  );

  featureFlags.cloudEditing.set(false);
  await waitFor(() =>
    expect(
      screen.queryByRole("spinbutton", { name: "Development slots" }),
    ).toBeNull(),
  );
  expect(
    screen.getByRole("spinbutton", { name: "Production slots" }),
  ).toBeTruthy();
});

describe.each([
  {
    field: "prodSlots",
    title: "Production slots",
  },
  {
    field: "devSlots",
    title: "Development slots",
  },
])("$title", ({ field, title }) => {
  it("validates slots, preserves edits during refetch, and sends only the selected environment", async () => {
    featureFlags.cloudEditing.set(true);
    render(ProjectSlotsSection, props);
    const input = screen.getByRole("spinbutton", { name: title });
    const save = screen.getByRole("button", {
      name: "Save",
    });
    expect(save.hasAttribute("disabled")).toBe(true);
    for (const value of ["", "0", "-1", "1.5"]) {
      await fireEvent.input(input, { target: { value } });
      expect(save.hasAttribute("disabled")).toBe(true);
    }
    await fireEvent.input(input, { target: { value: "4" } });
    project.update((p) => ({
      ...p,
      data: { ...p.data, project: { ...p.data.project, [field]: "2" } },
    }));
    await waitFor(() => expect((input as HTMLInputElement).value).toBe("4"));
    expect(save.hasAttribute("disabled")).toBe(false);
    await fireEvent.submit(input.closest("form")!);
    await waitFor(() =>
      expect(mocks.update).toHaveBeenCalledWith({
        org: "test-org",
        project: "test-project",
        data: { [field]: "4" },
      }),
    );
    await waitFor(() => expect(mocks.invalidate).toHaveBeenCalledTimes(3));
  });

  it("displays quota errors and keeps the input available for correction", async () => {
    mocks.update.mockRejectedValueOnce({
      response: { data: { message: "Slot quota exceeded" } },
    });
    featureFlags.cloudEditing.set(true);
    render(ProjectSlotsSection, props);
    const input = screen.getByRole("spinbutton", { name: title });
    await fireEvent.input(input, { target: { value: "8" } });
    await fireEvent.submit(input.closest("form")!);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toBe("Slot quota exceeded"),
    );
    expect((input as HTMLInputElement).value).toBe("8");
    expect(mocks.invalidate).not.toHaveBeenCalled();
  });

  it("disables editing without manage permission", () => {
    project.update((p) => ({
      ...p,
      data: { ...p.data, projectPermissions: { manageProject: false } },
    }));
    featureFlags.cloudEditing.set(true);
    render(ProjectSlotsSection, props);
    expect(
      screen.getByRole("spinbutton", { name: title }).hasAttribute("disabled"),
    ).toBe(true);
    expect(
      screen.getByRole("button", { name: "Save" }).hasAttribute("disabled"),
    ).toBe(true);
  });
});

it("saves both environments with one button and one request", async () => {
  featureFlags.cloudEditing.set(true);
  render(ProjectSlotsSection, props);
  expect(screen.getAllByRole("form")).toHaveLength(1);
  expect(screen.getAllByRole("button", { name: "Save" })).toHaveLength(1);
  const prodInput = screen.getByRole("spinbutton", {
    name: "Production slots",
  }) as HTMLInputElement;
  const devInput = screen.getByRole("spinbutton", {
    name: "Development slots",
  }) as HTMLInputElement;
  expect(prodInput.value).toBe("2");
  expect(devInput.value).toBe("3");
  expect(prodInput.id).not.toBe(devInput.id);
  await fireEvent.input(prodInput, { target: { value: "8" } });
  await fireEvent.input(devInput, { target: { value: "4" } });
  await fireEvent.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(mocks.update).toHaveBeenCalledExactlyOnceWith({
      org: "test-org",
      project: "test-project",
      data: { prodSlots: "8", devSlots: "4" },
    }),
  );
});

it("does not submit hidden development slots or let them block a production update", async () => {
  featureFlags.cloudEditing.set(true);
  render(ProjectSlotsSection, props);
  await fireEvent.input(
    screen.getByRole("spinbutton", { name: "Development slots" }),
    { target: { value: "0" } },
  );
  await fireEvent.input(
    screen.getByRole("spinbutton", { name: "Production slots" }),
    { target: { value: "4" } },
  );
  expect(
    screen.getByRole("button", { name: "Save" }).hasAttribute("disabled"),
  ).toBe(true);
  featureFlags.cloudEditing.set(false);
  await waitFor(() =>
    expect(
      screen.queryByRole("spinbutton", { name: "Development slots" }),
    ).toBeNull(),
  );
  await fireEvent.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(mocks.update).toHaveBeenCalledExactlyOnceWith({
      org: "test-org",
      project: "test-project",
      data: { prodSlots: "4" },
    }),
  );
});

describe("hibernate project", () => {
  it("keeps the confirmation open when hibernation fails", async () => {
    mocks.hibernate.mockRejectedValueOnce({
      response: { data: { message: "Unable to stop deployment" } },
    });
    render(HibernateProject, props);
    await fireEvent.click(
      screen.getByRole("button", { name: "Hibernate project" }),
    );
    await fireEvent.click(screen.getByRole("button", { name: "Hibernate" }));
    await waitFor(() =>
      expect(mocks.notify).toHaveBeenCalledWith("notification", {
        message: "Unable to stop deployment",
        type: "error",
      }),
    );
    expect(screen.getByRole("alertdialog")).toBeTruthy();
    expect(mocks.invalidate).not.toHaveBeenCalled();
  });

  it("disables hibernation without manage permission", () => {
    project.update((p) => ({
      ...p,
      data: { ...p.data, projectPermissions: { manageProject: false } },
    }));
    render(HibernateProject, props);
    expect(
      screen
        .getByRole("button", {
          name: "Hibernate project",
        })
        .hasAttribute("disabled"),
    ).toBe(true);
  });

  it("requires confirmation, supports cancellation, and refreshes project state", async () => {
    render(HibernateProject, props);
    await fireEvent.click(
      screen.getByRole("button", { name: "Hibernate project" }),
    );
    expect(screen.getByRole("alertdialog")).toBeTruthy();
    expect(mocks.hibernate).not.toHaveBeenCalled();
    await fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(mocks.hibernate).not.toHaveBeenCalled();
    await fireEvent.click(
      screen.getByRole("button", { name: "Hibernate project" }),
    );
    await fireEvent.click(screen.getByRole("button", { name: "Hibernate" }));
    await waitFor(() =>
      expect(mocks.hibernate).toHaveBeenCalledWith({
        org: "test-org",
        project: "test-project",
      }),
    );
    await waitFor(() => expect(mocks.invalidate).toHaveBeenCalledTimes(3));
  });
});
