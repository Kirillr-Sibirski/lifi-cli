class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  license "Apache-2.0"
  url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v0.1.6.tar.gz"
  sha256 "d105b322828258399a834d89af715ed2cc9fa192751f634a7ccd35280bd850de"
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
