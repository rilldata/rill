import { cellInspectorStore } from "../stores/cellInspectorStore";

export function useCellInspector() {
  let hoverTimeout: number | null = null;
  let lastHoveredValue: string | null = null;
  let isOpen = false;

  const inspectCell = (value: string, event: MouseEvent, persist = false) => {
    if (!value) return;

    const stringValue = String(value);

    if (persist) {
      isOpen = true;
      cellInspectorStore.open(stringValue, {
        x: event.clientX,
        y: event.clientY,
      });
    } else if (!isOpen) {
      cellInspectorStore.open(stringValue, {
        x: event.clientX,
        y: event.clientY,
      });
    }
  };

  const handleMouseEnter = (value: string, e: MouseEvent) => {
    const target = e.currentTarget as HTMLElement;

    // Clear any existing timeout
    if (hoverTimeout) {
      clearTimeout(hoverTimeout);
      hoverTimeout = null;
    }

    // Only track the hovered value but don't open the inspector
    hoverTimeout = window.setTimeout(() => {
      if (value !== lastHoveredValue) {
        lastHoveredValue = value;
        // Store the value but don't open the inspector
        if (isOpen) {
          inspectCell(value, e, true);
        }
      }
    }, 100); // Small delay to prevent flickering
  };

  const handleMouseLeave = () => {
    if (hoverTimeout) {
      clearTimeout(hoverTimeout);
      hoverTimeout = null;
    }
    lastHoveredValue = null;
  };

  const getCellProps = (value: string) => {
    if (!value) return {};

    const stringValue = String(value);

    return {
      onMouseEnter: (e: MouseEvent) => {
        const target = e.currentTarget as HTMLElement;
        if (
          target.scrollWidth > target.offsetWidth ||
          target.scrollHeight > target.offsetHeight
        ) {
          target.title = stringValue;
        }
        handleMouseEnter(stringValue, e);
      },
      onMouseLeave: handleMouseLeave,
      onDblClick: (e: MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        inspectCell(stringValue, e, true);
      },
      onKeyDown: (e: KeyboardEvent) => {
        if (e.code === "Space" || e.code === "Enter") {
          e.preventDefault();
          e.stopPropagation();
          isOpen = !isOpen;
          inspectCell(stringValue, e as unknown as MouseEvent, isOpen);
        } else if (e.key === "Escape" && isOpen) {
          e.preventDefault();
          e.stopPropagation();
          isOpen = false;
          cellInspectorStore.close();
        }
      },
      tabIndex: 0,
      role: "button",
      "aria-label": "Inspect cell",
      "data-cell-value": stringValue,
    };
  };

  return {
    inspectCell: (value: string, event: MouseEvent) =>
      inspectCell(value, event, true),
    getCellProps,
  };
}
