package sign

import (
	"crypto/rsa"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"CodeSigner/pkg/binary"
	"CodeSigner/pkg/crypto"

	"github.com/fatih/color"
)

// Authenticode signs a PE binary with Authenticode or creates detached signature for ELF
func Authenticode(certPath, keyPath, binaryPath, outputPath, timestampServer string) {
	// Detect binary format
	binaryFormat, err := binary.DetectFormat(binaryPath)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if binaryFormat == "ELF" {
		// Handle ELF binaries with detached signatures
		signELF(certPath, keyPath, binaryPath, outputPath)
		return
	}

	// Validate binary exists and is PE
	if err := binary.ValidatePE(binaryPath); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// Check if osslsigncode is available
	if _, err := exec.LookPath("osslsigncode"); err != nil {
		color.Red("Error: osslsigncode not found. Install it with:")
		color.Red("  Ubuntu/Debian: sudo apt-get install osslsigncode")
		color.Red("  macOS: brew install osslsigncode")
		color.Red("  Or build from: https://github.com/mtrojnar/osslsigncode")
		os.Exit(1)
	}

	// Check if using PKCS#12 (.pfx) or PEM files
	isPKCS12 := false
	var signingFile string

	if _, err := os.Stat(keyPath); err == nil && (keyPath == "appCert.pfx" || keyPath[len(keyPath)-4:] == ".pfx") {
		// Using PKCS#12 file
		isPKCS12 = true
		signingFile = keyPath
		fmt.Println("Using PKCS#12 bundle for signing...")
	} else {
		// Using separate PEM files
		if _, err := os.Stat(certPath); err != nil {
			color.Red("Error: certificate file not found: %v", err)
			os.Exit(1)
		}
		if _, err := os.Stat(keyPath); err != nil {
			color.Red("Error: private key file not found: %v", err)
			os.Exit(1)
		}
		fmt.Println("Using PEM certificate and key for signing...")
	}

	fmt.Println("Checking output path...")
	// Determine output path
	if outputPath == "" || outputPath == binaryPath {
		// Error if no output path specified
		color.Red("Error: output path must be specified with -output to avoid file conflict with the original binary")
		os.Exit(1)
	} else {
		if _, err := os.Stat(outputPath); err == nil {
			color.Yellow("Warning: output file %s already exists and will be overwritten", outputPath)
		}
		// Remove existing output file if exists and print any error
		if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
			color.Red("Error: failed to remove existing output file: %v\n", err)
			os.Exit(1)
		}
	}

	// Build osslsigncode command
	var cmd *exec.Cmd
	if isPKCS12 {
		// Use PKCS#12 format with empty password
		args := []string{"sign",
			"-pkcs12", signingFile,
			"-pass", "",
			"-in", binaryPath,
			"-out", outputPath,
		}
		// Only add timestamp server if provided and not empty
		if timestampServer != "" {
			args = append(args, "-t", timestampServer)
		}
		cmd = exec.Command("osslsigncode", args...)
	} else {
		// Use separate certificate and key files
		args := []string{"sign",
			"-certs", certPath,
			"-key", keyPath,
			"-in", binaryPath,
			"-out", outputPath,
		}
		// Only add timestamp server if provided and not empty
		if timestampServer != "" {
			args = append(args, "-t", timestampServer)
		}
		cmd = exec.Command("osslsigncode", args...)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		color.Red("Error: osslsigncode failed: %v", err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		color.Red("")
		color.Red("Troubleshooting:")
		color.Red("1. Try regenerating certificate with: go run . -mode gencert")
		color.Red("2. Then sign with PKCS#12: go run . -mode sign -key appCert.pfx -binary <file>")
		color.Red("3. Or try without timestamp: go run . -mode sign -key appCert.pfx -ts '' -binary <file>")
		os.Exit(1)
	}

	fmt.Println("")
	color.Green("✓ Binary signed with Authenticode signature")
	color.Green("  Binary: %s", binaryPath)
	color.Green("  Output: %s", outputPath)
	if isPKCS12 {
		color.Green("  Certificate: %s (PKCS#12)", signingFile)
	} else {
		color.Green("  Certificate: %s", certPath)
	}
	fmt.Println("")
}

// Verify verifies an Authenticode signature for PE or detached signature for ELF
func Verify(binaryPath string) {
	// Detect binary format
	binaryFormat, err := binary.DetectFormat(binaryPath)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if binaryFormat == "ELF" {
		// Handle ELF verification with detached signature
		verifyELF(binaryPath)
		return
	}

	// Check if osslsigncode is available
	if _, err := exec.LookPath("osslsigncode"); err != nil {
		color.Red("Error: osslsigncode not found for verification")
		os.Exit(1)
	}

	// Verify signature using osslsigncode
	cmd := exec.Command("osslsigncode", "verify", binaryPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		color.Red("✗ Signature verification failed")
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		os.Exit(1)
	}

	fmt.Println("")
	color.Green("✓ Authenticode signature verified")
	color.Green("  Binary: %s", binaryPath)
	if len(output) > 0 {
		fmt.Println("\nSignature details:")
		fmt.Println(string(output))
	}
	fmt.Println("")
}

// verifyELF verifies a detached RSA signature for ELF binaries
func verifyELF(binaryPath string) {
	fmt.Println("Detected ELF binary - verifying detached signature...")

	// Look for signature file
	sigPath := binaryPath + ".sig"
	if _, err := os.Stat(sigPath); err != nil {
		color.Red("Error: signature file not found: %s", sigPath)
		color.Red("Expected signature file: %s.sig", binaryPath)
		os.Exit(1)
	}

	// Try to find certificate for public key extraction
	certPath := "certificate.pem"
	if _, err := os.Stat(certPath); err != nil {
		color.Red("Error: certificate file not found: %s", certPath)
		color.Red("Please specify certificate with -cert flag")
		os.Exit(1)
	}

	// Load public key from certificate
	publicKey, err := crypto.ExtractPublicKeyFromCert(certPath)
	if err != nil {
		color.Red("Error: failed to extract public key from certificate: %v", err)
		os.Exit(1)
	}

	// Verify signature
	if err := crypto.VerifySignature(binaryPath, sigPath, publicKey); err != nil {
		color.Red("✗ Signature verification failed: %v", err)
		os.Exit(1)
	}

	fmt.Println("")
	color.Green("✓ ELF binary signature verified")
	color.Green("  Binary: %s", binaryPath)
	color.Green("  Signature: %s", sigPath)
	color.Green("  Algorithm: RSA-SHA256")
	fmt.Println("")
}

// signELF creates a detached RSA signature for ELF binaries
func signELF(certPath, keyPath, binaryPath, outputPath string) {
	fmt.Println("Detected ELF binary - creating detached signature...")

	// Determine signature output path
	var sigPath string
	if outputPath != "" {
		sigPath = outputPath
	} else {
		sigPath = binaryPath + ".sig"
	}

	// Load private key from PKCS#12 or PEM
	var privateKey *rsa.PrivateKey
	var err error

	if strings.HasSuffix(keyPath, ".pfx") || strings.HasSuffix(keyPath, ".p12") {
		// Extract private key from PKCS#12
		privateKey, err = crypto.ExtractPrivateKeyFromPKCS12(keyPath)
		if err != nil {
			color.Red("Error: failed to extract private key from PKCS#12: %v", err)
			os.Exit(1)
		}
		fmt.Println("Using private key from PKCS#12 bundle...")
	} else {
		// Load from PEM file
		privateKey, err = crypto.LoadPrivateKey(keyPath)
		if err != nil {
			color.Red("Error: failed to load private key: %v", err)
			os.Exit(1)
		}
		fmt.Println("Using PEM private key...")
	}

	// Sign the binary
	signature, err := crypto.Sign(binaryPath, privateKey)
	if err != nil {
		color.Red("Error: failed to sign binary: %v", err)
		os.Exit(1)
	}

	// Write signature to file
	if err := os.WriteFile(sigPath, signature, 0644); err != nil {
		color.Red("Error: failed to write signature file: %v", err)
		os.Exit(1)
	}

	fmt.Println("")
	color.Green("✓ ELF binary signed with detached RSA signature")
	color.Green("  Binary: %s", binaryPath)
	color.Green("  Signature: %s", sigPath)
	color.Green("  Algorithm: RSA-SHA256")
	fmt.Println("")
	fmt.Println("To verify the signature:")
	fmt.Printf("  go run . -mode verify -binary %s\n", binaryPath)
	fmt.Println("")
}
