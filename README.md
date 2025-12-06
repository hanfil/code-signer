# Binary Signing Tool

A utility for signing PE (Windows) executables with Authenticode signatures from Linux, so they can be verified as signed with `Get-AuthenticodeSignature` in Windows PowerShell.

## Features

- **Authenticode signing**: Embed Windows-compatible signatures directly in PE binaries
- **Cross-platform signing**: Sign Windows binaries from Linux
- **Timestamp server support**: Include trusted timestamps in signatures
- **Binary format validation**: Automatically validates PE (MZ header) format
- **OpenSSL-based**: Uses `osslsigncode` for RFC3161-compliant Authenticode signatures

## Prerequisites

- `osslsigncode` (for creating Authenticode signatures)
- Optional: Self-signed certificate and private key (PEM format)
- Optional: Timestamp server for trusted timestamps

## Installation

### Install osslsigncode

**Ubuntu/Debian:**
```bash
sudo apt-get install osslsigncode
```

## Usage

### Generating Certificates (Recommended)

The easiest way is to use the built-in certificate generator:

```bash
./CodeSigner -mode gencert
```

This interactive tool:
1. Prompts whether to use an existing CA or create a new one
2. If creating new CA: Generates a Certificate Authority (CA) root certificate with:
   - Configurable location (default: Oslo, Norway)
   - Email address in Subject Alternative Name (SAN)
   - Multiple URLs/URIs in SAN (e.g., GitHub profile)
   - 4096-bit RSA key, valid for 20 years
3. Generates an application code-signing certificate:
   - Organization and application name
   - Email address in SAN
   - 2048-bit RSA key, valid for 10 years
   - Signed by the CA

**Files generated:**
- `ca-key.pem` - CA private key (4096-bit RSA) - KEEP SECURE!
- `ca-cert.pem` - CA certificate (valid 20 years)
- `private_key.pem` - Application private key (2048-bit RSA)
- `certificate.pem` - Application certificate (valid 10 years)
- `signing.pfx` - PKCS#12 bundle with full certificate chain (USE THIS FOR SIGNING)

**Reusing an existing CA:**

If you already have a CA and want to generate additional application certificates:

```bash
./CodeSigner -mode gencert
# When prompted "Use existing CA? (y/n) [n]:" type "y"
# Provide paths to your existing ca-key.pem and ca-cert.pem
```

This allows you to maintain a single CA and sign multiple applications with it.

### Manual Certificate Generation (Alternative)

If you prefer to generate certificates manually with OpenSSL:

```bash
# Generate private key (2048-bit RSA)
openssl genrsa -out private_key.pem 2048

# Create self-signed certificate (valid 3650 days = 10 years)
openssl req -new -x509 -key private_key.pem -out certificate.pem -days 3650 \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=subfuz.local"
```

**For production use**, obtain a code-signing certificate from a trusted Certificate Authority (CA) instead of using self-signed certificates.

### Signing a Windows Binary

Sign a PE executable with an Authenticode signature using the PKCS#12 bundle:

```bash
# Using PKCS#12 bundle (recommended, most compatible)
./CodeSigner -mode sign -key subfuz.pfx -binary ./subfuz.exe -output ./subfuz-signed.exe

# Or using separate PEM files
./CodeSigner -mode sign -cert certificate.pem -key private_key.pem -binary ./subfuz.exe -output ./subfuz-signed.exe
```

**Note:** The `-output` flag is required to prevent accidentally overwriting your original unsigned binary.

**Command-line flags:**
- `-mode sign` - Sign the binary (default)
- `-cert <path>` - Path to certificate file (default: `certificate.pem`, ignored if using .pfx)
- `-key <path>` - Path to private key file OR PKCS#12 bundle (default: `private_key.pem` or `subfuz.pfx`)
- `-binary <path>` - Path to the binary to sign (required)
- `-output <path>` - Output path for signed binary (required to avoid overwriting original)
- `-ts <url>` - Timestamp server URL (optional, leave empty to skip timestamp)

### Verifying a Signature (Linux-side)

Verify an Authenticode signature using osslsigncode:

```bash
./CodeSigner -mode verify -binary ./subfuz-signed.exe
```

### Verifying a Signature (Windows-side)

Once signed, verify the signature in Windows PowerShell:

```powershell
# Check if signed
Get-AuthenticodeSignature .\subfuz-signed.exe

# Expected output for self-signed:
# Status        : UnknownError
# IsOSBinary    : False
# IsMssigned    : False
# SignerCertificate : [Certificate details]
# SignatureType     : Authenticode
# SignatureAlgorithm : sha256RSA
```

**Note**: Self-signed certificates show "UnknownError" because they're not trusted by Windows. For production, use a CA-signed certificate.

## Supported Binary Formats

Currently only PE (Windows) binaries with Authenticode signatures are supported:

| Format | Extension | Platform |
|--------|-----------|----------|
| PE     | .exe, .dll | Windows |

ELF (Linux) binaries cannot use Authenticode signatures—they use different signing mechanisms (e.g., gpg signatures). When signing an ELF binary, the tool will generate a separate `.sig` signature file instead.

## Cryptographic Details

- **Signature Algorithm**: RSA with SHA-256
- **CA Key Size**: 4096-bit RSA
- **Application Key Size**: 2048-bit RSA
- **Certificate Chain**: CA root → Application certificate → PKCS#12 bundle
- **Standard**: Authenticode (RFC 3161), X.509 certificates
- **Extensions**: Subject Alternative Name (SAN) with email and URI fields (RFC 5280)
- **Validity**: CA 20 years, Application 10 years
- **Timestamp Authority**: Optional (recommended for production)
- **Signature embedding**: Embedded in PE certificate table

## Timestamp Servers

Using a trusted timestamp server ensures the signature remains valid even after the certificate expires. However, timestamp servers can be unreliable or cause issues on some systems.

**To add a timestamp server during signing:**

```bash
# With a timestamp server (optional)
./CodeSigner -mode sign -key subfuz.pfx -ts "http://timestamp.globalsign.com/scripts/timstamp.dll" -binary subfuz.exe -output subfuz-signed.exe
```

**Popular timestamp servers:**
- GlobalSign: `http://timestamp.globalsign.com/scripts/timstamp.dll`
- Sectigo: `http://timestamp.sectigo.com`
- DigiCert: `http://timestamp.digicert.com`

**Troubleshooting timestamp issues:**
- If signing fails with timestamp server, try signing without it (default behavior)
- Timestamp servers may be rate-limited or temporarily unavailable
- Self-signed certificates don't require timestamps

## Error Handling

The tool provides clear error messages for:
- Missing osslsigncode installation
- Invalid PE binaries
- Missing or invalid certificate/key files
- Timestamp server errors
- File I/O errors

Exit code `0` indicates success. Any error results in exit code `1`.

## Building as a Standalone Tool

To compile the signer as a standalone executable:

```bash
# Build for Linux
go build -o CodeSigner .
```

Then use without `go run`:

```bash
./CodeSigner -mode sign -cert certificate.pem -key private_key.pem -binary subfuz.exe
./CodeSigner -mode verify -binary subfuz.exe
```

## Troubleshooting

**Error: "osslsigncode not found"**
- Install osslsigncode for your platform (see Installation section)
- Verify it's in your PATH: `which osslsigncode`

**Error: "osslsigncode failed: signal: segmentation fault"**
- This usually means the certificate/key format is incompatible with osslsigncode
- **The timestamp server can also cause this issue**
- **Solution**: Use the PKCS#12 bundle without timestamp:
  ```bash
./CodeSigner -mode gencert
  ./CodeSigner -mode sign -key subfuz.pfx -binary subfuz.exe -output subfuz-signed.exe
  ```
- If you need timestamps, try adding after signing works:
  ```bash
  ./CodeSigner -mode sign -key subfuz.pfx -ts "http://timestamp.globalsign.com/scripts/timstamp.dll" -binary subfuz.exe -output subfuz-signed.exe
  ```

**Error: "certificate file not found" or "private key file not found"**
- Generate them automatically:
  ```bash
  ./CodeSigner -mode gencert
  ```
- Or check that files exist in current directory: `ls -la *.pem *.pfx`
- Or specify full paths:
  ```bash
  ./CodeSigner -mode sign -cert /full/path/certificate.pem -key /full/path/private_key.pem -binary subfuz.exe -output subfuz-signed.exe
  ```

**Error: "binary is not a valid PE (Windows) executable"**
- Ensure you're signing a compiled Windows .exe, not a source file
- Verify the binary wasn't corrupted: `file subfuz.exe`

**Windows shows "NotSigned"**
- The binary may not have been signed correctly
- Try verifying on Linux first: `./CodeSigner -mode verify -binary subfuz-signed.exe`
- Rebuild and sign again:
  ```bash
  ./CodeSigner -mode gencert
  ./CodeSigner -mode sign -key subfuz.pfx -binary subfuz.exe -output subfuz-signed.exe
  ```

**Windows shows "UnknownError" for signature status**
- This is expected for self-signed certificates
- Windows doesn't trust self-signed certificates by default
- For production, use a certificate from a trusted CA (see Certificate Options section)

## Security Notes

- **Keep private keys secure**: Never share or commit private keys to version control (especially `ca-key.pem`!)
- **CA certificate reuse**: You can reuse the same CA to sign multiple applications - this is the recommended approach
- **Self-signed limitations**: Self-signed certificates aren't trusted by Windows, but signatures are still valid cryptographically
- **Production certificates**: For distribution, obtain code-signing certificates from trusted CAs (DigiCert, Sectigo, GlobalSign, etc.)
- **Key protection**: Encrypt private keys with passphrases and store securely
- **Signature verification**: Always verify signatures are present before distributing
- **Certificate information**: Email addresses and URLs in certificates are visible to anyone who inspects them

## Certificate Options

### Self-Signed (Development)
```bash
openssl req -new -x509 -key private_key.pem -out certificate.pem -days 3650 \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=subfuz.local"
```

### From a Certificate Authority (Production)
1. Generate a Certificate Signing Request (CSR):
```bash
openssl req -new -key private_key.pem -out certificate.csr \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourcompany.com"
```

2. Submit the CSR to a Certificate Authority and download the signed certificate

3. Use the CA-signed certificate:
```bash
go run . -mode sign \
  -cert ca-signed-certificate.pem \
  -key private_key.pem \
  -binary subfuz.exe
```

## Contributing

For project structure, development workflow, detailed flags, and advanced guidance, see `CONTRIBUTING.md`.

## License

Apache License 2.0. See `LICENSE`.