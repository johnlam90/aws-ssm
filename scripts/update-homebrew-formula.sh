#!/bin/bash
set -e

# Script to update the Homebrew formula for aws-ssm
# Usage: ./scripts/update-homebrew-formula.sh <version>
# Example: ./scripts/update-homebrew-formula.sh 0.2.0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if version is provided
if [ -z "$1" ]; then
  echo -e "${RED}Error: Version number required${NC}"
  echo "Usage: $0 <version>"
  echo "Example: $0 0.2.0"
  exit 1
fi

VERSION=$1
TAG="v${VERSION}"

echo -e "${GREEN}==> Updating Homebrew formula to version ${VERSION}...${NC}"
echo ""

# Verify the release exists
echo -e "${YELLOW}Checking if release ${TAG} exists...${NC}"
if ! curl -sf "https://api.github.com/repos/johnlam90/aws-ssm/releases/tags/${TAG}" > /dev/null; then
  echo -e "${RED}Error: Release ${TAG} not found on GitHub${NC}"
  echo "Please create the release first by pushing the tag:"
  echo "  git tag -a ${TAG} -m 'Release version ${VERSION}'"
  echo "  git push origin ${TAG}"
  exit 1
fi
echo -e "${GREEN}✓ Release ${TAG} found${NC}"
echo ""

# Download checksums
echo -e "${YELLOW}Downloading checksums...${NC}"
if ! curl -sL "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/checksums.txt" -o /tmp/checksums.txt; then
  echo -e "${RED}Error: Failed to download checksums.txt${NC}"
  echo "Make sure the release workflow has completed and checksums.txt is available."
  exit 1
fi
echo -e "${GREEN}✓ Checksums downloaded${NC}"
echo ""

# Extract SHA256 hashes
echo -e "${YELLOW}Extracting SHA256 hashes...${NC}"
DARWIN_AMD64_SHA=$(grep "darwin-amd64" /tmp/checksums.txt | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep "darwin-arm64" /tmp/checksums.txt | awk '{print $1}')
LINUX_AMD64_SHA=$(grep "linux-amd64" /tmp/checksums.txt | awk '{print $1}')
LINUX_ARM64_SHA=$(grep "linux-arm64" /tmp/checksums.txt | awk '{print $1}')

# Verify all hashes were extracted
if [ -z "$DARWIN_AMD64_SHA" ] || [ -z "$DARWIN_ARM64_SHA" ] || [ -z "$LINUX_AMD64_SHA" ] || [ -z "$LINUX_ARM64_SHA" ]; then
  echo -e "${RED}Error: Failed to extract all SHA256 hashes${NC}"
  echo "Checksums file content:"
  cat /tmp/checksums.txt
  exit 1
fi

echo "SHA256 Hashes:"
echo "  darwin-amd64: $DARWIN_AMD64_SHA"
echo "  darwin-arm64: $DARWIN_ARM64_SHA"
echo "  linux-amd64:  $LINUX_AMD64_SHA"
echo "  linux-arm64:  $LINUX_ARM64_SHA"
echo ""

# Clone or update homebrew-tap
if [ -d "/tmp/homebrew-tap" ]; then
  echo -e "${YELLOW}Updating existing homebrew-tap clone...${NC}"
  cd /tmp/homebrew-tap
  git fetch origin
  git reset --hard origin/main
else
  echo -e "${YELLOW}Cloning homebrew-tap...${NC}"
  git clone https://github.com/johnlam90/homebrew-tap.git /tmp/homebrew-tap
  cd /tmp/homebrew-tap
fi
echo -e "${GREEN}✓ Repository ready${NC}"
echo ""

# Create updated formula
echo -e "${YELLOW}Updating formula...${NC}"
cat > aws-ssm.rb << EOF
# typed: false
# frozen_string_literal: true

class AwsSsm < Formula
  desc "Native Golang CLI tool for managing AWS SSM sessions"
  homepage "https://github.com/johnlam90/aws-ssm"
  version "${VERSION}"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-darwin-amd64.tar.gz"
      sha256 "${DARWIN_AMD64_SHA}"

      def install
        bin.install "aws-ssm-darwin-amd64" => "aws-ssm"
      end
    end

    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-darwin-arm64.tar.gz"
      sha256 "${DARWIN_ARM64_SHA}"

      def install
        bin.install "aws-ssm-darwin-arm64" => "aws-ssm"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-linux-amd64.tar.gz"
      sha256 "${LINUX_AMD64_SHA}"

      def install
        bin.install "aws-ssm-linux-amd64" => "aws-ssm"
      end
    end

    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM64_SHA}"

      def install
        bin.install "aws-ssm-linux-arm64" => "aws-ssm"
      end
    end
  end

  test do
    system "#{bin}/aws-ssm", "version"
  end
end
EOF

echo -e "${GREEN}✓ Formula updated${NC}"
echo ""

# Show diff
echo -e "${YELLOW}Changes to be committed:${NC}"
git diff aws-ssm.rb
echo ""

# Commit and push
read -p "Commit and push changes? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  git add aws-ssm.rb
  git commit -m "Update aws-ssm formula to v${VERSION}"
  
  echo -e "${YELLOW}Pushing to GitHub...${NC}"
  git push origin main
  
  echo ""
  echo -e "${GREEN}✓ Formula updated successfully!${NC}"
  echo ""
  echo "Next steps:"
  echo "  1. Test the formula: brew upgrade aws-ssm"
  echo "  2. Verify version: aws-ssm version"
  echo "  3. Monitor for user issues"
else
  echo -e "${YELLOW}Changes not committed.${NC}"
  echo "Formula file is available at: /tmp/homebrew-tap/aws-ssm.rb"
fi

# Cleanup
rm /tmp/checksums.txt

echo ""
echo -e "${GREEN}Done!${NC}"

