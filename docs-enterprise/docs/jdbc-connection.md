---
title: "JDBC Connection"
slug: "jdbc-connection"
hidden: false
createdAt: "2020-10-31T00:38:02.158Z"
updatedAt: "2021-08-03T01:08:16.229Z"
---
#Overview
You can use a JDBC connection to query your Druid service. 

#Credentials
To authenticate via a JDBC connection, you will need to use either an [API Password](doc:api-password)  or a [Service Account](doc:service-accounts). If using an API password, when you connect you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

# Setup
1. Download the [Avatica JDBC driver](https://calcite.apache.org/avatica/downloads/), version 1.17.0 or later. Note that as of the time of this writing, Avatica 1.17.0, the latest version, does not support passing connection string parameters from the URL to Druid, so you must pass them using a Properties object.

1. Add the Avatica client jar to your class path.

2. Copy the JDBC Connection URL from the Integrations page in RCC. In the example below, this url has been used in the database_url variable.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/a684bdb-Screen_Shot_2021-07-01_at_11.12.25_AM.png",
        "Screen Shot 2021-07-01 at 11.12.25 AM.png",
        1354,
        214,
        "#f7f7f8"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
3. To authenticate, in your code you will supply either your username and API password or service account and service account password, as described above.

# Example Program (For Avatica 1.17.0 or later)

```java
"code": "import java.sql.Connection;\nimport java.sql.DriverManager;\nimport java.sql.ResultSet;\nimport java.sql.SQLException;\nimport java.sql.Statement;\nimport java.util.Properties;\n\npublic class AvaticaConn\n{\n  public static void main(String[] args) throws SQLException\n  {\n    String datababase_url = \"https://druid.ws1.public.rilldata.com/druid/v2/sql/avatica\"\n\n    String connectUrl = String.format(\n        \"jdbc:avatica:remote:url=\",\n        database_url\n    );\n\n    Properties connectionProperties = new Properties();\n    connectionProperties.setProperty(\"<user>\", user);\n    connectionProperties.setProperty(\"<password>\", pwd);\n\n    String query = \"SELECT 1337\";\n    try (Connection connection = DriverManager.getConnection(connectUrl, connectionProperties)) {\n      try (\n          final Statement statement = connection.createStatement();\n          final ResultSet resultSet = statement.executeQuery(query)\n      ) {\n        while (resultSet.next()) {\n          System.out.println(resultSet.getString(1)); // Do something\n        }\n      }\n    }",
```

## Example Program (For Avatica versions before 1.17.0)

```java
"code": "import java.sql.Connection;\nimport java.sql.DriverManager;\nimport java.sql.ResultSet;\nimport java.sql.SQLException;\nimport java.sql.Statement;\nimport java.util.Properties;\n\npublic class AvaticaConn\n{\n  public static void main(String[] args) throws SQLException\n  {\n    // path to jdk trust store (change as per OS)\n    String trustStore = \"/Library/Java/JavaVirtualMachines/jdk1.8.0_211.jdk/Contents/Home/jre/lib/security/cacerts\";\n    String trustStorePwd = \"changeit\"; // default java trust store pwd\n    String url = \"https://<druid_url>/druid/v2/sql/avatica/\";\n\n    String connectUrl = String.format(\n        \"jdbc:avatica:remote:url=%s;truststore=%s;truststore_password=%s\",\n        url,\n        trustStore,\n        trustStorePwd\n    );\n\n    Properties connectionProperties = new Properties();\n    connectionProperties.setProperty(\"<user>\", user);\n    connectionProperties.setProperty(\"<password>\", pwd);\n\n    String query = \"SELECT 1337\";\n    try (Connection connection = DriverManager.getConnection(connectUrl, connectionProperties)) {\n      try (\n          final Statement statement = connection.createStatement();\n          final ResultSet resultSet = statement.executeQuery(query)\n      ) {\n        while (resultSet.next()) {\n          System.out.println(resultSet.getString(1)); // Do something\n        }\n      }\n    }",
```

5. Additional information can be found in the Apache Druid documentation, [here](https://druid.apache.org/docs/latest/querying/sql.html#jdbc)