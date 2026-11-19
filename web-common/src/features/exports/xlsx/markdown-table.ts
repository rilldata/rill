import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type { MarkdownTable } from "../markdown-content";
import { quoted } from "./number-format";
import type { ExportColumn, ExportTable } from "./payload";

// A table of a markdown text as a typed table for a data sheet, like the one behind a table component.
// `shown` holds the cells as the dashboard shows them and `raw` the same table before measure values were formatted
// (see the text payload in payload.ts), which has the precision the shown text rounds away.
//
// The header rows name the columns. A cell that shows a number becomes that number with a format that keeps its
// currency, percent sign, thousands separators and decimals; a magnitude ("1.2M") is written out in full, as in a
// table component's data sheet. A column counts as numeric when it shows numbers and nothing else that holds a digit
// (so "n/a" may sit between amounts, but codes and dates stay text). A format is per column where the column's
// numbers share one and else per row, as in a scorecard whose rows are different measures.
export function markdownExportTable(
  shown: MarkdownTable,
  raw?: MarkdownTable,
): ExportTable {
  const header = shown.rows.slice(0, shown.headerRows);
  const body = shown.rows.slice(shown.headerRows);
  // Formatting only replaces values, so the tables line up cell by cell; if they somehow don't, the shown text stands alone.
  const aligned =
    raw?.headerRows === shown.headerRows &&
    raw.rows.length === shown.rows.length &&
    raw.rows[0]?.length === shown.rows[0]?.length;
  const numbers = body.map((row, i) =>
    row.map((text, j) =>
      shownNumber(
        text,
        aligned ? raw.rows[shown.headerRows + i][j] : undefined,
      ),
    ),
  );

  const columns: ExportColumn[] = (shown.rows[0] ?? []).map((_, j) => {
    // A label that spans rows or repeats under a group is said once.
    const labels = header
      .map((row) => row[j])
      .filter((label, r, all) => label && label !== all[r - 1]);
    const isNumber =
      numbers.some((row) => row[j]) &&
      body.every((row, i) => numbers[i][j] || !/\d/.test(row[j]));
    return {
      name: `col_${j + 1}`,
      displayName: labels.join(" ") || `Column ${j + 1}`,
      type: isNumber ? WorkbookColumnType.NUMBER : WorkbookColumnType.STRING,
      numberFormat: isNumber
        ? commonFormat(numbers.flatMap((row) => row[j]?.look ?? []))
        : undefined,
    };
  });
  const isNumberCell = (i: number, j: number) =>
    columns[j].type === WorkbookColumnType.NUMBER && !!numbers[i][j];

  // A row's format covers its numbers in the columns that have no format of their own.
  const rowNumberFormats = body.map(
    (_, i) =>
      commonFormat(
        numbers[i].flatMap((number, j) =>
          isNumberCell(i, j) && columns[j].numberFormat === undefined
            ? [number!.look]
            : [],
        ),
      ) ?? "",
  );

  return {
    columns,
    rows: body.map((row, i) =>
      row.map((text, j) =>
        isNumberCell(i, j) ? numbers[i][j]!.value : text || null,
      ),
    ),
    ...(rowNumberFormats.some(Boolean) && { rowNumberFormats }),
  };
}

// How a number is shown, apart from its digits.
interface NumberLook {
  // Currency symbols before and after the number.
  prefix: string;
  suffix: string;
  percent: boolean;
  grouped: boolean;
  decimals: number;
}

interface ShownNumber {
  value: number;
  look: NumberLook;
}

// A number as a dashboard shows it: an optional sign and currency symbol, digits with optional thousands
// separators and decimals, then an optional magnitude, percent sign and currency symbol.
// A leading zero ("007") makes it a code rather than a number.
const SHOWN_NUMBER =
  /^([-−+])?(\p{Sc})?\s?([-−+])?([1-9]\d{0,2}(?:,\d{3})+|0|[1-9]\d*)(?:\.(\d+))?([kKMBT])?\s?(%)?\s?(\p{Sc})?$/u;

const MAGNITUDES: Record<string, number> = { k: 3, K: 3, M: 6, B: 9, T: 12 };

// A number as JavaScript prints it, which is how an unformatted measure value reads.
const PRINTED_NUMBER = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:e[+-]?\d+)?$/;

// Digits a cell can hold exactly; a longer run is an identifier.
const MAX_DIGITS = 15;

const PLAIN: NumberLook = {
  prefix: "",
  suffix: "",
  percent: false,
  grouped: false,
  decimals: 0,
};

// The number a cell shows, or undefined when it shows something else.
// `rawText` is the cell before measure values were formatted, when there is such a version of the text.
function shownNumber(
  text: string,
  rawText: string | undefined,
): ShownNumber | undefined {
  const raw =
    rawText !== undefined && rawText !== text && PRINTED_NUMBER.test(rawText)
      ? Number(rawText)
      : undefined;
  if (raw === undefined) return readNumber(text);

  // A formatted measure is the value behind it, whatever its text reads as: a locale that groups thousands
  // with "." shows 1234 as "1.234". The text says how to show the value when it reads as that value rounded,
  // either as written or with "." and "," swapped, as such a locale writes them.
  // A text that a cell format can't follow, such as an interval ("2.3 s"), leaves the value plain.
  const swapped = text.replace(/[.,]/g, (c) => (c === "." ? "," : "."));
  for (const reading of [text, swapped]) {
    const number = readNumber(reading);
    if (number && Math.abs(raw - number.value) <= number.step) {
      return { value: raw, look: number.look };
    }
  }
  return { value: raw, look: PLAIN };
}

// Reads a text as a number written with "," between thousands and "." before the decimals.
// `step` is the value of its last digit, which the text is rounded to.
function readNumber(
  text: string,
): (ShownNumber & { step: number }) | undefined {
  const match = SHOWN_NUMBER.exec(text);
  const [
    ,
    sign,
    prefix = "",
    innerSign,
    integer = "",
    fraction = "",
    magnitude,
    percent,
    suffix = "",
  ] = match ?? [];
  const digits = integer.replaceAll(",", "");
  if (
    !match ||
    (sign && innerSign) ||
    digits.length + fraction.length > MAX_DIGITS
  ) {
    return undefined;
  }

  const exponent = (MAGNITUDES[magnitude] ?? 0) - (percent ? 2 : 0);
  const negative = (sign ?? innerSign ?? "+") !== "+";
  return {
    // Scaled as text, so "83.3%" is 0.833 and not 83.3 / 100 with its binary rounding.
    value: Number(
      `${negative ? "-" : ""}${digits}.${fraction || "0"}e${exponent}`,
    ),
    step: 10 ** (exponent - fraction.length),
    look: {
      prefix,
      suffix,
      percent: !!percent,
      // A magnitude stands for digits the full number shows: grouped, with cents for an amount of money.
      grouped: integer.includes(",") || !!magnitude,
      decimals: magnitude ? (prefix || suffix ? 2 : 0) : fraction.length,
    },
  };
}

// The Excel number format for numbers that are shown alike: with the same currency or percent sign,
// grouped if any is, to the most decimals any shows. Undefined when they differ in unit, or there are none.
function commonFormat(looks: NumberLook[]): string | undefined {
  const [first] = looks;
  if (
    !first ||
    looks.some(
      (look) =>
        look.prefix !== first.prefix ||
        look.suffix !== first.suffix ||
        look.percent !== first.percent,
    )
  ) {
    return undefined;
  }
  const grouped = looks.some((look) => look.grouped);
  const decimals = Math.max(...looks.map((look) => look.decimals));
  const digits =
    (grouped ? "#,##0" : "0") + (decimals ? "." + "0".repeat(decimals) : "");
  if (first.percent) return `${digits}%`;
  // Plain numbers are best left to Excel, which shows as many decimals as each has.
  if (!first.prefix && !first.suffix && !grouped) return "General";
  return `${quoted(first.prefix)}${digits}${quoted(first.suffix)}`;
}
