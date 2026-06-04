# Homebrew formula for Confire.
#
# This file lives in confire-ai/mono and is copied to the tap repo
# (confire-ai/homebrew-confire) as part of the release workflow.
#
# To install:
#   brew tap confire-ai/confire
#   brew install confire
#
# After a release, the release workflow auto-updates version + sha256
# and pushes to github.com/confire-ai/homebrew-confire.

class Confire < Formula
  desc "Deterministic client-side policy evaluation for AI systems"
  homepage "https://confire.dev"
  version "0.11.0"
  license :cannot_represent

  on_macos do
    on_intel do
      url "https://releases.confire.dev/v#{version}/confire_v#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end

    on_arm do
      url "https://releases.confire.dev/v#{version}/confire_v#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
  end

  on_linux do
    on_intel do
      url "https://releases.confire.dev/v#{version}/confire_v#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end

    on_arm do
      url "https://releases.confire.dev/v#{version}/confire_v#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
  end

  def install
    bin.install "confire"
  end

  def caveats
    <<~EOS
      To set up Confire for your AI agent:
        confire setup

      To connect your account:
        confire login

      To start the optimizer:
        confire start
    EOS
  end

  test do
    assert_match "Confire", shell_output("#{bin}/confire version")
  end
end
