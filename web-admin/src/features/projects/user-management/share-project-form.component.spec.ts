// @vitest-environment jsdom
import { httpClient } from "@rilldata/web-admin/client/http-client";
import { QueryClient } from "@tanstack/svelte-query";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import type { AxiosRequestConfig } from "axios";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import ShareProjectForm from "./ShareProjectForm.svelte";

vi.mock("@rilldata/web-admin/client/http-client", () => ({
  httpClient: vi.fn(),
}));
vi.mock("$app/stores", async () => {
  const { readable } = await import("svelte/store");
  return {
    page: readable({
      url: new URL("http://localhost/acme/sales"),
      params: { organization: "acme", project: "sales" },
    }),
  };
});

const ORG = "acme";
const PROJECT = "sales";
const PROJECT_MEMBERS_URL = `/v1/orgs/${ORG}/projects/${PROJECT}/members`;
const PROJECT_INVITES_URL = `/v1/orgs/${ORG}/projects/${PROJECT}/invites`;
const PROJECT_USERGROUPS_URL = `/v1/orgs/${ORG}/projects/${PROJECT}/usergroups`;
// These need the org-level permission to read org members, which project admins may not have.
const ORG_MEMBERS_URL = `/v1/orgs/${ORG}/members`;
const ORG_INVITES_URL = `/v1/orgs/${ORG}/invites`;
const ORG_USERGROUPS_URL = `/v1/orgs/${ORG}/usergroups`;
const ORG_AUTOGROUP_MEMBERS_URL = `/v1/orgs/${ORG}/usergroups/autogroup:members/members`;

const request = vi.mocked(httpClient);

function requestedUrls() {
  return request.mock.calls.map(([config]) => config.url);
}

function renderForm(readOrgMembers: boolean) {
  return render(ShareProjectForm, {
    props: {
      organization: ORG,
      project: PROJECT,
      manageProjectAdmins: true,
      manageOrgAdmins: false,
      manageOrgMembers: false,
      readOrgMembers,
      enabled: true,
    },
    context: new Map<string | symbol, unknown>([
      [
        "$$_queryClient",
        new QueryClient({ defaultOptions: { queries: { retry: false } } }),
      ],
    ]),
  });
}

/** Waits for the invite input's focus handler, which searches the server for users to suggest. */
async function focusInviteInput() {
  await fireEvent.focus(screen.getByRole("textbox"));
  await new Promise((resolve) => setTimeout(resolve, 50));
}

beforeEach(() => {
  request.mockReset();
  request.mockImplementation(((config: AxiosRequestConfig) =>
    Promise.resolve(
      config.url === "/v1/users/current"
        ? { user: { email: "me@acme.com", displayName: "Me" } }
        : // Every list is empty: the tests only check which lists the form asks for.
          { members: [], invites: [], nextPageToken: "" },
    )) as typeof httpClient);
  // The form observes its member list for infinite scrolling, which jsdom does not implement.
  vi.stubGlobal(
    "IntersectionObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("ShareProjectForm", () => {
  it("lists the org's invites, groups and members for a user who can read them", async () => {
    renderForm(true);

    await waitFor(() =>
      expect(requestedUrls()).toEqual(
        expect.arrayContaining([
          PROJECT_MEMBERS_URL,
          PROJECT_INVITES_URL,
          PROJECT_USERGROUPS_URL,
          ORG_INVITES_URL,
          ORG_USERGROUPS_URL,
          ORG_AUTOGROUP_MEMBERS_URL,
        ]),
      ),
    );

    await focusInviteInput();
    expect(requestedUrls()).toContain(ORG_MEMBERS_URL);
  });

  it("only lists the project's own members for a project admin who cannot read the org's members", async () => {
    renderForm(false);

    await waitFor(() =>
      expect(requestedUrls()).toEqual(
        expect.arrayContaining([
          PROJECT_MEMBERS_URL,
          PROJECT_INVITES_URL,
          PROJECT_USERGROUPS_URL,
        ]),
      ),
    );
    await focusInviteInput();

    // The admin server would reject each of these with a 403.
    const urls = requestedUrls();
    expect(urls).not.toContain(ORG_MEMBERS_URL);
    expect(urls).not.toContain(ORG_INVITES_URL);
    expect(urls).not.toContain(ORG_USERGROUPS_URL);
    expect(urls).not.toContain(ORG_AUTOGROUP_MEMBERS_URL);
  });
});
