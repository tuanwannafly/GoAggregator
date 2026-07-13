# Contributing to GoAggregator

Thank you for your interest in contributing to GoAggregator! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

We are committed to providing a welcoming and respectful community for all. All contributors are expected to:

- Be polite and constructive in all communications
- Treat others with respect and dignity
- Focus on what is best for the community
- Show courtesy and respect toward differing viewpoints and experiences

## Getting Started

### Fork the Repository

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/GoAggregator.git
cd GoAggregator
```

3. Add the upstream remote:

```bash
git remote add upstream https://github.com/tuanwannafly/GoAggregator.git
```

### Understanding the Project

Before making changes, familiarize yourself with:

- **Architecture**: Review the [README.md](README.md) for system architecture
- **Design System**: Review [DESIGN.md](DESIGN.md) for frontend design guidelines
- **Project Structure**: Understand the codebase organization

## Development Setup

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Node.js 18+ (for frontend development)
- Git

### Backend Setup

```bash
# Start all services with Docker
docker compose up --build

# Or run locally
go mod download
go run ./cmd/api
go run ./cmd/mockprovider
```

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific test
go test -v ./test/integration/... -run TestIntegrationChaosTwoProvidersFail
```

## Making Changes

### Branch Naming Convention

Use descriptive branch names:

- `feature/description` - New features
- `fix/description` - Bug fixes
- `refactor/description` - Code refactoring
- `docs/description` - Documentation updates
- `test/description` - Test improvements

Examples:
- `feature/add-hotel-search`
- `fix/circuit-breaker-timeout`
- `docs/update-api-reference`

### Code Style

#### Go Code

- Follow Go's standard formatting (run `go fmt`)
- Use meaningful variable and function names
- Add comments for non-obvious code
- Keep functions focused and small
- Write tests for new functionality

```go
// Good: Clear function name and purpose
func calculateFlightPrice(route Route, date time.Time) (Money, error)

// Bad: Unclear naming
func calc(r string, t time.Time) (Money, error)
```

#### TypeScript/React Code

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Keep components small and focused
- Use the design system components

```typescript
// Good: Descriptive names and proper typing
interface FlightSearchFormProps {
  onSearch: (criteria: SearchCriteria) => void;
  isLoading: boolean;
}

// Bad: Unclear props
interface Props {
  onSearch: (criteria: any) => void;
  loading: boolean;
}
```

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, no code change
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Examples:

```
feat(frontend): add flight search form component

- Added FlightSearchForm with origin, destination, and date inputs
- Integrated with backend API
- Added loading and error states

Closes #123
```

```
fix(breaker): resolve circuit breaker race condition

The circuit breaker state machine had a race condition when multiple
goroutines attempted to transition simultaneously. Added mutex lock
to ensure thread-safe state transitions.
```

## Testing

### Writing Tests

#### Go Tests

- Unit tests for business logic
- Integration tests for API endpoints
- Chaos tests for resilience patterns

```go
func TestFlightSearch(t *testing.T) {
    // Test implementation
}

func TestIntegrationChaosTwoProvidersFail(t *testing.T) {
    // Integration test implementation
}
```

#### Frontend Tests

```bash
# Run frontend tests
cd frontend
npm test
```

### Running the Demo Scripts

Test your changes with the demo scripts:

```bash
# Circuit breaker demo
./scripts/circuit-breaker-demo.sh

# Chaos engineering demo
./scripts/chaos-demo.sh

# Timeout isolation demo
./scripts/timeout-demo.sh

# Rate limiter demo
./scripts/rate-limiter-demo.sh
```

## Submitting Changes

### Pull Request Process

1. **Create a branch** from `main`:

```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes** and commit them:

```bash
git add .
git commit -m "feat(scope): description of changes"
```

3. **Push to your fork**:

```bash
git push origin feature/your-feature-name
```

4. **Open a Pull Request** on GitHub:
   - Use a clear title
   - Describe the changes
   - Reference any related issues
   - Ensure all tests pass

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
```

### Code Review

- Respond to review comments promptly
- Be open to feedback and suggestions
- Ask clarifying questions if needed
- Keep discussions constructive and focused

## Reporting Issues

### Bug Reports

Include the following information:

- Description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Relevant logs or error messages

### Feature Requests

- Clear description of the feature
- Use case or motivation
- Potential alternatives considered

### Security Issues

For security vulnerabilities, please DO NOT open a public issue. Contact the maintainers directly via email or through GitHub's security advisories.

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Docker Documentation](https://docs.docker.com/)

## License

By contributing to GoAggregator, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to:

- Open an issue for questions
- Join project discussions
- Contact the maintainers

Thank you for contributing!
