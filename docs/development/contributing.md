# Contributing

## Development Process

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Run tests
5. Submit a pull request

## Code Style

### Go

- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Follow standard Go conventions

### TypeScript

- Use TypeScript strict mode
- Use functional components with hooks
- Keep components focused and small

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add evaluation history panel
fix: resolve spec save race condition
docs: update installation guide
refactor: extract API client to service
```

## Pull Requests

- Keep PRs focused on a single change
- Include tests for new functionality
- Update documentation as needed
- Ensure CI passes

## Project Structure

### Adding a New Component

1. Create component in `desktop/renderer/src/components/`
2. Export from `components/index.ts`
3. Add types to `types/index.ts` if needed

### Adding a New API Endpoint

1. Add handler in `cmd/daemon/main.go`
2. Add types in `pkg/api/types.go`
3. Add client method in `desktop/renderer/src/services/api.ts`
