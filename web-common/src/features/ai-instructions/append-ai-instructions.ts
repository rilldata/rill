import { Scalar, isMap, isSeq, parse, parseAllDocuments, parseDocument } from "yaml";

// appendAIInstructions appends a rule to the file's top-level ai_instructions,
// creating the property if absent. It edits the YAML AST so comments and key
// order are preserved. Used to apply admin-tuned AI-context suggestions
// (feedback suggest-fix, eval suggest-fix) to rill.yaml or a metrics view file.
export function appendAIInstructions(content: string, rule: string): string {
  // Multi-document files are rejected: re-encoding would silently drop
  // everything after the first document.
  if (content.trim() !== "" && parseAllDocuments(content).length > 1) {
    throw new Error("Multi-document YAML files are not supported.");
  }

  const doc = parseDocument(content);
  if (doc.errors.length > 0) {
    throw new Error(`Failed to parse the file as YAML: ${doc.errors[0].message}`);
  }
  if (doc.contents !== null && !isMap(doc.contents)) {
    throw new Error("The file's top level is not a YAML mapping.");
  }

  const existing = doc.get("ai_instructions");
  const scalar = new Scalar(
    typeof existing === "string" && existing.trim() !== ""
      ? `${existing.replace(/\n+$/, "")}\n\n${rule}\n`
      : `${rule}\n`,
  );
  scalar.type = Scalar.BLOCK_LITERAL;
  doc.set("ai_instructions", scalar);
  return doc.toString();
}

// appendYAMLMeasure appends a measure definition (a YAML mapping with name, expression, etc.)
// to a metrics view file's measures list, creating the list if absent. Like appendAIInstructions,
// it edits the YAML AST so the file's comments and key order are preserved.
export function appendYAMLMeasure(content: string, measureYAML: string): string {
  if (content.trim() !== "" && parseAllDocuments(content).length > 1) {
    throw new Error("Multi-document YAML files are not supported.");
  }

  const measure: unknown = parse(measureYAML);
  if (measure === null || typeof measure !== "object" || Array.isArray(measure)) {
    throw new Error("The measure definition must be a YAML mapping.");
  }

  const doc = parseDocument(content);
  if (doc.errors.length > 0) {
    throw new Error(`Failed to parse the file as YAML: ${doc.errors[0].message}`);
  }
  if (doc.contents !== null && !isMap(doc.contents)) {
    throw new Error("The file's top level is not a YAML mapping.");
  }

  const measures = doc.get("measures", true);
  if (measures === undefined) {
    doc.set("measures", doc.createNode([measure]));
  } else if (isSeq(measures)) {
    measures.add(doc.createNode(measure));
  } else {
    throw new Error("The file's measures property is not a list.");
  }
  return doc.toString();
}
