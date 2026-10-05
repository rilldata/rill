<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import EvaluationsEditor from "@rilldata/web-common/features/evaluations/EvaluationsEditor.svelte";
  import { parseDocument } from "yaml";
  import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
  import {
    queryServiceMetricsViewEvaluate,
    type V1EvaluateQuestion,
    type V1EvaluateRequest,
    type V1MetricsViewAggregationRequest,
  } from "@rilldata/web-common/runtime-client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    EvaluateChoiceQuestion,
    EvaluateNoulQuestion,
    EvaluateNoulQuestion_Criteria,
    EvaluateQuestion,
    EvaluateScoreQuestion,
  } from "@rilldata/web-common/proto/gen/rill/ai/v1/ai_pb.ts";
  import { Value, type JsonValue } from "@bufbuild/protobuf";

  let {
    open = $bindable(false),
    aggregationRequest,
  }: {
    open: boolean;
    aggregationRequest: V1MetricsViewAggregationRequest;
  } = $props();

  const runtimeClient = useRuntimeClient();

  const INIT_DOC = `state:
  data: {{ .data }}
  meta: {{ .meta }}
  prompt: "What was the impact of this dimension?"
questions:
  high_impressions:
    type: "noul"
    instructions: "Was this rendered a lot?"
    criteria:
      true: "Has high impressions"
      false: "Has low impressions"
  high_spend:
    type: "noul"
    instructions: "Did this spend a lot?"
    criteria:
      true: "Has high spend"
      false: "Has low spend"

`;
  let contentsStore = $derived(
    SvelteLocalStorage.createStringStore("eval", INIT_DOC),
  );
  let parsedDocument = $derived(parseDocument(contentsStore.value));

  async function evaluate() {
    const rawRequest = parsedDocument.toJSON();

    const request = {
      state: rawRequest.state,
      questions: Object.fromEntries(
        Object.entries(rawRequest.questions)
          .map(([l, q]) => [l, mapRawQuestion(q as any)])
          .filter(([, q]) => q !== undefined),
      ),
    } satisfies V1EvaluateRequest;
    console.log(request);

    const evaluationResp = await queryServiceMetricsViewEvaluate(
      runtimeClient,
      {
        metricsViewName: aggregationRequest.metricsView,
        query: {
          ...aggregationRequest,
          limit: "10",
        },
        request,
      },
    );
    console.log(evaluationResp);
  }

  function mapRawQuestion(
    rawQuestion: Record<string, any>,
  ): V1EvaluateQuestion | undefined {
    const instructions = toValue(rawQuestion.instructions);
    const criteria = rawQuestion.criteria;

    let question: EvaluateQuestion["question"];
    switch (rawQuestion.type) {
      case "noul":
        question = {
          case: "noul",
          value: new EvaluateNoulQuestion({
            instructions,
            criteria: criteria
              ? new EvaluateNoulQuestion_Criteria({
                  true: toValue(criteria.true),
                  false: toValue(criteria.false),
                })
              : undefined,
          }),
        };
        break;
      case "choice":
        question = {
          case: "choice",
          value: new EvaluateChoiceQuestion({
            instructions,
            criteria: Object.fromEntries(
              Object.entries(criteria ?? {}).map(([k, v]) => [
                k,
                Value.fromJson(v as JsonValue),
              ]),
            ),
          }),
        };
        break;
      case "score":
        question = {
          case: "score",
          value: new EvaluateScoreQuestion({
            instructions,
            criteria: ((criteria ?? []) as JsonValue[]).map((v) =>
              Value.fromJson(v),
            ),
          }),
        };
        break;
      default:
        return undefined;
    }

    return new EvaluateQuestion({ question }).toJson() as V1EvaluateQuestion;
  }

  function toValue(raw: unknown): Value | undefined {
    if (raw === undefined) return undefined;
    return Value.fromJson(raw as JsonValue);
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="h-[600px]">
    <Dialog.Title>Typesafe evaluation</Dialog.Title>

    <EvaluationsEditor
      contents={contentsStore.value}
      onChange={contentsStore.setter}
    />

    <Dialog.Footer>
      <Button onClick={evaluate}>Eval</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
