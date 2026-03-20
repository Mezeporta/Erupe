# Contributing to Erupe

Thank you for your interest in contributing to Erupe! This guide will help you get started.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- Git

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/Erupe.git
   cd Erupe
   ```

3. Set up the database following the [Installation guide](README.md#installation)
4. Copy `config.example.json` to `config.json` and set your database password (see `config.reference.json` for all available options)
5. Install dependencies:

   ```bash
   go mod download
   ```

6. Build and run:

   ```bash
   go build
   ./erupe-ce
   ```

## Code Contribution Workflow

1. **Create a branch** for your changes:

   ```bash
   git checkout -b feature/your-feature-name
   ```

   Use descriptive branch names:
   - `feature/` for new features
   - `fix/` for bug fixes
   - `refactor/` for code refactoring
   - `docs/` for documentation changes

2. **Make your changes** and commit them with clear, descriptive messages:

   ```bash
   git commit -m "feat: add new quest loading system"
   git commit -m "fix: resolve database connection timeout"
   git commit -m "docs: update configuration examples"
   ```

3. **Test your changes** (see [Testing Requirements](#testing-requirements))

4. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request** on GitHub with:
   - Clear description of what changes you made
   - Why the changes are needed
   - Any related issue numbers

6. **Respond to code review feedback** promptly

## Coding Standards

### Go Style

- Run `gofmt` before committing:

  ```bash
  gofmt -w .
  ```

- Use `golangci-lint` for linting:

  ```bash
  golangci-lint run ./...
  ```

- Follow standard Go naming conventions
- Keep functions focused and reasonably sized
- Add comments for exported functions and complex logic
- Handle errors explicitly (don't ignore them)

### Code Organization

- Place new handlers in appropriate files under `server/channelserver/`
- Keep database queries in structured locations
- Use the existing pattern for message handlers

## Testing Requirements

Before submitting a pull request:

1. **Run all tests**:

   ```bash
   go test -v ./...
   ```

2. **Check for race conditions**:

   ```bash
   go test -v -race ./...
   ```

3. **Ensure your code has adequate test coverage**:

   ```bash
   go test -v -cover ./...
   ```

### Writing Tests

- Add tests for new features in `*_test.go` files
- Test edge cases and error conditions
- Use table-driven tests for multiple scenarios
- Mock external dependencies where appropriate

Example:

```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name string
        input int
        want int
    }{
        {"basic case", 1, 2},
        {"edge case", 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := YourFunction(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Database Schema Changes

Erupe uses an embedded auto-migrating schema system in `server/migrations/`.

When adding schema changes:

1. Create a new file in `server/migrations/sql/` with format: `NNNN_description.sql` (e.g. `0002_add_new_table.sql`)
2. Increment the number from the last migration
3. Test the migration on both a fresh and existing database
4. Document what the migration does in SQL comments

Migrations run automatically on startup in order. Each runs in its own transaction and is tracked in the `schema_version` table.

For seed/demo data (shops, events, gacha), add files to `server/migrations/seed/`. Seed data is applied automatically on fresh databases and can be re-applied via the setup wizard.

## Documentation Requirements

### Always Update

- **[CHANGELOG.md](CHANGELOG.md)**: Document your changes under "Unreleased" section
  - Use categories: Added, Changed, Fixed, Removed, Security
  - Be specific about what changed and why

### When Applicable

- **[README.md](README.md)**: Update if you change:
  - Installation steps
  - Configuration options
  - Requirements
  - Usage instructions

- **Code Comments**: Add or update comments for:
  - Exported functions and types
  - Complex algorithms
  - Non-obvious business logic
  - Packet structures and handling

## Getting Help

### Questions and Discussion

- **[Mogapedia's Discord](https://discord.gg/f77VwBX5w7)**: Active development discussions
- **[Mezeporta Square Discord](https://discord.gg/DnwcpXM488)**: Community support
- **GitHub Issues**: For bug reports and feature requests

### Reporting Bugs

When filing a bug report, include:

1. **Erupe version** (git commit hash or release version)
2. **Client version** (ClientMode setting)
3. **Go version**: `go version`
4. **PostgreSQL version**: `psql --version`
5. **Steps to reproduce** the issue
6. **Expected behavior** vs actual behavior
7. **Relevant logs** (enable debug logging if needed)
8. **Configuration** (sanitize passwords!)

### Requesting Features

For feature requests:

1. Check existing issues first
2. Describe the feature and its use case
3. Explain why it would benefit the project
4. Be open to discussion about implementation

## Code of Conduct

- Be respectful and constructive
- Welcome newcomers and help them learn
- Focus on the code, not the person
- Assume good intentions

## License

By contributing to Erupe, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Erupe!
