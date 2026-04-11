class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  license "MIT"
  url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v0.1.1.tar.gz"
  sha256 "5e5c97bef32294eb55691fb50a8761b0d28bce9078f8ab94a9220ae3937fecfe"
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
