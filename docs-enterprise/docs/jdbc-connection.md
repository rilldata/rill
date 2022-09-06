---
title: "JDBC Connection"
slug: "jdbc-connection"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

## Overview
You can use a JDBC connection to query your Druid service. 

## Credentials
To authenticate via a JDBC connection, you will need to use either an [API Password](/api-password)  or a [Service Account](/service-accounts). If using an API password, when you connect you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

## Setup
1. Download the [Avatica JDBC driver](https://calcite.apache.org/avatica/downloads/), version 1.17.0 or later. Note that as of the time of this writing, Avatica 1.17.0, the latest version, does not support passing connection string parameters from the URL to Druid, so you must pass them using a Properties object.
2. Add the Avatica client jar to your class path.
3. Copy the JDBC Connection URL from the Integrations page in RCC. In the example below, this url has been used in the database_url variable.
![](https://images.contentful.com/ve6smfzbifwz/3O2fyAFUOMfeEIBeUX7JGL/843b359d27ae75a3229e139ee7825c92/a684bdb-Screen_Shot_2021-07-01_at_11.12.25_AM.png)
4. To authenticate, in your code you will supply either your username and API password or service account and service account password, as described above.


## Example Program (For Avatica 1.17.0 or later)

 ```java title="Java"
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.Properties;

public class AvaticaConn
{
  public static void main(String[] args) throws SQLException
  {
    String datababase_url = "https://druid.ws1.public.rilldata.com/druid/v2/sql/avatica"

    String connectUrl = String.format(
        "jdbc:avatica:remote:url=",
        database_url
    );

    Properties connectionProperties = new Properties();
    connectionProperties.setProperty("<user>", user);
    connectionProperties.setProperty("<password>", pwd);

    String query = "SELECT 1337";
    try (Connection connection = DriverManager.getConnection(connectUrl, connectionProperties)) {
      try (
          final Statement statement = connection.createStatement();
          final ResultSet resultSet = statement.executeQuery(query)
      ) {
        while (resultSet.next()) {
          System.out.println(resultSet.getString(1)); // Do something
        }
      }
    }
```

## Example Program (For Avatica versions before 1.17.0)

 ```java title="Java"
"code": "import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.Properties;

public class AvaticaConn
{
  public static void main(String[] args) throws SQLException
  {
    // path to jdk trust store (change as per OS)
    String trustStore = "/Library/Java/JavaVirtualMachines/jdk1.8.0_211.jdk/Contents/Home/jre/lib/security/cacerts";
    String trustStorePwd = "changeit"; // default java trust store pwd
    String url = "https://<druid_url>/druid/v2/sql/avatica/";

    String connectUrl = String.format(
        "jdbc:avatica:remote:url=%s;truststore=%s;truststore_password=%s",
        url,
        trustStore,
        trustStorePwd
    );

    Properties connectionProperties = new Properties();
    connectionProperties.setProperty("<user>", user);
    connectionProperties.setProperty("<password>", pwd);

    String query = "SELECT 1337";
    try (Connection connection = DriverManager.getConnection(connectUrl, connectionProperties)) {
      try (
          final Statement statement = connection.createStatement();
          final ResultSet resultSet = statement.executeQuery(query)
      ) {
        while (resultSet.next()) {
          System.out.println(resultSet.getString(1)); // Do something
        }
      }
    }",
```

5. Additional information can be found in the Apache Druid documentation, [here](https://druid.apache.org/docs/latest/querying/sql.html#jdbc)