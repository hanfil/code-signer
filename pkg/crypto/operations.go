package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
)

// LoadPrivateKey loads an RSA private key from the specified file path.
// The key should be in PEM format. It returns an rsa.PrivateKey and any error encountered.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load private key: %v", err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return nil, fmt.Errorf("parse PEM block: key not found")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %v", err)
	}

	privKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("type assertion to rsa.PrivateKey failed")
	}
	return privKey, nil
}

// LoadPublicKey loads an RSA public key from the specified file path.
// The key should be in PEM format. It returns an rsa.PublicKey and any error encountered.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	publicKeyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load public key: %v", err)
	}

	block, _ := pem.Decode(publicKeyData)
	if block == nil {
		return nil, fmt.Errorf("parse PEM block: key not found")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %v", err)
	}

	pubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("type assertion to rsa.PublicKey failed")
	}
	return pubKey, nil
}

// Sign creates a signature for a binary file located at binaryPath using the provided RSA private key.
// It returns the signature as a byte slice and any error encountered.
func Sign(binaryPath string, privateKey *rsa.PrivateKey) ([]byte, error) {
	binaryData, err := os.ReadFile(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("read binary data: %v", err)
	}

	hashed := sha256.Sum256(binaryData)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("sign data: %v", err)
	}
	return signature, nil
}

// VerifySignature checks the signature of a binary file located at binaryPath against a signature file at signaturePath.
// It uses the provided RSA public key for verification. It returns any error encountered in the verification process.
func VerifySignature(binaryPath, signaturePath string, publicKey *rsa.PublicKey) error {
	binaryData, err := os.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("read binary data: %v", err)
	}

	signature, err := os.ReadFile(signaturePath)
	if err != nil {
		return fmt.Errorf("read signature data: %v", err)
	}

	hashed := sha256.Sum256(binaryData)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature); err != nil {
		return fmt.Errorf("verify signature: %v", err)
	}
	return nil
}

// ExtractPublicKeyFromCert extracts the RSA public key from a certificate
func ExtractPublicKeyFromCert(certPath string) (*rsa.PublicKey, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read certificate: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %v", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate does not contain RSA public key")
	}

	return publicKey, nil
}

// ExtractPrivateKeyFromPKCS12 extracts the private key from a PKCS#12 bundle using OpenSSL
func ExtractPrivateKeyFromPKCS12(pfxPath string) (*rsa.PrivateKey, error) {
	// Use OpenSSL to extract private key from PKCS#12
	cmd := exec.Command("openssl", "pkcs12",
		"-in", pfxPath,
		"-nocerts",
		"-nodes",
		"-passin", "pass:")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("openssl extraction failed: %v", err)
	}

	// Parse the PEM-encoded private key
	block, _ := pem.Decode(output)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in output")
	}

	// Try PKCS8 format first
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := priv.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("key is not RSA")
	}

	// Try PKCS1 format
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return rsaKey, nil
}
