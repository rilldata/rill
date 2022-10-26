package com.rilldata;

import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.protobuf.generated.Requests;
import com.rilldata.protobuf.generated.SqlNodeProto;
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

  @Test
  public void testTranspileWithCatalog() {
    Requests.Request request = Requests.Request
        .newBuilder()
        .setTranspileRequest(Requests.TranspileRequest
                                 .newBuilder()
                                 .setSql("select \"name\" from \"earth\".\"heroes\" h where h.\"power\" = 'speed'")
                                 .setDialect(Requests.Dialect.DUCKDB)
                                 .setCatalog("""
                                    {
                                     "schemas": [
                                       {
                                         "name": "earth",
                                         "tables": [
                                           {
                                             "name": "heroes",
                                             "columns": [
                                               {
                                                 "name": "name",
                                                 "type": "varchar"
                                               },
                                               {
                                                 "name": "power",
                                                 "type": "varchar"
                                               }
                                                            \s
                                             ]
                                           }
                                         ]
                                       }
                                     ],
                                     "artifacts": []
                                    }
                                 """)
                                 .build()

        )
        .build();
    Requests.Response response = SqlConverterEntrypoint.transpile(request);
    Assertions.assertEquals(
        "SELECT \"H\".\"name\"\nFROM \"earth\".\"heroes\" AS \"H\"\nWHERE \"H\".\"power\" = 'speed'",
        response.getTranspileResponse().getSql()
    );
  }

  @Test
  public void testParseSanity() {
    Requests.Request request = Requests.Request
        .newBuilder()
        .setParseRequest(Requests.ParseRequest
                                 .newBuilder()
                                 .setSql("select \"name\" from \"earth\".\"heroes\" where \"power\" = 'speed'")
                                 .setAddTypeInfo(true)
                                 .setCatalog("""
                                     {
                                      "schemas": [
                                        {
                                          "name": "earth",
                                          "tables": [
                                            {
                                              "name": "heroes",
                                              "columns": [
                                                {
                                                  "name": "name",
                                                  "type": "varchar"
                                                },
                                                {
                                                  "name": "power",
                                                  "type": "varchar"
                                                }
                                                              
                                              ]
                                            }
                                          ]
                                        }
                                      ],
                                      "artifacts": []
                                     }
                                 """)
                                 .build()

        )
        .build();
    Requests.Response response = SqlConverterEntrypoint.parse(request); System.out.println(response.getError().getStackTrace());
    SqlNodeProto ast = response.getParseResponse().getAst();
    Assertions.assertNotNull(ast.getSqlSelectProto().getSelectList().getTypeInformation());
  }
}
