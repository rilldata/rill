import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { afterEach, describe, expect, it, onTestFinished, vi } from "vitest";
import { MessageContentType, ToolName } from "../../types";
import { getToolConfig } from "./tool-registry";

afterEach(() => {
  document.body.replaceChildren();
  vi.restoreAllMocks();
  queryClient.clear();
});

describe("click_ui tool", () => {
  it("clicks only after the server returns a successful result", () => {
    const button = document.createElement("button");
    button.dataset.aiAction = "dashboard.start-pivot";
    vi.spyOn(button, "getBoundingClientRect").mockReturnValue({
      width: 100,
      height: 30,
    } as DOMRect);
    const onClick = vi.fn();
    button.addEventListener("click", onClick);
    document.body.appendChild(button);

    const callMessage: V1Message = {
      contentData: JSON.stringify({ action_id: "dashboard.start-pivot" }),
    };
    const config = getToolConfig(ToolName.CLICK_UI);

    expect(config.onCall).toBeUndefined();
    config.onResult?.(
      callMessage,
      { contentType: MessageContentType.ERROR },
      "default",
    );
    expect(onClick).not.toHaveBeenCalled();

    config.onResult?.(
      callMessage,
      { contentType: MessageContentType.JSON },
      "default",
    );
    expect(onClick).toHaveBeenCalledOnce();
  });
});

describe("develop_personal_canvas tool", () => {
  it("refreshes the lists that show the dashboard once it is saved, even with none on screen", async () => {
    // Cached but unobserved, as after leaving the page: a list mounted later
    // reuses its cache (refetchOnMount is off), so it must be refetched now.
    const resourcesKey = getRuntimeServiceListResourcesQueryKey("default", {});
    const fetchResources = vi.fn().mockResolvedValue({ resources: [] });
    queryClient.setQueryDefaults(resourcesKey, { queryFn: fetchResources });
    queryClient.setQueryData(resourcesKey, { resources: [] });
    const onPersonalFilesChanged = vi.fn();
    const unsubscribe = eventBus.on(
      "personal-files-changed",
      onPersonalFilesChanged,
    );
    onTestFinished(() => {
      unsubscribe();
    });

    const config = getToolConfig(ToolName.DEVELOP_PERSONAL_CANVAS);

    config.onResult?.(
      undefined,
      { contentType: MessageContentType.ERROR },
      "default",
    );
    await new Promise((resolve) => setTimeout(resolve));
    expect(fetchResources).not.toHaveBeenCalled();
    expect(onPersonalFilesChanged).not.toHaveBeenCalled();

    config.onResult?.(
      undefined,
      { contentType: MessageContentType.JSON },
      "default",
    );
    await vi.waitFor(() => expect(fetchResources).toHaveBeenCalledOnce());
    expect(onPersonalFilesChanged).toHaveBeenCalledOnce();
  });
});
