import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

export function createSizeChangeHandler(name: string, type: string) {
  return (node: HTMLElement) => {
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        eventBus.emit("dashboard-resized", {
          name,
          type,
          width,
          height,
        });
      }
    });

    resizeObserver.observe(node);

    return {
      destroy() {
        resizeObserver.disconnect();
      },
    };
  };
}
