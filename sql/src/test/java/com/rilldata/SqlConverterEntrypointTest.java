package com.rilldata;

import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.protobuf.generated.Requests;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

public class SqlConverterEntrypointTest
{
  @Test
  public void testDialectSanity() {
    Assertions.assertNotEquals(null, Dialects.valueOf("DUCKDB"));
    Assertions.assertNotEquals(null, Requests.Dialect.valueOf("DUCKDB"));
  }

  @Test
  public void testTranspileSanity() {
    Requests.Request request = Requests.Request
        .newBuilder()
        .setTranspileRequest(Requests.TranspileRequest
                                 .newBuilder()
                                 .setSql("select 1")
                                 .setDialect(Requests.Dialect.DUCKDB)
                                 .setCatalog("""
                                 { 
                                  "artifacts": [],
                                  "schemas": []
                                 }
                                 """)
                                 .build()

        )
        .build();
    Requests.Response response = SqlConverterEntrypoint.transpile(request);
    Assertions.assertEquals("SELECT 1", response.getTranspileResponse().getSql());
  }
}
