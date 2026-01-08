#!/bin/bash
#
# Release validation and tagging script for Kion CLI.
# Validates preconditions and creates a release tag.

set -e

# Colors and formatting
RED=$(tput setaf 1)
GRN=$(tput setaf 2)
YLW=$(tput setaf 3)
BLU=$(tput setaf 4)
B=$(tput bold)
NRM=$(tput sgr0)

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$REPO_ROOT/VERSION.md"
CHANGELOG_FILE="$REPO_ROOT/CHANGELOG.md"

# Track validation status
ERRORS=()

info() {
    printf "${B}${BLU}▸${NRM} %s\n" "$1"
}

success() {
    printf "${B}${GRN}✓${NRM} %s\n" "$1"
}

warn() {
    printf "${B}${YLW}⚠${NRM} %s\n" "$1"
}

error() {
    printf "${B}${RED}✗${NRM} %s\n" "$1"
    ERRORS+=("$1")
}

header() {
    printf "\n${B}${BLU}%s${NRM}\n" "$1"
    printf "${B}${BLU}%s${NRM}\n" "${1//?/-}"
}

# Read and validate version from VERSION.md
validate_version() {
    header "Validating VERSION.md"

    if [[ ! -f "$VERSION_FILE" ]]; then
        error "VERSION.md file not found"
        return
    fi

    VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: '$VERSION' (expected vN.N.N)"
        return
    fi

    success "Version format valid: $VERSION"
}

# Ensure we're on main branch with clean working directory
validate_git_state() {
    header "Validating Git State"

    # Check current branch
    CURRENT_BRANCH=$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD)
    if [[ "$CURRENT_BRANCH" != "main" ]]; then
        error "Not on main branch (currently on: $CURRENT_BRANCH)"
    else
        success "On main branch"
    fi

    # Check for uncommitted changes
    if [[ -n $(git -C "$REPO_ROOT" status --porcelain) ]]; then
        error "Working directory has uncommitted changes"
    else
        success "Working directory is clean"
    fi

    # Fetch latest from origin
    info "Fetching latest from origin..."
    git -C "$REPO_ROOT" fetch origin --tags --quiet

    # Check if local main is up to date with origin/main
    LOCAL_SHA=$(git -C "$REPO_ROOT" rev-parse HEAD)
    REMOTE_SHA=$(git -C "$REPO_ROOT" rev-parse origin/main 2>/dev/null || echo "")

    if [[ -z "$REMOTE_SHA" ]]; then
        warn "Could not determine origin/main state"
    elif [[ "$LOCAL_SHA" != "$REMOTE_SHA" ]]; then
        error "Local main is not up to date with origin/main"
    else
        success "Local main is up to date with origin/main"
    fi
}

# Check that tag doesn't already exist
validate_tag() {
    header "Validating Tag"

    # Check local tags
    if git -C "$REPO_ROOT" tag -l | grep -q "^${VERSION}$"; then
        error "Tag $VERSION already exists locally"
    else
        success "Tag $VERSION does not exist locally"
    fi

    # Check remote tags
    if git -C "$REPO_ROOT" ls-remote --tags origin | grep -q "refs/tags/${VERSION}$"; then
        error "Tag $VERSION already exists on origin"
    else
        success "Tag $VERSION does not exist on origin"
    fi
}

# Validate changelog has entry for this version
validate_changelog() {
    header "Validating CHANGELOG.md"

    if [[ ! -f "$CHANGELOG_FILE" ]]; then
        error "CHANGELOG.md file not found"
        return
    fi

    # Extract version without 'v' prefix for changelog lookup
    VERSION_NUM="${VERSION#v}"

    if grep -q "^\[${VERSION_NUM}\]" "$CHANGELOG_FILE"; then
        success "Changelog entry found for version $VERSION_NUM"
    else
        error "No changelog entry found for version $VERSION_NUM"
    fi

    # Check that changelog section has content (not just headers)
    # Look for content between this version and the next version marker
    # Use sed '$d' to remove last line (POSIX-compliant, unlike head -n -1)
    SECTION=$(sed -n "/^\[${VERSION_NUM}\]/,/^\[/p" "$CHANGELOG_FILE" | sed '$d')
    CONTENT_LINES=$(echo "$SECTION" | grep -c "^-" || true)

    if [[ "$CONTENT_LINES" -eq 0 ]]; then
        error "Changelog section for $VERSION_NUM appears to be empty"
    else
        success "Changelog section has $CONTENT_LINES item(s)"
    fi
}

# Run tests and linting
validate_build() {
    header "Validating Build"

    info "Running tests..."
    if make -C "$REPO_ROOT" test > /dev/null 2>&1; then
        success "Tests passed"
    else
        error "Tests failed"
    fi

    info "Running linter..."
    if make -C "$REPO_ROOT" lint > /dev/null 2>&1; then
        success "Linting passed"
    else
        error "Linting failed"
    fi

    info "Building binary..."
    if make -C "$REPO_ROOT" build > /dev/null 2>&1; then
        success "Build succeeded"
    else
        error "Build failed"
    fi
}

# Display summary and prompt for confirmation
confirm_release() {
    header "Release Summary"

    printf "\n"
    printf "  ${B}Version:${NRM}  %s\n" "$VERSION"
    printf "  ${B}Branch:${NRM}   %s\n" "$CURRENT_BRANCH"
    printf "  ${B}Commit:${NRM}   %s\n" "$(git -C "$REPO_ROOT" rev-parse --short HEAD)"
    printf "\n"

    if [[ ${#ERRORS[@]} -gt 0 ]]; then
        printf "${B}${RED}Validation failed with %d error(s):${NRM}\n\n" "${#ERRORS[@]}"
        for err in "${ERRORS[@]}"; do
            printf "  ${RED}•${NRM} %s\n" "$err"
        done
        printf "\n"
        exit 1
    fi

    printf "${B}${GRN}All validations passed!${NRM}\n\n"

    printf "This will:\n"
    printf "  1. Create annotated tag: ${B}%s${NRM}\n" "$VERSION"
    printf "  2. Push tag to origin\n"
    printf "\n"

    read -p "Proceed with release? [y/N] " -n 1 -r
    printf "\n"

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "Release cancelled"
        exit 0
    fi
}

# Create and push the tag
create_release() {
    header "Creating Release"

    info "Creating annotated tag $VERSION..."
    git -C "$REPO_ROOT" tag -a "$VERSION" -m "Release $VERSION"
    success "Tag created"

    info "Pushing tag to origin..."
    git -C "$REPO_ROOT" push origin "$VERSION"
    success "Tag pushed"

    printf "\n${B}${GRN}Release $VERSION complete!${NRM}\n\n"
    printf "Next steps:\n"
    printf "  1. Monitor the release pipeline\n"
    printf "  2. Verify the Homebrew formula is updated\n"
    printf "  3. Test: ${B}brew update && brew upgrade kionsoftware/tap/kion-cli${NRM}\n"
    printf "\n"
}

# Main
main() {
    validate_version
    validate_git_state
    validate_tag
    validate_changelog
    validate_build
    confirm_release
    create_release
}

main "$@"
