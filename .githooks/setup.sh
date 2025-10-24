#!/bin/bash
# Setup git hooks for MarkGo development
# Run this after cloning the repository: ./githooks/setup.sh

set -e

echo "🔧 Setting up MarkGo git hooks..."

# Get the repository root
REPO_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="$REPO_ROOT/.githooks"
GIT_HOOKS_DIR="$REPO_ROOT/.git/hooks"

# Check if .githooks directory exists
if [ ! -d "$HOOKS_DIR" ]; then
    echo "❌ .githooks directory not found!"
    exit 1
fi

# Install pre-commit hook
echo "📋 Installing pre-commit hook..."
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    cp "$HOOKS_DIR/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
    chmod +x "$GIT_HOOKS_DIR/pre-commit"
    echo "✅ pre-commit hook installed"
else
    echo "⚠️  pre-commit hook not found in .githooks/"
fi

# Configure git to use .githooks directory (Git 2.9+)
if git config core.hooksPath .githooks 2>/dev/null; then
    echo "✅ Git configured to use .githooks directory"
else
    echo "ℹ️  Using manual hook installation (Git < 2.9)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ Git hooks setup complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "The following checks will run before each commit:"
echo "  🔐 Secret detection"
echo "  ✨ Code formatting (gofmt)"
echo "  🔬 Go vet analysis"
echo "  🧪 Test execution"
echo ""
echo "To bypass checks (not recommended):"
echo "  git commit --no-verify"
echo ""
