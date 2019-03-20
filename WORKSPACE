load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.1/rules_go-0.18.1.tar.gz"],
    sha256 = "77dfd303492f2634de7a660445ee2d3de2960cbd52f97d8c0dffa9362d3ddef9",
)

http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "com_github_sirupsen_logrus",
    commit = "dae0fa8d5b0c810a8ab733fbd5510c7cae84eca4",
    importpath = "github.com/sirupsen/logrus",
)

go_repository(
    name = "com_github_spf13_cobra",
    commit = "ba1052d4cbce7aac421a96de820558f75199ccbc",
    importpath = "github.com/spf13/cobra",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "942825909afd4837bd761b032cfb269ec165c3d3",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "beaecbbdd8af86aa3acf14180d53828ce69400b2",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "org_golang_google_api",
    commit = "2dc3ad4d67ba9a37200c5702d36687a940df1111",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "com_github_spf13_pflag",
    commit = "24fa6976df40757dce6aea913e7b81ade90530e1",
    importpath = "github.com/spf13/pflag",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "a1f597ede03a7bef967a422b5b3a5bd08805a01e",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "c85d3e98c914e3a33234ad863dcbff5dbc425bb8",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "io_opencensus_go",
    commit = "8a36f74db452c3eb69935e0ce66aecc030cf5142",
    importpath = "go.opencensus.io",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "7087cb70de9f7a8bc0a10c375cb0d2280a8edf9c",
    importpath = "github.com/hashicorp/golang-lru",
)
