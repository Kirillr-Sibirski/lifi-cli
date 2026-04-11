class Lifi < Formula
  desc "CLI for LI.FI Earn and Composer"
  homepage "https://github.com/Kirillr-Sibirski/lifi-cli"
  head "https://github.com/Kirillr-Sibirski/lifi-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/Kirillr-Sibirski/lifi-cli/internal/cli.version=head"
    system "go", "build", *std_go_args(ldflags:), "./cmd/lifi"
  end

  test do
    assert_match "lifi", shell_output("#{bin}/lifi version")
  end
end
