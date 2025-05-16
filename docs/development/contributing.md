---
title: Contributing Guide
description: Guidelines for contributing to the Fetchfy MCP Gateway Operator
---

# Contributing Guide

Thank you for considering contributing to the Fetchfy MCP Gateway Operator! This document provides guidelines and instructions to help you contribute effectively.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct, which ensures a welcoming and inclusive environment for all contributors.

## Getting Started

### Setting Up Your Development Environment

Before you start contributing, make sure you have set up your development environment by following the [Development Setup Guide](./setup.md).

### Finding Issues to Work On

- Check the [GitHub issues](https://github.com/fetchfy/fetchfy-operator/issues) for tasks labeled as `good first issue` or `help wanted`
- Feel free to ask questions in the issues if you need clarification
- Comment on an issue to let others know you're working on it

## Contribution Workflow

### 1. Fork the Repository

Start by forking the repository to your GitHub account.

### 2. Create a Branch

Create a branch in your fork with a descriptive name related to the issue you're addressing:

```bash
git checkout -b feature/add-new-feature
# or
git checkout -b fix/issue-description
```

### 3. Make Your Changes

- Make your changes following the coding conventions (see below)
- Write or update tests as necessary
- Update documentation to reflect your changes

### 4. Test Your Changes

- Run unit tests: `make test`
- Run integration tests: `make test-integration`
- Verify the operator works in your local environment: `make run`

### 5. Commit Your Changes

Write clear, concise commit messages that explain the changes you've made:

```bash
git commit -m "Add feature: brief description of what was added"
# or
git commit -m "Fix: description of what was fixed"
```

### 6. Push to Your Fork

Push your changes to your fork on GitHub:

```bash
git push origin your-branch-name
```

### 7. Create a Pull Request

- Go to the original repository and create a pull request from your branch
- Fill in the PR template with all relevant information
- Link any related issues using keywords like "Fixes #123" or "Addresses #456"

### 8. Code Review

- Maintainers will review your PR and may request changes or ask questions
- Address any feedback by making additional commits to your branch
- Once approved, your PR will be merged into the main codebase

## Coding Conventions

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `golangci-lint run` before submitting code

### Test Coverage

- Aim for high test coverage on all new code
- Write both unit tests and integration tests where appropriate
- Use table-driven tests for functions with multiple input/output combinations

### Documentation

- Document all public functions and types
- Update relevant documentation when adding features or making changes
- Use clear, concise language in comments and documentation

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or fewer
- Reference issues and pull requests after the first line

## Pull Request Guidelines

### PR Title and Description

- Use a clear, descriptive title
- Include a summary of what changed and why
- Reference related issues
- Mention any breaking changes

### PR Size

- Keep PRs focused on a single issue or feature
- Break down large changes into smaller, manageable PRs
- If a change affects multiple components, consider splitting it into separate PRs

### PR Checklist

Before submitting a PR, ensure:

- [ ] Code builds without errors
- [ ] All tests pass
- [ ] New code has appropriate test coverage
- [ ] Documentation is updated
- [ ] Code follows the project's style guidelines
- [ ] Commit messages follow the guidelines

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

### Creating a Release

Only project maintainers can create releases:

1. Update the version in relevant files
2. Update the changelog with all notable changes
3. Create a new git tag with the version number
4. Push the tag to GitHub
5. GitHub Actions will build and publish the release

## Community

- Join our [Slack channel](#) for discussions
- Participate in community meetings (scheduled on our [community calendar](#))
- Follow our [Twitter account](#) for updates

## Additional Resources

- [Development Setup Guide](./setup.md)
- [Debugging Guide](./debugging.md)
- [API Reference](../api-reference/gateway-crd.md)

Thank you for contributing to the Fetchfy MCP Gateway Operator!
