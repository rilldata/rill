import { describe, it, expect, beforeEach, vi } from "vitest";
import { isRedirect } from "@sveltejs/kit";
import { load } from "./+layout";

const { isProjectWelcomeStepMock } = vi.hoisted(() => ({
  isProjectWelcomeStepMock: vi.fn<(project: string) => boolean>(),
}));

vi.mock(
  "@rilldata/web-admin/features/welcome/project/welcome-status.ts",
  () => ({
    projectWelcomeStatus: { isProjectWelcomeStep: isProjectWelcomeStepMock },
  }),
);

const ORG = "rilldata";
const PROJECT = "openrtb";

async function callLoad(routeId: string): Promise<unknown> {
  try {
    await load({
      params: { organization: ORG, project: PROJECT },
      route: { id: routeId },
    } as never);
    return undefined;
  } catch (e) {
    return e;
  }
}

describe("edit/+layout load", () => {
  beforeEach(() => {
    isProjectWelcomeStepMock.mockReset();
  });

  it("redirects welcome-step projects off non-welcome edit pages", async () => {
    isProjectWelcomeStepMock.mockReturnValue(true);

    const result = await callLoad("/[organization]/[project]/-/edit");
    expect(isRedirect(result)).toBe(true);
    if (!isRedirect(result)) return;
    expect(result.status).toBe(307);
    expect(result.location).toBe(`/${ORG}/${PROJECT}/@dev/-/edit/welcome`);
  });

  it("does not redirect when already on the welcome page", async () => {
    isProjectWelcomeStepMock.mockReturnValue(true);

    const result = await callLoad("/[organization]/[project]/-/edit/welcome");
    expect(result).toBeUndefined();
  });

  it("does not redirect when project is not in the welcome step", async () => {
    isProjectWelcomeStepMock.mockReturnValue(false);

    const result = await callLoad("/[organization]/[project]/-/edit");
    expect(result).toBeUndefined();
  });
});
