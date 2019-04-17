load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "77dfd303492f2634de7a660445ee2d3de2960cbd52f97d8c0dffa9362d3ddef9",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.1/rules_go-0.18.1.tar.gz"],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "co_honnef_go_tools",
    commit = "3f1c8253044a",
    importpath = "honnef.co/go/tools",
)

go_repository(
    name = "com_github_alecthomas_template",
    commit = "a0175ee3bccc",
    importpath = "github.com/alecthomas/template",
)

go_repository(
    name = "com_github_alecthomas_units",
    commit = "2efee857e7cf",
    importpath = "github.com/alecthomas/units",
)

go_repository(
    name = "com_github_anmitsu_go_shlex",
    commit = "648efa622239",
    importpath = "github.com/anmitsu/go-shlex",
)

go_repository(
    name = "com_github_apache_thrift",
    importpath = "github.com/apache/thrift",
    tag = "v0.12.0",
)

go_repository(
    name = "com_github_beorn7_perks",
    commit = "3a771d992973",
    importpath = "github.com/beorn7/perks",
)

go_repository(
    name = "com_github_bradfitz_go_smtpd",
    commit = "deb6d6237625",
    importpath = "github.com/bradfitz/go-smtpd",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    tag = "v0.3.1",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    tag = "v0.3.4",
)

go_repository(
    name = "com_github_coreos_go_systemd",
    commit = "c6f51f82210d",
    importpath = "github.com/coreos/go-systemd",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_eapache_go_resiliency",
    importpath = "github.com/eapache/go-resiliency",
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_eapache_go_xerial_snappy",
    commit = "776d5712da21",
    importpath = "github.com/eapache/go-xerial-snappy",
)

go_repository(
    name = "com_github_eapache_queue",
    importpath = "github.com/eapache/queue",
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_flynn_go_shlex",
    commit = "3f9db97f8568",
    importpath = "github.com/flynn/go-shlex",
)

go_repository(
    name = "com_github_fsnotify_fsnotify",
    importpath = "github.com/fsnotify/fsnotify",
    tag = "v1.4.7",
)

go_repository(
    name = "com_github_gliderlabs_ssh",
    importpath = "github.com/gliderlabs/ssh",
    tag = "v0.1.1",
)

go_repository(
    name = "com_github_go_kit_kit",
    importpath = "github.com/go-kit/kit",
    tag = "v0.8.0",
)

go_repository(
    name = "com_github_go_logfmt_logfmt",
    importpath = "github.com/go-logfmt/logfmt",
    tag = "v0.3.0",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    tag = "v1.8.0",
)

go_repository(
    name = "com_github_gogo_protobuf",
    importpath = "github.com/gogo/protobuf",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_golang_snappy",
    commit = "2e65f85255db",
    importpath = "github.com/golang/snappy",
)

go_repository(
    name = "com_github_google_btree",
    commit = "4030bb1f1f0c",
    importpath = "github.com/google/btree",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    tag = "v0.2.0",
)

go_repository(
    name = "com_github_google_go_github",
    importpath = "github.com/google/go-github",
    tag = "v17.0.0",
)

go_repository(
    name = "com_github_google_go_querystring",
    importpath = "github.com/google/go-querystring",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_google_martian",
    importpath = "github.com/google/martian",
    tag = "v2.1.0",
)

go_repository(
    name = "com_github_google_pprof",
    commit = "3ea8567a2e57",
    importpath = "github.com/google/pprof",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    importpath = "github.com/googleapis/gax-go",
    tag = "v2.0.4",
)

go_repository(
    name = "com_github_gorilla_context",
    importpath = "github.com/gorilla/context",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_mux",
    importpath = "github.com/gorilla/mux",
    tag = "v1.6.2",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "9cad4c3443a7",
    importpath = "github.com/gregjones/httpcache",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    tag = "v0.5.0",
)

go_repository(
    name = "com_github_hpcloud_tail",
    importpath = "github.com/hpcloud/tail",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_jellevandenhooff_dkim",
    commit = "f50fe3d243e1",
    importpath = "github.com/jellevandenhooff/dkim",
)

go_repository(
    name = "com_github_jstemmer_go_junit_report",
    commit = "af01ea7f8024",
    importpath = "github.com/jstemmer/go-junit-report",
)

go_repository(
    name = "com_github_julienschmidt_httprouter",
    importpath = "github.com/julienschmidt/httprouter",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_kr_logfmt",
    commit = "b84e30acd515",
    importpath = "github.com/kr/logfmt",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    tag = "v1.1.3",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_mwitkow_go_conntrack",
    commit = "cc309e4a2223",
    importpath = "github.com/mwitkow/go-conntrack",
)

go_repository(
    name = "com_github_onsi_ginkgo",
    importpath = "github.com/onsi/ginkgo",
    tag = "v1.7.0",
)

go_repository(
    name = "com_github_onsi_gomega",
    importpath = "github.com/onsi/gomega",
    tag = "v1.4.3",
)

go_repository(
    name = "com_github_openzipkin_zipkin_go",
    importpath = "github.com/openzipkin/zipkin-go",
    tag = "v0.1.6",
)

go_repository(
    name = "com_github_pierrec_lz4",
    importpath = "github.com/pierrec/lz4",
    tag = "v2.0.5",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    tag = "v0.8.0",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "3c4408c8b829",
    importpath = "github.com/prometheus/client_golang",
)

go_repository(
    name = "com_github_prometheus_client_model",
    commit = "56726106282f",
    importpath = "github.com/prometheus/client_model",
)

go_repository(
    name = "com_github_prometheus_common",
    importpath = "github.com/prometheus/common",
    tag = "v0.2.0",
)

go_repository(
    name = "com_github_prometheus_procfs",
    commit = "bf6a532e95b1",
    importpath = "github.com/prometheus/procfs",
)

go_repository(
    name = "com_github_rcrowley_go_metrics",
    commit = "3113b8401b8a",
    importpath = "github.com/rcrowley/go-metrics",
)

go_repository(
    name = "com_github_shopify_sarama",
    importpath = "github.com/Shopify/sarama",
    tag = "v1.19.0",
)

go_repository(
    name = "com_github_shopify_toxiproxy",
    importpath = "github.com/Shopify/toxiproxy",
    tag = "v2.1.4",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    tag = "v1.4.0",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    tag = "v0.0.3",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    tag = "v1.0.3",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    tag = "v0.1.1",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    tag = "v1.2.2",
)

go_repository(
    name = "com_github_tarm_serial",
    commit = "98f6abe2eb07",
    importpath = "github.com/tarm/serial",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    tag = "v0.37.2",
)

go_repository(
    name = "in_gopkg_alecthomas_kingpin_v2",
    importpath = "gopkg.in/alecthomas/kingpin.v2",
    tag = "v2.2.6",
)

go_repository(
    name = "in_gopkg_check_v1",
    commit = "788fd7840127",
    importpath = "gopkg.in/check.v1",
)

go_repository(
    name = "in_gopkg_fsnotify_v1",
    importpath = "gopkg.in/fsnotify.v1",
    tag = "v1.4.7",
)

go_repository(
    name = "in_gopkg_inf_v0",
    importpath = "gopkg.in/inf.v0",
    tag = "v0.9.1",
)

go_repository(
    name = "in_gopkg_tomb_v1",
    commit = "dd632973f1e7",
    importpath = "gopkg.in/tomb.v1",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    tag = "v2.2.2",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    tag = "v0.19.2",
)

go_repository(
    name = "org_go4",
    commit = "417644f6feb5",
    importpath = "go4.org",
)

go_repository(
    name = "org_go4_grpc",
    commit = "11d0a25b4919",
    importpath = "grpc.go4.org",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    tag = "v0.3.0",
)

go_repository(
    name = "org_golang_google_appengine",
    importpath = "google.golang.org/appengine",
    tag = "v1.4.0",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "d831d65fe17d",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    tag = "v1.19.1",
)

go_repository(
    name = "org_golang_x_build",
    commit = "5284462c4bec",
    importpath = "golang.org/x/build",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "c2843e01d9a2",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_exp",
    commit = "509febef88a4",
    importpath = "golang.org/x/exp",
)

go_repository(
    name = "org_golang_x_lint",
    commit = "5614ed5bae6f",
    importpath = "golang.org/x/lint",
)

go_repository(
    name = "org_golang_x_net",
    commit = "d8887717615a",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "e64efc72b421",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_perf",
    commit = "6e6d33e29852",
    importpath = "golang.org/x/perf",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "e225da77a7e6",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "d0b11bdaac8a",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "org_golang_x_text",
    commit = "17ff2d5776d2",
    importpath = "golang.org/x/text",
)

go_repository(
    name = "org_golang_x_time",
    commit = "85acf8d2951c",
    importpath = "golang.org/x/time",
)

go_repository(
    name = "org_golang_x_tools",
    commit = "e65039ee4138",
    importpath = "golang.org/x/tools",
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    importpath = "github.com/googleapis/gax-go/v2",
    tag = "v2.0.4",
)
