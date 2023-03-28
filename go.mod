module github.com/rilldata/rill

go 1.19

require (
	cloud.google.com/go/storage v1.27.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/MicahParks/keyfunc v1.9.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/apache/arrow/go/v11 v11.0.0
	github.com/apache/calcite-avatica-go/v5 v5.1.0
	github.com/aws/aws-sdk-go v1.44.118
	github.com/benbjohnson/clock v1.1.0
	github.com/bmatcuk/doublestar/v4 v4.4.0
	github.com/bradleyfalzon/ghinstallation v1.1.1
	github.com/briandowns/spinner v1.22.0
	github.com/c2h5oh/datasize v0.0.0-20220606134207-859f65c6625b
	github.com/coreos/go-oidc/v3 v3.4.0
	github.com/eapache/go-resiliency v1.3.0
	github.com/envoyproxy/protoc-gen-validate v0.9.1
	github.com/fatih/color v1.13.0
	github.com/go-git/go-git/v5 v5.5.2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/google/go-github/v50 v50.1.0
	github.com/gorilla/sessions v1.2.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jackc/pgx/v4 v4.16.1
	github.com/jinzhu/copier v0.3.5
	github.com/joho/godotenv v1.5.1
	github.com/lensesio/tableprinter v0.0.0-20201125135848-89e81fc956e7
	github.com/manifoldco/promptui v0.9.0
	github.com/marcboeker/go-duckdb v1.2.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.8.2
	github.com/spf13/cobra v1.6.1
	github.com/stretchr/testify v1.8.0
	github.com/testcontainers/testcontainers-go v0.13.0
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.21.0
	gocloud.dev v0.27.0
	golang.org/x/exp v0.0.0-20221230185412-738e83a70c30
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1
	google.golang.org/grpc v1.49.0
	gopkg.in/square/go-jose.v2 v2.6.0
	modernc.org/sqlite v1.10.6
)

require (
	cloud.google.com/go v0.104.0 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20221026131551-cf6655e29de4 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.8 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.15.15 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.10 // indirect
	github.com/aws/smithy-go v1.12.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chzyer/readline v1.5.0 // indirect
	github.com/cloudflare/circl v1.1.0 // indirect
	github.com/containerd/cgroups v1.0.3 // indirect
	github.com/containerd/containerd v1.6.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.17+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.4.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-github/v29 v29.0.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.5.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/icholy/digest v0.1.15 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kataras/tablewriter v0.0.0-20180708051242-e063d29b7c23 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.2 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.5.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pjbgf/sha1cd v0.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/tools v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.97.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.1.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.21.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.4.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)

require (
	github.com/alecthomas/participle/v2 v2.0.0-beta.5
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2 v2.0.0-20220505152843-58a406687f2d
	github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2 v2.0.0-20220505152843-58a406687f2d
	github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2 v2.0.0-20220505152843-58a406687f2d
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20220505152843-58a406687f2d
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.3
	github.com/jmoiron/sqlx v1.3.5
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.1.0
	golang.org/x/sys v0.5.0
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20220920201722-2b89144ce006
	google.golang.org/protobuf v1.28.1
)

replace github.com/apache/calcite-avatica-go/v5 v5.1.0 => github.com/begelundmuller/calcite-avatica-go/v5 v5.0.0-20230323102415-9c0eba2fbe15

exclude modernc.org/sqlite v1.18.1
