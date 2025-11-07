# typed: false
# frozen_string_literal: true

# Local testing formula for aws-ssm
class AwsSsmLocal < Formula
  desc "Native Golang CLI tool for managing AWS SSM sessions"
  homepage "https://github.com/johnlam90/aws-ssm"
  version "0.1.0"
  license "MIT"

  # For local testing, we'll install from the built binary
  def install
    bin.install "/Users/nokia/aws-ssm/aws-ssm"
  end

  test do
    system "#{bin}/aws-ssm", "version"
  end
end

