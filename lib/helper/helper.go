// Package helper provides utility functions for the Kion CLI application.
//
// This package contains common functionality used across the CLI commands
// including user interaction, data transformation, output formatting, shell
// integration, browser operations, and configuration management. It serves as
// a central location for shared utilities that support the core CLI
// operations.
//
// Key functionalities include:
//
//   - User prompts and interactive selection wizards for projects, accounts,
//     and cloud access roles
//   - Data transformation utilities for converting API responses into
//     user-friendly formats and selection maps
//   - Output formatting for AWS credentials in various formats (environment
//     variables, credential process JSON, AWS credentials file)
//   - Shell integration for creating sub-shells with AWS credentials
//   - Browser operations for opening federated console URLs
//   - Configuration file management for loading and saving CLI settings
//
// The package uses the charmbracelet/huh library for interactive prompts with
// custom Kion brand theming, providing a consistent user experience across
// all CLI interactions.
//
// Most functions in this package follow the pattern of accepting necessary
// parameters and returning results with proper error handling, making them
// suitable for use in CLI command implementations where user interaction and
// data transformation are common requirements.
package helper
