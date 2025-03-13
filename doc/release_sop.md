Release SOP
===========

1. Bump `VERSION.md` file according to [semver](https://semver.org/) practices
2. Create section in `CHANGELOG.md` that matches the new version
    - Write release paragraph at the top of the new release section in changelog
    - Ensure all non-backend, non-administrative PR's are accounted for in subsections (changed, added, etc)
    - Stub out new placeholder for next release at top of changelog
3. Push and merge into main
4. Checkout main locally, pull from upstream
5. Tag with the release version and push
    ```bash
    git tag -a v0.5.0 -m "Release v0.5.0"
    git push origin v0.5.0
    ```
6. Monitor release pipeline and ensure it passes, along with the brew repo pipeline
7. If there is an issue with a pipeline
    ```bash
    # pull tag, fix an issue, re-tag (tag v0.5.0 for example)

    git push origin :refs/tags/v0.5.0        # removes the tag upstream

    # do git changes and git push origin [branch] then pr back into master

    git tag -fa v0.5.0 -m "Release v0.5.0"   # force a move / retag
    git push origin v0.5.0
    ```
8. Validate the release
    - Homebrew repo shows new commit and the formula has been updated
    - Cli repo shows new release and release notes look correct
    - Upgrade locally via brew `brew update && brew upgrade kionsoftware/tap/kion-cli && kion --version`
