package gencert

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// Generate creates a self-signed CA root and application certificate interactively
func Generate(certPath, keyPath string) {
	// Check if openssl is available
	if _, err := exec.LookPath("openssl"); err != nil {
		color.Red("Error: openssl not found. Install it with:")
		color.Red("  Ubuntu/Debian: sudo apt-get install openssl")
		color.Red("  macOS: brew install openssl")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("")
	color.Cyan("═══════════════════════════════════════════════════════════")
	color.Cyan("    Certificate Authority & Code Signing Certificate Setup")
	color.Cyan("═══════════════════════════════════════════════════════════")
	fmt.Println("")

	// ===== CA ROOT GENERATION =====
	fmt.Println("Step 1: CA Root Certificate")
	fmt.Println("-------------------------------")

	caKeyPath := "ca-key.pem"
	caCertPath := "ca-cert.pem"

	// Declare location variables that will be used for both CA and app cert
	var caCountry, caState, caCity string

	fmt.Print("Use existing CA? (y/n) [n]: ")
	useExistingCA, _ := reader.ReadString('\n')
	useExistingCA = strings.TrimSpace(strings.ToLower(useExistingCA))

	if useExistingCA == "y" || useExistingCA == "yes" {
		fmt.Print("CA Private Key Path [ca-key.pem]: ")
		inputCAKey, _ := reader.ReadString('\n')
		inputCAKey = strings.TrimSpace(inputCAKey)
		if inputCAKey != "" {
			caKeyPath = inputCAKey
		}

		fmt.Print("CA Certificate Path [ca-cert.pem]: ")
		inputCACert, _ := reader.ReadString('\n')
		inputCACert = strings.TrimSpace(inputCACert)
		if inputCACert != "" {
			caCertPath = inputCACert
		}

		// Verify files exist
		if _, err := os.Stat(caKeyPath); err != nil {
			color.Red("Error: CA key file not found: %s", caKeyPath)
			os.Exit(1)
		}
		if _, err := os.Stat(caCertPath); err != nil {
			color.Red("Error: CA certificate file not found: %s", caCertPath)
			os.Exit(1)
		}

		color.Green("✓ Using existing CA:")
		color.Green("  CA Key: %s", caKeyPath)
		color.Green("  CA Certificate: %s", caCertPath)
		fmt.Println("")

		// Get location info for application certificate
		fmt.Print("Country Code (2 letters) [NO]: ")
		caCountry, _ = reader.ReadString('\n')
		caCountry = strings.TrimSpace(caCountry)
		if caCountry == "" {
			caCountry = "NO"
		}

		fmt.Print("State/Province [Oslo]: ")
		caState, _ = reader.ReadString('\n')
		caState = strings.TrimSpace(caState)
		if caState == "" {
			caState = "Oslo"
		}

		fmt.Print("City [Oslo]: ")
		caCity, _ = reader.ReadString('\n')
		caCity = strings.TrimSpace(caCity)
		if caCity == "" {
			caCity = "Oslo"
		}
	} else {
		fmt.Println("Generating new CA Root Certificate...")
		fmt.Println("")

		fmt.Print("Country Code (2 letters) [NO]: ")
		caCountry, _ := reader.ReadString('\n')
		caCountry = strings.TrimSpace(caCountry)
		if caCountry == "" {
			caCountry = "NO"
		}

		fmt.Print("State/Province [Oslo]: ")
		caState, _ := reader.ReadString('\n')
		caState = strings.TrimSpace(caState)
		if caState == "" {
			caState = "Oslo"
		}

		fmt.Print("City [Oslo]: ")
		caCity, _ := reader.ReadString('\n')
		caCity = strings.TrimSpace(caCity)
		if caCity == "" {
			caCity = "Oslo"
		}

		fmt.Print("CA Organization Name [Individual Developer]: ")
		caOrg, _ := reader.ReadString('\n')
		caOrg = strings.TrimSpace(caOrg)
		if caOrg == "" {
			caOrg = "Individual Developer"
		}

		fmt.Print("CA Common Name [Filip Fog]: ")
		caCN, _ := reader.ReadString('\n')
		caCN = strings.TrimSpace(caCN)
		if caCN == "" {
			caCN = "Filip Fog"
		}

		fmt.Print("Email Address [hanfil@outlook.com]: ")
		caEmail, _ := reader.ReadString('\n')
		caEmail = strings.TrimSpace(caEmail)
		if caEmail == "" {
			caEmail = "hanfil@outlook.com"
		}

		// Collect multiple URLs/SANs
		fmt.Println("\nAdd URLs to Subject Alternative Name (press Enter on empty line to finish):")
		fmt.Println("Example: https://github.com/hanfil, https://yoursite.com, etc.")
		sanURIs := []string{}
		uriIndex := 1
		for {
			fmt.Printf("URL %d [press Enter to skip/finish]: ", uriIndex)
			sanURI, _ := reader.ReadString('\n')
			sanURI = strings.TrimSpace(sanURI)

			if sanURI == "" {
				// First URL gets default if nothing entered
				if uriIndex == 1 {
					sanURIs = append(sanURIs, "https://github.com/hanfil")
				}
				break
			}

			sanURIs = append(sanURIs, sanURI)
			uriIndex++
		}

		fmt.Println("\nGenerating CA Root Certificate...")
		fmt.Println("  Generating RSA private key (4096-bit)...")

		// Generate CA private key (4096-bit for root)
		caKeyCmd := exec.Command("openssl", "genrsa", "-out", caKeyPath, "4096")
		if output, err := caKeyCmd.CombinedOutput(); err != nil {
			color.Red("Error: failed to generate CA private key: %v", err)
			if len(output) > 0 {
				fmt.Println(string(output))
			}
			os.Exit(1)
		}

		// Create OpenSSL config file for SAN extension
		configPath := "ca-openssl.cnf"

		// Build alt_names section with email and URIs
		altNamesSection := ""
		if caEmail != "" {
			altNamesSection += fmt.Sprintf("email = %s\n", caEmail)
		}
		for i, uri := range sanURIs {
			altNamesSection += fmt.Sprintf("URI.%d = %s\n", i+1, uri)
		}

		configContent := fmt.Sprintf(`[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
prompt = no

[req_distinguished_name]
C = %s
ST = %s
L = %s
O = %s
CN = %s

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical,CA:TRUE
keyUsage = critical,keyCertSign,cRLSign
subjectAltName = @alt_names

[alt_names]
%s`, caCountry, caState, caCity, caOrg, caCN, altNamesSection)

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			color.Red("Error: failed to create OpenSSL config: %v", err)
			os.Exit(1)
		}
		defer os.Remove(configPath)

		// Generate CA self-signed certificate with SAN extension
		fmt.Println("  Generating CA self-signed certificate with URL link...")
		caCertCmd := exec.Command("openssl", "req", "-new", "-x509",
			"-key", caKeyPath,
			"-out", caCertPath,
			"-days", "7300",
			"-config", configPath)
		if output, err := caCertCmd.CombinedOutput(); err != nil {
			color.Red("Error: failed to generate CA certificate: %v", err)
			if len(output) > 0 {
				fmt.Println(string(output))
			}
			os.Exit(1)
		}

		fmt.Println("")
		color.Green("✓ CA Root Certificate Generated")
		color.Green("  CA Key: %s", caKeyPath)
		color.Green("  CA Certificate: %s (valid 20 years)", caCertPath)
	} // End of else block for generating new CA

	// ===== APPLICATION CERTIFICATE GENERATION =====
	fmt.Println("")
	fmt.Println("Step 2: Generate Code Signing Certificate (signed by CA)")
	fmt.Println("--------------------------------------------------------")

	fmt.Print("Organisation Name [Name (Individual)]: ")
	appOrg, _ := reader.ReadString('\n')
	appOrg = strings.TrimSpace(appOrg)
	if appOrg == "" {
		appOrg = "Name (Individual)"
	}

	fmt.Print("Application Name [Code Signing Certificate]: ")
	appCN, _ := reader.ReadString('\n')
	appCN = strings.TrimSpace(appCN)
	if appCN == "" {
		appCN = "Code Signing Certificate"
	}

	fmt.Print("Email Address [example@example.com]: ")
	appEmail, _ := reader.ReadString('\n')
	appEmail = strings.TrimSpace(appEmail)
	if appEmail == "" {
		appEmail = "example@example.com"
	}

	fmt.Println("\nGenerating Code Signing Certificate...")
	fmt.Println("  Generating RSA private key (2048-bit)...")

	// Generate app private key (2048-bit)
	appKeyCmd := exec.Command("openssl", "genrsa", "-out", keyPath, "2048")
	if output, err := appKeyCmd.CombinedOutput(); err != nil {
		color.Red("Error: failed to generate private key: %v", err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		os.Exit(1)
	}

	// Create certificate signing request
	fmt.Println("  Creating certificate signing request...")
	csrPath := "app.csr"
	csrCmd := exec.Command("openssl", "req", "-new",
		"-key", keyPath,
		"-out", csrPath,
		"-subj", fmt.Sprintf("/C=%s/ST=%s/L=%s/O=%s/CN=%s", caCountry, caState, caCity, appOrg, appCN))
	if output, err := csrCmd.CombinedOutput(); err != nil {
		color.Red("Error: failed to create CSR: %v", err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		os.Exit(1)
	}

	// Create OpenSSL config for app certificate with SAN
	appConfigPath := "app-openssl.cnf"
	appAltNamesSection := ""
	if appEmail != "" {
		appAltNamesSection += fmt.Sprintf("email = %s\n", appEmail)
	}

	appConfigContent := fmt.Sprintf(`[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
C = %s
ST = %s
L = %s
O = %s
CN = %s

[v3_req]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical,CA:FALSE
keyUsage = critical,digitalSignature
extendedKeyUsage = codeSigning
subjectAltName = @alt_names

[alt_names]
%s`, caCountry, caState, caCity, appOrg, appCN, appAltNamesSection)

	if err := os.WriteFile(appConfigPath, []byte(appConfigContent), 0644); err != nil {
		color.Red("Error: failed to create app OpenSSL config: %v", err)
		os.Exit(1)
	}

	// Sign CSR with CA using extension config
	fmt.Println("  Signing certificate with CA...")
	signCmd := exec.Command("openssl", "x509", "-req",
		"-in", csrPath,
		"-CA", caCertPath,
		"-CAkey", caKeyPath,
		"-CAcreateserial",
		"-out", certPath,
		"-days", "3650",
		"-sha256",
		"-extfile", appConfigPath,
		"-extensions", "v3_req")
	if output, err := signCmd.CombinedOutput(); err != nil {
		color.Red("Error: failed to sign certificate: %v", err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		os.Exit(1)
	}

	// Create PKCS#12 file for compatibility with osslsigncode
	pfxPath := "appCert.pfx"
	fmt.Println("  Creating PKCS#12 bundle with full certificate chain...")
	pfxCmd := exec.Command("openssl", "pkcs12", "-export",
		"-in", certPath,
		"-inkey", keyPath,
		"-certfile", caCertPath,
		"-out", pfxPath,
		"-passout", "pass:")
	if output, err := pfxCmd.CombinedOutput(); err != nil {
		color.Red("Error: failed to create PKCS#12 file: %v", err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		os.Exit(1)
	}

	// Clean up CSR file
	os.Remove(csrPath)

	fmt.Println("")
	color.Green("════════════════════════════════════════════════════════════")
	color.Green("✓ Certificate Generation Complete!")
	color.Green("════════════════════════════════════════════════════════════")
	fmt.Println("")

	color.Cyan("CA Root Files:")
	color.Cyan("  Private Key: %s (KEEP SECURE!)", caKeyPath)
	color.Cyan("  Certificate: %s", caCertPath)
	fmt.Println("")

	color.Green("Code Signing Files:")
	color.Green("  Private Key: %s", keyPath)
	color.Green("  Certificate: %s", certPath)
	color.Green("  PKCS#12 Bundle: %s (USE THIS FOR SIGNING)", pfxPath)
	fmt.Println("")

	fmt.Println("Next steps:")
	fmt.Println("  1. Sign your binary:")
	fmt.Printf("     go run . -mode sign -key %s -binary <your.exe>\n", pfxPath)
	fmt.Println("")
	fmt.Println("  2. Verify the signature:")
	fmt.Println("     go run . -mode verify -binary <your.exe>")
	fmt.Println("")
	fmt.Println("Security Notes:")
	fmt.Printf("  - Keep the CA private key (%s) secure!\n", caKeyPath)
	fmt.Println("  - Back up your CA certificate and key")
	fmt.Println("  - You can reuse the CA to sign multiple application certificates")
	fmt.Println("")
}
