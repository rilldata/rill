import { parseDocument } from "yaml";
import { metricsTemplate } from "./template";

type MeasuresOrDimensions = "measures" | "dimensions";

export function getEmptyDashboardDocument() {
  const templateDocument = parseDocument(metricsTemplate);

  return templateDocument.toString();
}

export function getNode(metricsYamlString, nodeKey) {
  const document = parseDocument(metricsYamlString);
  return document.get(nodeKey);
}

// export function getNodeAndReduce(metricsYamlString, nodeKey) {
//   const document = parseDocument(metricsYamlString);
//   return document.get(nodeKey);
// }

export function deleteMetricOrDimension(
  metricsYamlString: string,
  nodeType: MeasuresOrDimensions,
  index: number
) {
  const document = parseDocument(metricsYamlString);

  const node = document.get(nodeType) as Array<MentionType>;
  node.splice(index, 1);

  return document.toString();
}

export function editMetricOrDimension(
  metricsYamlString: string,
  nodeType: MeasuresOrDimensions,
  index: number,
  name,
  value
) {
  const document = parseDocument(metricsYamlString);

  const node = document.get(nodeType) as Array<MentionType>; // fix types
  node[index].forEach((e) => {
    if (name in e) {
      node[index][name] = value;
    }
  });

  return document.toString();
}

export function addNewMeasure(metricsYamlString: string) {
  const document = parseDocument(metricsYamlString);
  document.addIn(
    ["measures"],
    `[label: "",
    expression: "",
    description: "",
    format_preset: "", 
    visible: false ]`
  );

  return document.toString();
}

export function addNewDimension(metricsYamlString: string) {
  const document = parseDocument(metricsYamlString);
  document.addIn(
    ["dimension"],
    `[label: "",
      property: "",
      description: "",
      visible: false ]`
  );

  return document.toString();
}
