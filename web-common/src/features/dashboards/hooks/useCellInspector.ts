import { cellInspectorStore } from "../stores/cellInspectorStore";

export function useCellInspector() {
  const inspectCell = (value: string, persist = false) => {
    if (!value) return;

    const stringValue = String(value);

    if (persist) {
      // For double-click, we still want to open the inspector directly
      cellInspectorStore.open(stringValue);
    } else {
      // Just update the value in the store, visibility is controlled at the dashboard level
      cellInspectorStore.updateValue(stringValue);
    }
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

        // Just update the value in the store, visibility is controlled at the dashboard level
        cellInspectorStore.updateValue(stringValue);
      },
      onDblClick: (e: MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        inspectCell(stringValue, true);
      },
      // Add tabindex for keyboard accessibility
      tabIndex: 0,
      role: "cell",
    };
  };

  return {
    inspectCell: (value: string) => inspectCell(value, true),
    getCellProps,
  };
}
