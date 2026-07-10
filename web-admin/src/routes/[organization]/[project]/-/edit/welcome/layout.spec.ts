import { CreateProjectBranchName } from "@rilldata/web-admin/features/projects/publish-project";
import { isRedirect } from "@sveltejs/kit";
import { beforeEach, describe, expect, it, vi } from "vitest";
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

async function callLoad(pathname: string): Promise<unknown> {
  try {
    await load({
      params: { organization: ORG, project: PROJECT },
      url: new URL(`http://localhost${pathname}`),
    } as never);
    return undefined;
  } catch (e) {
    return e;
  }
}

describe("edit/welcome/+layout load", () => {
  beforeEach(() => {
    isProjectWelcomeStepMock.mockReset();
  });

  it("does not redirect when the project is still in the welcome step", async () => {
    isProjectWelcomeStepMock.mockReturnValue(true);

    const result = await callLoad(
      `/${ORG}/${PROJECT}/@${CreateProjectBranchName}/-/edit/welcome`,
    );
    expect(result).toBeUndefined();
  });

  it("redirects to /-/edit on the same branch when no longer in the welcome step", async () => {
    isProjectWelcomeStepMock.mockReturnValue(false);

    const result = await callLoad(
      `/${ORG}/${PROJECT}/@${CreateProjectBranchName}/-/edit/welcome`,
    );
    expect(isRedirect(result)).toBe(true);
    if (!isRedirect(result)) return;
    expect(result.status).toBe(307);
    expect(result.location).toBe(
      `/${ORG}/${PROJECT}/@${CreateProjectBranchName}/-/edit`,
    );
  });

  it("does not redirect when no branch is present in the URL", async () => {
    isProjectWelcomeStepMock.mockReturnValue(false);

    const result = await callLoad(`/${ORG}/${PROJECT}/-/edit/welcome`);
    expect(result).toBeUndefined();
  });
});
