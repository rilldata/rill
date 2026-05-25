// Parser borrowed from dotenv:
// https://github.com/motdotla/dotenv/blob/master/lib/main.js
const LINE_RE =
  /^\s*(?:export\s+)?([\w.-]+)(?:\s*=\s*?|:\s+?)(\s*'(?:\\'|[^'])*'|\s*"(?:\\"|[^"])*"|\s*`(?:\\`|[^`])*`|[^#\r\n]+)?\s*(?:#.*)?$/gm;

export function parseDotEnv(src: string): Record<string, string> {
  const obj: Record<string, string> = {};
  const normalized = src.replace(/\r\n?/g, "\n");

  LINE_RE.lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = LINE_RE.exec(normalized)) != null) {
    const key = match[1];
    let value = (match[2] ?? "").trim();
    const quote = value[0];
    value = value.replace(/^(['"`])([\s\S]*)\1$/, "$2");
    if (quote === '"') {
      value = value.replace(/\\n/g, "\n").replace(/\\r/g, "\r");
    }
    obj[key] = value;
  }
  return obj;
}

// Characters that change a bare value's meaning: whitespace (would be trimmed),
// `#` (starts a comment), or any quote char (would be read as the opening quote).
const NEEDS_QUOTING_RE = /[\s#"'`]/;

export function serializeDotEnv(entries: Record<string, string>): string {
  return Object.entries(entries)
    .map(([k, v]) => {
      if (v === "") return `${k}=`;
      if (!NEEDS_QUOTING_RE.test(v)) return `${k}=${v}`;
      // Prefer single quotes when possible: dotenv treats single-quoted
      // content literally, so we don't have to escape anything inside.
      // Newlines can't appear inside single quotes since the value must
      // stay on one line; fall through to double quotes in that case.
      if (!v.includes("'") && !/[\n\r]/.test(v)) return `${k}='${v}'`;
      // Double-quoted: the parser only unescapes \n and \r, so those are
      // the only sequences we escape. Backslashes are passed through
      // verbatim because the parser would not unescape them either.
      const escaped = v.replace(/\n/g, "\\n").replace(/\r/g, "\\r");
      return `${k}="${escaped}"`;
    })
    .join("\n");
}
