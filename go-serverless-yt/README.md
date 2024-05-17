# Usage

- `build -o bootstrap cmd/main.go`
- `zip -jrm bootstrap.zip bootstrap`
- Upload `bootstrap.zip` to AWS Lambda.

## IMPORTANT

**You MUST name your build output as `bootstrap` !!!**
