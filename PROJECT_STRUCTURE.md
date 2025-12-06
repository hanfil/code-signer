# Code Signer - Project Reorganization Complete

## New Project Structure

The project has been reorganized into a clean, modular structure following Go best practices:

```
code-signer/
├── main.go                     # Entry point (ONLY Go file in root)
├── go.mod                      # Go module definition
├── go.sum                       # Dependency checksums
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
├── ca-cert.pem               # Generated CA certificate
├── ca-key.pem                # Generated CA private key
├── ca-cert.srl               # Certificate serial numbers
│
├── README.md                  # Project documentation
└── FILE_GUIDE.md              # Quick reference guide
```

## Package Organization

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

---

### cmd/gencert/generate.go
**Purpose**: Certificate generation and management

**Function**: `Generate(certPath, keyPath)`
- Interactive CA root certificate generation
- Code signing certificate generation
- OpenSSL integration
- PKCS#12 bundle creation

---

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

---

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

---

### pkg/binary/format.go
**Purpose**: Binary format detection and validation

**Exported Functions**:
- `DetectFormat(binaryPath)` - Identify PE (Windows) or ELF (Linux) format
- `ValidatePE(binaryPath)` - Validate PE executable format

**Detection**:
- PE: Magic number 0x4D5A (MZ header)
- ELF: Magic number 0x7F454C46 (0x7F 'E' 'L' 'F')

---

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

## Building and Running

```bash
# Build the application
go build

# Run directly
go run .

# Run with specific mode
go run . -mode gencert
go run . -mode sign -key file.pem -binary app.exe
go run . -mode verify -binary app.exe
```



