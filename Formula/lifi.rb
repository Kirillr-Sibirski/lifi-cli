class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  license "MIT"
  url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "5a30330da7db3632f89c37071e0b601da8f2aba61b4996ee1e909f966ba5cf64"
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
