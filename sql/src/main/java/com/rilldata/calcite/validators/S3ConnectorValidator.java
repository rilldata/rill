package com.rilldata.calcite.validators;

import org.apache.calcite.tools.ValidationException;

import java.util.List;
import java.util.Map;
import java.util.Set;

public class S3ConnectorValidator
{
  public final static String PREFIX_PROP = "prefix";
  public final static String FORMAT_PROP = "format";
  public final static String ACCESS_PROP = "aws.access.key";
  public final static String SECRET_PROP = "aws.secret.key";

  final static List<String> requiredS3Props = List.of(PREFIX_PROP, FORMAT_PROP);
  final static List<String> optionalS3Props = List.of(ACCESS_PROP, SECRET_PROP);
  final static Set<String> supportedFormats = Set.of("csv", "parquet");

  public static void validate(Map<String, String> properties) throws ValidationException
  {
    for (String requiredProp : requiredS3Props) {
      if (!properties.containsKey(requiredProp) || properties.get(requiredProp) == null || properties.get(requiredProp)
          .isBlank()) {
        throw new ValidationException(
            String.format("Required property [%s] not present or blank for s3 connector", requiredProp));
      }
    }
    String format = properties.get(FORMAT_PROP).toLowerCase();
    if (!supportedFormats.contains(format)) {
      throw new ValidationException(
          String.format("Format [%s] not supported, supported formats are %s", format, supportedFormats));
    }
    for (String optionalProp : optionalS3Props) {
      if (properties.containsKey(optionalProp) && (properties.get(optionalProp) == null || properties.get(optionalProp)
          .isBlank())) {
        throw new ValidationException(
            String.format("No value specified for property [%s] for s3 connector", optionalProp));
      }
    }
  }
}
