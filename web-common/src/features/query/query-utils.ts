export interface TextRange {
  from: number;
  to: number;
}

export interface SqlExecution extends TextRange {
  sql: string;
  startLine: number;
  endLine: number;
}

/**
 * Finds the semicolon-delimited SQL statement containing the cursor.
 * Semicolons inside comments and quoted values are ignored. This only chooses
 * text to execute; it does not validate or classify the SQL.
 */
export function getStatementRange(
  sql: string,
  cursor: number,
): TextRange | null {
  const ranges: TextRange[] = [];
  let start = 0;
  let index = 0;
  let state:
    | "normal"
    | "single-quote"
    | "double-quote"
    | "backtick"
    | "line-comment"
    | "block-comment" = "normal";
  let dollarQuote = "";

  while (index < sql.length) {
    const character = sql[index];
    const nextCharacter = sql[index + 1];

    if (dollarQuote) {
      if (sql.startsWith(dollarQuote, index)) {
        index += dollarQuote.length;
        dollarQuote = "";
      } else {
        index += 1;
      }
      continue;
    }

    if (state === "line-comment") {
      if (character === "\n") state = "normal";
      index += 1;
      continue;
    }

    if (state === "block-comment") {
      if (character === "*" && nextCharacter === "/") {
        state = "normal";
        index += 2;
      } else {
        index += 1;
      }
      continue;
    }

    if (state === "single-quote") {
      if (character === "'" && nextCharacter === "'") {
        index += 2;
      } else {
        if (character === "'") state = "normal";
        index += 1;
      }
      continue;
    }

    if (state === "double-quote") {
      if (character === '"' && nextCharacter === '"') {
        index += 2;
      } else {
        if (character === '"') state = "normal";
        index += 1;
      }
      continue;
    }

    if (state === "backtick") {
      if (character === "`" && nextCharacter === "`") {
        index += 2;
      } else {
        if (character === "`") state = "normal";
        index += 1;
      }
      continue;
    }

    if (character === "-" && nextCharacter === "-") {
      state = "line-comment";
      index += 2;
      continue;
    }
    if (character === "/" && nextCharacter === "*") {
      state = "block-comment";
      index += 2;
      continue;
    }
    if (character === "'") {
      state = "single-quote";
      index += 1;
      continue;
    }
    if (character === '"') {
      state = "double-quote";
      index += 1;
      continue;
    }
    if (character === "`") {
      state = "backtick";
      index += 1;
      continue;
    }
    if (character === "$") {
      const match = sql.slice(index).match(/^\$[A-Za-z_][A-Za-z0-9_]*\$|^\$\$/);
      if (match) {
        dollarQuote = match[0];
        index += dollarQuote.length;
        continue;
      }
    }
    if (character === ";") {
      ranges.push({ from: start, to: index + 1 });
      start = index + 1;
    }
    index += 1;
  }

  ranges.push({ from: start, to: sql.length });

  const clampedCursor = Math.max(0, Math.min(cursor, sql.length));
  let selectedIndex = ranges.findIndex(
    (range) =>
      clampedCursor >= range.from &&
      (clampedCursor < range.to ||
        (clampedCursor === sql.length && range.to === sql.length)),
  );
  if (selectedIndex === -1) selectedIndex = ranges.length - 1;

  const selected = trimRange(sql, ranges[selectedIndex]);
  if (selected) return selected;

  for (let offset = 1; offset < ranges.length; offset += 1) {
    const previous = ranges[selectedIndex - offset];
    if (previous) {
      const trimmed = trimRange(sql, previous);
      if (trimmed) return trimmed;
    }

    const next = ranges[selectedIndex + offset];
    if (next) {
      const trimmed = trimRange(sql, next);
      if (trimmed) return trimmed;
    }
  }

  return null;
}

export function trimRange(sql: string, range: TextRange): TextRange | null {
  let { from, to } = range;
  while (from < to && /\s/.test(sql[from])) from += 1;
  while (to > from && /\s/.test(sql[to - 1])) to -= 1;
  return from === to ? null : { from, to };
}

export function formatDataType(code: string | undefined): string {
  if (!code) return "UNKNOWN";
  const normalized = code.replace(/^CODE_/, "");
  return normalized.startsWith("UNKNOWN(") ? "UNKNOWN" : normalized;
}

export function formatExecutionTime(milliseconds: number): string {
  return milliseconds < 1000
    ? `${milliseconds}ms`
    : `${(milliseconds / 1000).toFixed(1)}s`;
}
