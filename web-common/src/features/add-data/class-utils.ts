import {
  type AddDataState,
  AddDataStep,
  type AddDataStepWithSchema,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";

const AddDataClassByStepMap: Partial<Record<AddDataStep, string>> = {
  [AddDataStep.SelectConnector]: "h-fit md:w-[900px] w-[550px]",
  [AddDataStep.Import]: "h-fit w-[550px]",
};
const AddDataClassBySchemaMap: Partial<Record<string, string>> = {
  local_file: "h-[300px] my-auto w-[550px]",
};
const DefaultAddDataClass = "h-[630px] md:w-[900px] w-[550px]";

export function getAddDataClass(addDataState: AddDataState) {
  const schema = (addDataState as AddDataStepWithSchema).schema ?? undefined;
  if (schema && schema in AddDataClassBySchemaMap)
    return AddDataClassBySchemaMap[schema];
  return AddDataClassByStepMap[addDataState.step] ?? DefaultAddDataClass;
}

const FormClassBySchemaMap: Partial<Record<string, string>> = {
  local_file: "px-6 my-auto h-fit",
};
const DefaultFormClass = "p-6 flex-grow overflow-auto";

export function getFormClass(addDataState: AddDataState) {
  const schema = (addDataState as AddDataStepWithSchema).schema ?? undefined;
  if (schema && schema in FormClassBySchemaMap)
    return FormClassBySchemaMap[schema];
  return DefaultFormClass;
}
