import {
  connectors,
  toConnectorDriver,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";

export function load({ params }) {
  const connectorName = params.name;

  const connectorInfo = connectors.find((c) => c.name === connectorName);
  const connectorDriver = connectorInfo
    ? toConnectorDriver(connectorInfo)
    : null;
  console.log(connectorDriver, connectorInfo);

  return {
    connectorName,
    connectorInfo,
    connectorDriver,
  };
}
