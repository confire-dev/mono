# Homebrew formula for Confire.
#
# ABOUT THIS FILE
# ───────────────
# This file lives in confire-ai/mono (the private source repo) and is
# copied to the public tap repo (confire-ai/homebrew-confire) as part of
# the release workflow.
#
# WHY THE TAP CAN BE PUBLIC EVEN THOUGH THE SOURCE IS PRIVATE
# ─────────────────────────────────────────────────────────────
# Homebrew taps are just Git repos that contain formula (.rb) files.
# The formula only needs to reference the *release artifacts* (pre-built
# binaries hosted on get.confire.dev), not the source code. Users who
# run `brew install confire-ai/confire/confire` download a signed tarball,
# not source. The private mono repo never needs to be exposed.
#
# RELEASE WORKFLOW
# ────────────────
# After each release:
#   1. Build and push artifacts to get.confire.dev/releases/<version>/
#   2. Update `version`, `url`, and `sha256` fields in this file.
#   3. Copy/push the updated formula to github.com/confire-ai/homebrew-confire.
#
# INSTALL (end-user):
#   brew tap confire-ai/confire
#   brew install confire

class Confire < Formula
  desc "Deterministic client-side policy evaluation for AI systems"
  homepage "https://confire.dev"
  version "1.0.0"
  license :cannot_represent

  on_macos do
    if Hardware::CPU.arm?
      url "https://get.confire.dev/releases/v#{version}/confire_v#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_DARWIN_ARM64"
    else
      url "https://get.confire.dev/releases/v#{version}/confire_v#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://get.confire.dev/releases/v#{version}/confire_v#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_LINUX_ARM64"
    else
      url "https://get.confire.dev/releases/v#{version}/confire_v#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_LINUX_AMD64"
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

      To start the optimizer daemon:
        confire start

      Confire is proprietary software. See: https://confire.dev/terms
    EOS
  end

  test do
    assert_match "Confire", shell_output("#{bin}/confire version")
  end
end
