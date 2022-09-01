package com.rilldata.calcite;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;

public class StaticSchemaFactory
{
  public static StaticSchema create(String json) throws IOException
  {
    return new StaticSchema(new ObjectMapper().readValue(json, JsonSchema.class));
  }
}
