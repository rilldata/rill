---
title: "AWS Encryption"
slug: "aws-encryption"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="with Customer Managed Keys" />
We can encrypt the data with Customer managed AWS KMS Key.

## Create a new AWS KMS Key

### Using AWS Console

1. Open AWS KMS Console - http://console.aws.amazon.com/kms/home?#/kms/home 
2. Create a new KMS Key
![](https://images.contentful.com/ve6smfzbifwz/6yEhLEDbGXlxPcAxj7Ux5e/fbb5a5f967685e6de6763522636229d1/bf00a15-Screen_Shot_2020-10-14_at_9.18.53_PM.png)

3. Use *Key type:* Symmetric, *Key material origin:* KMS
![](https://images.contentful.com/ve6smfzbifwz/MWSRi0rb9nEwyQwZWeGpY/de3b90c417d5e36dd9ea56b0e2d1bc85/88755de-Screen_Shot_2020-10-14_at_9.19.44_PM.png)

4. Use the alias `rilldata` and add some tags
![](https://images.contentful.com/ve6smfzbifwz/ALR8jvFTH2I2OjEaSVfiO/74e417fd29e2d154f481f2dbd2b16ded/945fe6c-Screen_Shot_2020-10-14_at_9.31.47_PM.png)

5. Define the key administrative permissions as per your requirements. 
![](https://images.contentful.com/ve6smfzbifwz/tvHhNYjRTG5gdiL1u2Txt/831854e216be737b9d9b25cfd3672642/7e439c3-Screen_Shot_2020-10-14_at_9.32.30_PM.png)

6. For cross account access, add the use the Rill AWS account `248432388601` to use KMS key ID.
![](https://images.contentful.com/ve6smfzbifwz/15IzXDXCr9JtHnoDP32Kza/5eda0babdb2354499bfd89e2ae04dbc9/b14dc08-Screen_Shot_2020-10-14_at_9.33.30_PM.png)

7. Finish creating the KMS Key
![](https://images.contentful.com/ve6smfzbifwz/39zvxRGtE9P4Q0XvQinUcU/44560a3fcd41730c1fb5d15703e630e2/a3c7183-Screen_Shot_2020-10-14_at_9.48.51_PM.png)

8. Share the ARN for the Customer Managed Key
![](https://images.contentful.com/ve6smfzbifwz/7x6mLPP1WsVgd4Uo3JDcg5/e1770b59dd55b34f003697edb76a1801/ee51154-Screen_Shot_2020-10-14_at_11.58.17_PM.png)

