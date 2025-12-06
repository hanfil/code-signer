# Contributing to Code Signer

Thank you for your interest in contributing to Code Signer! This document provides guidance on how to work with the codebase and understand the project structure.

## Project Structure

Code Signer is organized into a clean, modular structure following Go best practices:

```
code-signer/
├── main.go                     # Entry point (ONLY Go file in root)
├── go.mod                      # Go module definition
│
├── cmd/                        # Executable commands
│   ├── gencert/
│   │   └── generate.go        # Certificate generation logic
│   │
│   └── sign/
│       └── handler.go         # Signing and verification logic
│
├── pkg/                        # Reusable packages
│   ├── binary/
│   │   └── format.go          # Binary format detection
│   │
│   └── crypto/
│       └── operations.go      # Cryptographic operations
│
│
├── README.md                  # Project documentation
└── CONTRIBUTING.md            # Contribution guidelines
```

## Code Organization

### main.go (Root Level)
**Purpose**: Application entry point and CLI dispatcher

**Content**:
- Flag parsing for all modes (gencert, sign, verify)
- Mode-based command routing
- Minimal responsibility

**Imports**:
```go
import (
    "CodeSigner/cmd/gencert"
    "CodeSigner/cmd/sign"
    "github.com/fatih/color"
)
```

### cmd/gencert/generate.go
**Purpose**: Certificate generation and management

**Function**: `Generate(certPath, keyPath)`
- Interactive CA root certificate generation
- Code signing certificate generation
- OpenSSL integration
- PKCS#12 bundle creation

### cmd/sign/handler.go
**Purpose**: Binary signing and verification

**Functions**:
- `Authenticode(certPath, keyPath, binaryPath, outputPath, timestampServer)` - Sign PE binaries
- `Verify(binaryPath)` - Verify signatures
- `signELF()` - Sign ELF binaries (private)
- `verifyELF()` - Verify ELF signatures (private)

**Imports**:
```go
import (
    "CodeSigner/pkg/binary"
    "CodeSigner/pkg/crypto"
)
```

### pkg/crypto/operations.go
**Purpose**: Cryptographic operations library

**Exported Functions**:
- `LoadPrivateKey(path)` - Load RSA private key from PEM
- `LoadPublicKey(path)` - Load RSA public key from PEM
- `Sign(binaryPath, privateKey)` - Create RSA-SHA256 signature
- `VerifySignature(binaryPath, signaturePath, publicKey)` - Verify signature
- `ExtractPublicKeyFromCert(certPath)` - Extract public key from certificate
- `ExtractPrivateKeyFromPKCS12(pfxPath)` - Extract private key from PKCS#12

**Key Feature**: Can be imported and used as a standalone cryptography library

### pkg/binary/format.go
**Purpose**: Binary format detection and validation

**Exported Functions**:
- `DetectFormat(binaryPath)` - Identify PE (Windows) or ELF (Linux) format
- `ValidatePE(binaryPath)` - Validate PE executable format

**Detection**:
- PE: Magic number 0x4D5A (MZ header)
- ELF: Magic number 0x7F454C46 (0x7F 'E' 'L' 'F')

## Import Structure

```
main.go (root)
    ↓
    ├→ cmd/gencert
    │   └→ github.com/fatih/color
    │
    └→ cmd/sign
        ├→ pkg/binary
        │   └→ (std lib only)
        │
        ├→ pkg/crypto
        │   └→ (std lib + os/exec)
        │
        └→ github.com/fatih/color
```

## Getting Started

### Prerequisites
- Go 1.16 or higher
- OpenSSL (for certificate generation)
- Git

### Building the Project

```bash
# Clone the repository
git clone https://github.com/hanfil/code-signer.git
cd code-signer

# Install dependencies
go mod download

# Build the application
go build

# Run tests (if available)
go test ./...
```

## Running the Application

```bash
# Run directly
go run .

# Run with specific mode
go run . -mode gencert
go run . -mode sign -key file.pem -binary app.exe
go run . -mode verify -binary app.exe

# Build and run the compiled binary
./CodeSigner -mode gencert
```

## Making Changes

### Code Style
- Follow Go conventions and best practices
- Use `gofmt` for consistent formatting
- Keep functions focused and well-documented
- Add comments for exported functions and complex logic

### Adding Features
1. Create features in appropriate packages (`cmd/` for commands, `pkg/` for libraries)
2. Keep dependencies minimal
3. Update this CONTRIBUTING.md if structure changes
4. Ensure backward compatibility when possible

### Testing
Before submitting changes:
- Test locally with `go test ./...`
- Test the built binary with various inputs
- Verify both Windows (PE) and Linux (ELF) compatibility where applicable

## Submitting Contributions

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/your-feature`)
3. **Make your changes** following the code style guidelines
4. **Test thoroughly** 
5. **Commit with clear messages** (`git commit -m 'Add feature description'`)
6. **Push** to your fork (`git push origin feature/your-feature`)
7. **Submit a Pull Request** with a clear description of changes

## Code Review Process

- All submissions will be reviewed for:
  - Code quality and style
  - Adherence to project structure
  - Performance implications
  - Security considerations
  - Documentation completeness

## License

By contributing to Code Signer, you agree that your contributions will be licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

## Questions?

If you have questions about contributing or the codebase structure, feel free to open an issue or discussion on GitHub.
