import * as yup from "yup";

export interface ConnectorSpec {
  name: string;
  title: string;
  description: string;
  fields: {
    [name: string]: {
      type: string;
      required: boolean;
      label: string;
      placeholder: string;
      hint?: string;
    };
  };
}

export const HTTP: ConnectorSpec = {
  name: "http",
  title: "HTTP(S)",
  description: "Connect to a CSV or Parquet file via HTTP(S)",
  fields: {
    url: {
      type: "text",
      required: true,
      label: "URL",
      placeholder: "https://example.com/data.parquet",
    },
    // sample1MRows: {
    //   type: "checkbox",
    //   label: "Sample 1M rows",
    //   placeholder: "",
    //   required: true,
    // },
  },
};

export const HTTPYupSchema: yup.AnyObjectSchema = yup.object().shape({
  url: yup.string().url().required(),
  sample1MRows: yup.boolean().required(),
});

export const S3: ConnectorSpec = {
  name: "s3",
  title: "S3",
  description:
    "Connect to CSV or Parquet files in an S3 bucket. For private buckets, provide an <a href=https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html target='_blank'>access key</a>.",
  fields: {
    url: {
      type: "text",
      label: "URL",
      placeholder: "s3://bucket-name/path/to/file.csv",
      required: true,
      hint: "Tip: use glob patterns to select multiple files",
    },
    region: {
      type: "text",
      label: "Region",
      placeholder: "us-east-1",
      required: true,
    },
    accessKeyId: {
      type: "text",
      label: "Access key ID",
      placeholder: "...",
      required: false,
    },
    secretAccessKey: {
      type: "text",
      label: "Secret access key",
      placeholder: "...",
      required: false,
    },
    // sample1MRows: {
    //   type: "checkbox",
    //   label: "Sample 1M rows",
    //   placeholder: "",
    //   required: true,
    // },
  },
};

export const S3YupSchema: yup.AnyObjectSchema = yup.object().shape({
  url: yup
    .string()
    .matches(/^s3:\/\//, "Must be an S3 URL")
    .required(),
  region: yup.string().required(),
  sessionToken: yup.string(),
  accessKeyId: yup.string(),
  secretAccessKey: yup.string(),
  sample1MRows: yup.boolean().required(),
});

export const GCS: ConnectorSpec = {
  name: "gcs",
  title: "GCS",
  description:
    "Connect to CSV or Parquet files in a GCS bucket. For private buckets, provide <a href=https://console.cloud.google.com/storage/settings;tab=interoperability target='_blank'>HMAC credentials</a>.",
  fields: {
    url: {
      type: "text",
      label: "URL",
      placeholder: "gcs://bucket-name/path/to/file.csv",
      required: true,
      hint: "Tip: use glob patterns to select multiple files",
    },
    region: {
      type: "text",
      label: "Region",
      placeholder: "us-east-1",
      required: true,
    },
    accessKeyId: {
      type: "text",
      label: "Access key ID",
      placeholder: "...",
      required: false,
    },
    secretAccessKey: {
      type: "text",
      label: "Secret access key",
      placeholder: "...",
      required: false,
    },
    // sample1MRows: {
    //   type: "checkbox",
    //   label: "Sample 1M rows",
    //   placeholder: "",
    //   required: true,
    // },
  },
};

export const GCSYupSchema = yup.object().shape({
  url: yup
    .string()
    .matches(/^gcs:\/\//, "Must be a GCS URL")
    .required(),
  region: yup.string().required(),
  sessionToken: yup.string(),
  accessKeyId: yup.string(),
  secretAccessKey: yup.string(),
  sample1MRows: yup.boolean().required(),
});
