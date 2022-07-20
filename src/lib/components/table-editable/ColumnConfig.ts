import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

import type { SvelteComponent } from "svelte";

import TableCellInput from "$lib/components/table-editable/TableCellInput.svelte";
import TableCellSelector from "../table-editable/TableCellSelector.svelte";
import TableCellSparkline from "$lib/components/metrics-definition/TableCellSparkline.svelte";

export type CellRendererComponent = new (
  // FIXME: these types are the   ones taken by the components
  // columnConfig: ColumnConfig,
  // index: number,
  // row: EntityRecord
  ...args: any[]
) => SvelteComponent;

export type SelectorOption = {
  label: string;
  value: string;
};

export type TableEventHandler = (
  rowIndex: number,
  columnName: string,
  value: string
) => void;

export interface CellConfig {
  component: CellRendererComponent;
}

export class CellSparkline implements CellConfig {
  component = TableCellSparkline;
}

export interface InputValidation {
  state: ValidationState;
  message: string;
}
export class CellConfigInput implements CellConfig {
  component = TableCellInput;
  constructor(
    public onchange: TableEventHandler,
    public getInputValidation?: (
      row: EntityRecord,
      value: unknown
    ) => InputValidation,
    /**
     * This is called per every keystroke.
     */
    public onKeystroke?: TableEventHandler
  ) {}
}

export class CellConfigSelector implements CellConfig {
  component = TableCellSelector;
  constructor(
    public onchange: TableEventHandler,
    public options: SelectorOption[],
    public placeholderLabel?: string
  ) {}
}

/**
 * config info for table columns
 *
 * name: the property name used in an EntityRecord
 * label?: label used for display in table header (`name` is used if not provided)
 * headerTooltip: tooltip when hovering over column header
 * cellRenderer: svelte component and other options used to render cell content (event handlers, etc.)
 */
export interface ColumnConfig<T extends CellConfig> {
  //FIXME: specify types based on CellRendererComponent
  name: string;
  label?: string;
  headerTooltip?: string;
  cellRenderer: T;
}
