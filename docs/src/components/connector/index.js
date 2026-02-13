/**
 * Connector documentation components
 *
 * These components provide reusable sections for connector documentation pages,
 * reducing boilerplate and ensuring consistency across all connector docs.
 *
 * IMPORTANT: DevProdSeparation and DeployToCloud do NOT include headings.
 * Add the markdown heading before the component so it appears in the ToC.
 *
 * Usage in MDX:
 * ```mdx
 * import { TwoStepFlowIntro, EnvPullTip, DevProdSeparation, DeployToCloud } from '@site/src/components/connector';
 *
 * ## Using the Add Data UI
 *
 * <TwoStepFlowIntro connector="Athena" />
 *
 * <EnvPullTip />
 *
 * ## Separating Dev and Prod Environments
 *
 * <DevProdSeparation />
 *
 * ## Deploy to Rill Cloud
 *
 * <DeployToCloud
 *   connector="Athena"
 *   connectorId="athena"
 *   credentialDescription="an access key and secret for an AWS service account"
 * />
 * ```
 */

export { default as TwoStepFlowIntro } from './TwoStepFlowIntro';
export { default as EnvPullTip } from './EnvPullTip';
export { default as DevProdSeparation } from './DevProdSeparation';
export { default as DeployToCloud } from './DeployToCloud';
