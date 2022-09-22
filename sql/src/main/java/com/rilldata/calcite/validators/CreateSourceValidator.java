package com.rilldata.calcite.validators;

import com.rilldata.calcite.models.SqlCreateSource;
import org.apache.calcite.tools.ValidationException;

import java.util.List;
import java.util.Map;

public class CreateSourceValidator
{
  public final static String CONNECTOR_PROP = "connector";
  final static List<String> requiredProps = List.of(CONNECTOR_PROP);

  public static void validateConnector(SqlCreateSource sqlCreateSource) throws ValidationException
  {
    Map<String, String> properties = sqlCreateSource.properties;
    String sourceName = sqlCreateSource.name.getSimple();
    for (String requiredProp : requiredProps) {
      if (!properties.containsKey(requiredProp)) {
        throw new ValidationException(
            String.format("Required property [%s] not found for source [%s]", requiredProp, sourceName));
      }
    }
    String connector = properties.get(CONNECTOR_PROP);
    switch (connector) {
    case "s3":
      S3ConnectorValidator.validate(properties);
      break;
    default:
      throw new ValidationException(
          String.format("No connector of type [%s] found for source [%s]", connector, sourceName));
    }
  }
}
