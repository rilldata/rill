/**
 * These components apply styling to a single <span> for consistency.
 * Additional Args
 * @prop {boolean} isNull whether the value is null.
 * @prop {boolean} inTable whether this value is inline or should be a table cell.
 */
import DataTypeIcon from "./DataTypeIcon.svelte";
import FormattedDataType from "./FormattedDataType.svelte";
import Number from "./Number.svelte";
import Timestamp from "./Timestamp.svelte";
import Varchar from "./Varchar.svelte";

export { Number, Timestamp, Varchar, FormattedDataType, DataTypeIcon };
