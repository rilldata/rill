// @vitest-environment jsdom
import { getAdminServiceListPersonalFilesQueryKey } from "@rilldata/web-admin/client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { cleanup, render } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import PersonalFilesChangeListener from "./PersonalFilesChangeListener.svelte";

afterEach(() => {
  cleanup();
  queryClient.clear();
});

describe("PersonalFilesChangeListener", () => {
  it("refetches the project's personal files on the bus event, even with no list on screen", async () => {
    // Cached but unobserved, as after leaving the page: a list mounted later
    // reuses its cache (refetchOnMount is off), so it must be refetched now.
    const key = getAdminServiceListPersonalFilesQueryKey("acme", "ecom");
    const otherProjectKey = getAdminServiceListPersonalFilesQueryKey(
      "acme",
      "other",
    );
    const fetchFiles = vi.fn().mockResolvedValue({ files: ["new"] });
    const fetchOtherFiles = vi.fn().mockResolvedValue({ files: [] });
    queryClient.setQueryDefaults(key, { queryFn: fetchFiles });
    queryClient.setQueryDefaults(otherProjectKey, { queryFn: fetchOtherFiles });
    queryClient.setQueryData(key, { files: [] });
    queryClient.setQueryData(otherProjectKey, { files: [] });

    const { unmount } = render(PersonalFilesChangeListener, {
      organization: "acme",
      project: "ecom",
    });

    eventBus.emit("personal-files-changed");
    await vi.waitFor(() =>
      expect(queryClient.getQueryData(key)).toEqual({ files: ["new"] }),
    );
    expect(fetchFiles).toHaveBeenCalledOnce();
    expect(fetchOtherFiles).not.toHaveBeenCalled();

    // Leaving the project stops listening.
    unmount();
    eventBus.emit("personal-files-changed");
    await new Promise((resolve) => setTimeout(resolve));
    expect(fetchFiles).toHaveBeenCalledOnce();
  });
});
