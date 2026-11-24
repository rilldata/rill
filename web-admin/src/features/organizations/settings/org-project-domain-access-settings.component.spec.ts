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
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import OrgProjectDomainAccessSettings from "./OrgProjectDomainAccessSettings.svelte";

vi.mock("@rilldata/web-admin/client/http-client", () => ({
  httpClient: vi.fn(),
}));

const ORG = "acme";
const request = vi.mocked(httpClient);

beforeEach(() => {
  request.mockReset();
  request.mockImplementation((() =>
    Promise.resolve({ organization: { name: ORG } })) as typeof httpClient);
});
afterEach(() => {
  cleanup();
});

describe("OrgProjectDomainAccessSettings", () => {
  it("saves the setting when the switch is toggled with the keyboard", async () => {
    render(OrgProjectDomainAccessSettings, {
      props: { organization: ORG },
      context: new Map<string | symbol, unknown>([
        [
          "$$_queryClient",
          new QueryClient({ defaultOptions: { queries: { retry: false } } }),
        ],
      ]),
    });

    // The switch is disabled until the org has loaded.
    const toggle = screen.getByRole("switch");
    await waitFor(() => expect(toggle.hasAttribute("disabled")).toBe(false));
    // Bits UI toggles on Space without a click event.
    await fireEvent.keyDown(toggle, { key: " " });

    await waitFor(() =>
      expect(request).toHaveBeenCalledWith(
        expect.objectContaining({
          method: "PATCH",
          url: `/v1/orgs/${ORG}`,
          data: { disableProjectDomainAccess: true },
        }),
      ),
    );
  });
});
