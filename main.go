package main

import (
	"flag"
	"os"

	"CodeSigner/cmd/gencert"
	"CodeSigner/cmd/sign"

	"github.com/fatih/color"
)

func main() {
	// Parse command-line arguments
	mode := flag.String("mode", "sign", "Operation mode: sign, verify, or gencert")
	certPath := flag.String("cert", "certificate.pem", "Path to certificate file (PEM format)")
	keyPath := flag.String("key", "private_key.pem", "Path to private key file (PEM format)")
	binaryPath := flag.String("binary", "", "Path to the binary to be signed/verified")
	outputPath := flag.String("output", "", "Output path for signed binary (default: overwrites input)")
	timestampServer := flag.String("ts", "", "Timestamp server URL (optional, leave empty for no timestamp)")
	flag.Parse()

	switch *mode {
	case "gencert":
		gencert.Generate(*certPath, *keyPath)
	case "sign":
		if *binaryPath == "" {
			color.Red("Error: -binary flag is required for sign mode")
			os.Exit(1)
		}
		sign.Authenticode(*certPath, *keyPath, *binaryPath, *outputPath, *timestampServer)
	case "verify":
		if *binaryPath == "" {
			color.Red("Error: -binary flag is required for verify mode")
			os.Exit(1)
		}
		sign.Verify(*binaryPath)
	default:
		color.Red("Error: invalid mode '%s'. Use 'gencert', 'sign', or 'verify'", *mode)
		os.Exit(1)
	}
}
