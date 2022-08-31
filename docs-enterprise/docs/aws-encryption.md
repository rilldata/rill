---
title: "AWS Encryption"
slug: "aws-encryption"
excerpt: "with Customer Managed Keys"
hidden: false
createdAt: "2020-10-14T18:13:35.244Z"
updatedAt: "2020-10-14T18:30:42.190Z"
---
We can encrypt the data with Customer managed AWS KMS Key.

# Create a new AWS KMS Key

## Using AWS Console

1. Open AWS KMS Console - http://console.aws.amazon.com/kms/home?#/kms/home 
2. Create a new KMS Key
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/bf00a15-Screen_Shot_2020-10-14_at_9.18.53_PM.png",
        "Screen Shot 2020-10-14 at 9.18.53 PM.png",
        2360,
        768,
        "#dedfe0"
      ]
    }
  ]
}
[/block]
3. Use *Key type:* Symmetric, *Key material origin:* KMS
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/88755de-Screen_Shot_2020-10-14_at_9.19.44_PM.png",
        "Screen Shot 2020-10-14 at 9.19.44 PM.png",
        2366,
        1320,
        "#e9eaeb"
      ]
    }
  ]
}
[/block]
4. Use the alias `rilldata` and add some tags
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/945fe6c-Screen_Shot_2020-10-14_at_9.31.47_PM.png",
        "Screen Shot 2020-10-14 at 9.31.47 PM.png",
        2366,
        1680,
        "#eceded"
      ]
    }
  ]
}
[/block]
5. Define the key administrative permissions as per your requirements. 

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/7e439c3-Screen_Shot_2020-10-14_at_9.32.30_PM.png",
        "Screen Shot 2020-10-14 at 9.32.30 PM.png",
        2364,
        1342,
        "#e8e8e9"
      ]
    }
  ]
}
[/block]
6. For cross account access, add the use the Rill AWS account `248432388601` to use KMS key ID.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/b14dc08-Screen_Shot_2020-10-14_at_9.33.30_PM.png",
        "Screen Shot 2020-10-14 at 9.33.30 PM.png",
        2354,
        1570,
        "#eaeaeb"
      ]
    }
  ]
}
[/block]
7. Finish creating the KMS Key
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/a3c7183-Screen_Shot_2020-10-14_at_9.48.51_PM.png",
        "Screen Shot 2020-10-14 at 9.48.51 PM.png",
        2360,
        1782,
        "#efeeef"
      ]
    }
  ]
}
[/block]
8. Share the ARN for the Customer Managed Key
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/ee51154-Screen_Shot_2020-10-14_at_11.58.17_PM.png",
        "Screen Shot 2020-10-14 at 11.58.17 PM.png",
        2848,
        1126,
        "#e3e5e6"
      ]
    }
  ]
}
[/block]