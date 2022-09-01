---
title: "AWS Encryption"
slug: "aws-encryption"
excerpt: "with Customer Managed Keys"
---
We can encrypt the data with Customer managed AWS KMS Key.

## Create a new AWS KMS Key

### Using AWS Console

1. Open AWS KMS Console - http://console.aws.amazon.com/kms/home?#/kms/home 
2. Create a new KMS Key
![](https://files.readme.io/bf00a15-Screen_Shot_2020-10-14_at_9.18.53_PM.png)

3. Use *Key type:* Symmetric, *Key material origin:* KMS
![](https://files.readme.io/88755de-Screen_Shot_2020-10-14_at_9.19.44_PM.png)

4. Use the alias `rilldata` and add some tags
![](https://files.readme.io/945fe6c-Screen_Shot_2020-10-14_at_9.31.47_PM.png)

5. Define the key administrative permissions as per your requirements. 
![](https://files.readme.io/7e439c3-Screen_Shot_2020-10-14_at_9.32.30_PM.png)

6. For cross account access, add the use the Rill AWS account `248432388601` to use KMS key ID.
![](https://files.readme.io/b14dc08-Screen_Shot_2020-10-14_at_9.33.30_PM.png)

7. Finish creating the KMS Key
![](https://files.readme.io/a3c7183-Screen_Shot_2020-10-14_at_9.48.51_PM.png)

8. Share the ARN for the Customer Managed Key
![](https://files.readme.io/ee51154-Screen_Shot_2020-10-14_at_11.58.17_PM.png)

