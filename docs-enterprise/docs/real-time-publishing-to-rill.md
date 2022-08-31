---
title: "Connect Real-time Publishing to Rill"
slug: "real-time-publishing-to-rill"
hidden: true
createdAt: "2021-10-06T14:10:41.272Z"
updatedAt: "2021-10-08T19:59:35.106Z"
---
[block:api-header]
{
  "title": "Streaming Platform"
}
[/block]
Connectivity is done with standard Apache Kafka client configuration allowing for easy in while also being secure. 
[block:api-header]
{
  "title": "Naming Conventions"
}
[/block]
A unique client identifier is provided.  For example, a company named "Foo Industries" could be provided a client identifier of "foo".  All topics for this client will be prefixed with "foo-" and the provided credentials will allow for you to produce and consume from those topics. In addition, only consumer groups with the same prefix are allowed.  **--group foo-123** is valid while **--group foo123** is invalid.

You are not allowed to create topics or perform any administration of the topics. Topics are created based on your business use case.  In addition a topic **{{client}}-text** is created for your connection testing.  

A unique client identifier is provided as part of onboarding. 
[block:api-header]
{
  "title": "Client Configuration"
}
[/block]
Data is encrypted in-flight with TLS certificates. These certificates are currently self-signed and require the installation of a certificate of trust.  Authentication and Authorization are handled by a provided key/secret pair leveraging SASL Scram. If your secret is internally compromised, please contact Rill Data immediately to generate and issue a new secret or even a new complete key/secret upon your request.

#### Provided Information

* Bootstrap Servers - the host(s) used to discover the Apache Kafka Cluster by your client(s).
* SASL Scram Credentials  - key and secret
* Rill Data CA Certificate - provided both as a .pem as well as in a pre-packaged .jks file with a simple password. 

Both the Scram `key` and `secret` however, are to be treated as confidential and should not be shared.  The secret is salted and hashed (SHA-512) and leverages the standard Apache Kafka Scram security module.  Security patches and updates are closely monitored.

For additional security, you can request the credentials have ONLY producing authorization, which would prevent your team to be able to read anything published to RillData services with these credentials.

#### Bootstrap Servers
[block:code]
{
  "codes": [
    {
      "code": "--bootstrap-servers {{ hostname }}:{{ port }},{{ hostname }}:{{ port }}",
      "language": "shell"
    }
  ]
}
[/block]
#### Java clients
[block:code]
{
  "codes": [
    {
      "code": "security.protocol=SASL_SSL\nsasl.mechanism=SCRAM-SHA-512\nsasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required \\\n    username=\"{{ key }}\" \\\n    password=\"{{ secret }}\";\nssl.truststore.location={{ path to your JKS file with RillData Kafka Truststore }}\nssl.truststore.password={{ password for your truststore }}",
      "language": "shell"
    }
  ]
}
[/block]
#### librdkafka clients
[block:code]
{
  "codes": [
    {
      "code": "security.protocol=SASL_SSL\nsasl.mechanism=SCRAM-SHA-512\nsasl.username={{ key }}\nsasl.password={{ secret }}\nssl.ca.location={{ path to the RillData pem file }}",
      "language": "shell"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Topics"
}
[/block]
Topics are created by RillData.  As indicated by naming conventions, they are prefixed with your client identifier. Beyond that, they can be named whatever makes sense to your publishing pipeline, as long as they follow Apache Kafka Topic naming conventions.  Topics are created by Rill Data.  Retention time and other settings are part of the onboarding process.

It is recommended to publish data with compression. `compression.type` of the topic will be set to `producer` meaning it will honor what is set.

It is also recommended to submit with `acks`=`all` and to ensure the client processes and handles accordingly.  Additional recommended Kafka producer settings are provided as part of the onboarding process.
[block:api-header]
{
  "title": "Testing Connectivity"
}
[/block]
If your client is written in Java, it is recommended to use the `kafka-console-` commands to validate connection, since the configuration would be identical. If your client is written in a non-Java language and leverages the librdkafka library, then use kafkacat `kcat` for validation of the credentials.

Leverage the appropriate configuration files as shown above in the client setup recommendations.

###Java

####Publishing
[block:code]
{
  "codes": [
    {
      "code": "kafka-console-producer \\\n\t\t--bootstrap-server {{ server }}:{{ port }},{{ server }}:{{ port }} \\\n    --producer.config ./client.conf \\\n    --topic {{ client }}-text\n    ",
      "language": "shell"
    }
  ]
}
[/block]
####Consuming
[block:code]
{
  "codes": [
    {
      "code": "kafka-console-consumer \\\n\t\t--bootstrap-server {{ server }}:{{ port }},{{ server }}:{{ port }} \\\n    --consumer.config ./client.conf \\\n    --from-beginning \\\n    --group {{ client }}-000 \\\n    --topic {{ client }}-text",
      "language": "shell"
    }
  ]
}
[/block]
### librdkafka

If your client leverages **librdkafka** library, the best tool for establishing connectivity is the **kcat** (formally kafkacat) command-line tool.  This took leverages the librdkafka library so configuration for this would more closely resemble the configuration of your client.

####publishing
[block:code]
{
  "codes": [
    {
      "code": "kcat -b {{ server }}:{{ port }},{{ server }}:{{ port }} \\\n    -F ./client.conf \\\n    -P \\\n    -t {{ client }}-text",
      "language": "shell"
    }
  ]
}
[/block]
###consuming
[block:code]
{
  "codes": [
    {
      "code": "kcat -b {{ server }}:{{ port }},{{ server }}:{{ port }} \\\n    -F ./client.conf \\\n    -C \\\n    -t {{ client }}-text",
      "language": "shell"
    }
  ]
}
[/block]


### Trouble-Shooting

#### SSL Handshake

Leverage the following when starting your Kafka client (KAFKA_OPTS for kafka-console-producer or kafka-console-consumer) to help uncover any connectivity issues.

**-Djavax.net.debug=ssl,handshake**