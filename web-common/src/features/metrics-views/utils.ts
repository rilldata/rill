export function getModelOutOfPossiblyMalformedYAML(yaml: string): string {
  const modelRegex = /\nmodel: (.*)\n/g;
  const modelMatch = modelRegex.exec(yaml);
  if (modelMatch) {
    return modelMatch[1];
  }
  return yaml;
}
