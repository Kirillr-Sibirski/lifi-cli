class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  license "MIT"
  url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v0.1.3.tar.gz"
  sha256 "62503d0b11c8b4a9c3bfd7ad8718bf98774ba5cfc616ea3fe7148b988a0ae854"
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
