# Homebrew formula for Confire.
#
# This file lives in confire-ai/mono and is copied to the tap repo
# (confire-ai/homebrew-confire) as part of the release workflow.
#
# To install:
#   brew tap confire-ai/confire
#   brew install confire
#
# After a release, update the version, url, and sha256 fields below,
# then push to github.com/confire-ai/homebrew-confire.

class Confire < Formula
  desc "Context and tool firewall for Claude Code"
  homepage "https://confire.dev"
  version "0.9.1"
  license "Proprietary"

  on_macos do
    on_intel do
      url "https://releases.confire.dev/v#{version}/confire_darwin_amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end

    on_arm do
      url "https://releases.confire.dev/v#{version}/confire_darwin_arm64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
  end

  on_linux do
    on_intel do
      url "https://releases.confire.dev/v#{version}/confire_linux_amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end

    on_arm do
      url "https://releases.confire.dev/v#{version}/confire_linux_arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
  end

  def install
    bin.install stable.url.split("/").last => "confire"
  end

  def caveats
    <<~EOS
      To set up Confire hooks in Claude Code:
        confire setup

      To connect your account:
        confire login

      To start the optimizer:
        confire start
    EOS
  end

  test do
    assert_match "confire", shell_output("#{bin}/confire version")
  end
end
