---
inclusion: always
---

# Git Commit Standards

When I ask you to commit my code or changes:

1. Determine which files are new, modified, or deleted using `git status`
2. Stage all relevant files
3. Generate a commit message following these best practices:
   - Use the imperative mood ("Add feature" not "Added feature")
   - Start with a concise summary line (50 chars or less)
   - Structure: `<type>(<scope>): <description>`
     - Types: feat, fix, docs, style, refactor, test, chore
     - Scope: optional component name

Example:
```
feat(auth): implement OAuth2 authentication flow

- Add OAuth2 client configuration
- Create login and callback endpoints
- Store tokens securely in user session

Closes #123
```

Do not ask for confirmation before committing. Always show the generated commit message before executing.

# Message format
Kiro follows the Conventional Commits format with detailed body sections:
<type>(<scope>): <subject>
- First change or addition
- Second change or improvement
- Third change if applicable
- Why this change was needed (if relevant)

# Conventional commit types
feat: New features
fix: Bug fixes
docs: Documentation changes
style: Formatting changes
refactor: Code restructuring
test: Adding/updating tests
chore: Maintenance tasks
perf: Performance improvements
ci: CI/CD changes

# Example

feat(docs): add comprehensive Source Control documentation
- Create new documentation page for Source Control features
- Update interface documentation to link to Source Control page
- Provide detailed explanation of AI-powered commit message generation
- Describe diff context provider and commit message generation process
