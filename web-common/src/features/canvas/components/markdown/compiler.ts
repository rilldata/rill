/**
 * Compiler to convert custom syntax to Go template syntax
 */

import type { ParsedSyntax, MeasureMention, RepeatBlock } from "./syntax-parser";

export interface CompilationOptions {
  defaultMetricsView?: string;
}

/**
 * Compile @measurename to Go template: {{ metrics_sql("SELECT measurename FROM metrics_view LIMIT 1") }}
 */
function compileMeasureMention(
  mention: MeasureMention,
  options: CompilationOptions = {},
): string {
  const metricsView = mention.metricsViewName || options.defaultMetricsView || "metrics_view";
  const measureName = mention.measureName;

  // Escape the measure name for SQL
  const escapedMeasure = measureName.replace(/"/g, '""');

  // Generate metrics SQL query
  const sql = `SELECT ${escapedMeasure} FROM ${metricsView} LIMIT 1`;

  // Return Go template syntax
  return `{{ metrics_sql("${sql.replace(/"/g, '\\"')}") }}`;
}

/**
 * Compile [repeat ...] block to Go template with range
 */
function compileRepeatBlock(
  block: RepeatBlock,
  options: CompilationOptions = {},
): string {
  const metricsView = options.defaultMetricsView || "metrics_view";
  const measure = block.measure.replace(/"/g, '""');
  const dimension = block.dimension.replace(/"/g, '""');

  // Build SELECT clause
  let selectClause = `SELECT ${measure}, ${dimension}`;

  // Build WHERE clause
  let whereClause = "";
  if (block.where) {
    // Escape the WHERE clause
    const escapedWhere = block.where.replace(/"/g, '""');
    whereClause = ` WHERE ${escapedWhere}`;
  }

  // Build ORDER BY clause
  let orderByClause = "";
  if (block.orderBy) {
    // Parse ORDER BY - could be "column DESC" or just "column"
    const orderByParts = block.orderBy.trim().split(/\s+/);
    const column = orderByParts[0].replace(/"/g, '""');
    const direction = orderByParts[1]?.toUpperCase() === "DESC" ? " DESC" : "";
    orderByClause = ` ORDER BY ${column}${direction}`;
  }

  // Build LIMIT clause
  let limitClause = "";
  if (block.limit) {
    limitClause = ` LIMIT ${block.limit}`;
  }

  // Construct the full SQL query
  const sql = `${selectClause} FROM ${metricsView}${whereClause}${orderByClause}${limitClause}`;

  // Escape for Go template string
  const escapedSql = sql.replace(/"/g, '\\"');

  // Compile the inner content (may contain nested mentions)
  const compiledContent = compileContent(block.content, options);

  // Return Go template with range
  return `{{ range metrics_sql_rows("${escapedSql}") }}${compiledContent}{{ end }}`;
}

import { parseSyntax, type ParsedSyntax, type MeasureMention, type RepeatBlock } from "./syntax-parser";

/**
 * Compile content, replacing all custom syntax with Go templates
 */
export function compileContent(
  content: string,
  options: CompilationOptions = {},
): string {
  // First, parse the syntax
  const parsed = parseSyntax(content);

  if (parsed.length === 0) {
    return content;
  }

  // Build result by replacing syntax in reverse order (to preserve indices)
  let result = content;
  for (let i = parsed.length - 1; i >= 0; i--) {
    const syntax = parsed[i];
    let replacement: string;

    if (syntax.type === "measure") {
      replacement = compileMeasureMention(syntax, options);
    } else if (syntax.type === "repeat") {
      replacement = compileRepeatBlock(syntax, options);
    } else {
      continue;
    }

    // Replace the syntax with compiled version
    result =
      result.slice(0, syntax.startIndex) +
      replacement +
      result.slice(syntax.endIndex);
  }

  return result;
}

/**
 * Main compilation function
 */
export function compileToGoTemplate(
  html: string,
  options: CompilationOptions = {},
): string {
  return compileContent(html, options);
}

