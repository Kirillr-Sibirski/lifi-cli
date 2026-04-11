class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  license "Apache-2.0"
  url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v0.1.4.tar.gz"
  sha256 "a51d9d92144ef7bf67b6561e09016a36a286f21ece85babddc074c6e5eb38fed"
  head "https://github.com/Kirillr-Sibirski/lifi-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/Kirillr-Sibirski/lifi-cli/internal/cli.version=#{version}"
    system "go", "build", *std_go_args(ldflags:), "./cmd/lifi"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/lifi version")
  end
end
