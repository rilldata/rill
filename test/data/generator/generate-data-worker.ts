import "../../../web-local/moduleAlias";
import { DataGeneratorWorker } from "./DataGeneratorWorker";
import workerpool from "workerpool";

const dataGeneratorWorker = new DataGeneratorWorker();

function generate(type: string, startId: number) {
  return new Promise((resolve) => {
    resolve(dataGeneratorWorker.generate(type, startId));
  });
}

workerpool.worker({
  generate,
});
