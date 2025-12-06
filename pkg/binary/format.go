package binary

import (
	"fmt"
	"os"
)

// DetectFormat detects if the binary is PE (Windows) or ELF (Linux)
func DetectFormat(binaryPath string) (string, error) {
	file, err := os.Open(binaryPath)
	if err != nil {
		return "", fmt.Errorf("cannot open binary: %v", err)
	}
	defer file.Close()

	// Read first four bytes to check magic number
	header := make([]byte, 4)
	_, err = file.Read(header)
	if err != nil {
		return "", fmt.Errorf("cannot read binary header: %v", err)
	}

	// Check for PE (Windows) magic number: MZ (0x4D5A)
	if header[0] == 0x4D && header[1] == 0x5A {
		return "PE", nil
	}

	// Check for ELF (Linux) magic number: 0x7F454C46 (0x7F 'E' 'L' 'F')
	if header[0] == 0x7F && header[1] == 0x45 && header[2] == 0x4C && header[3] == 0x46 {
		return "ELF", nil
	}

	return "", fmt.Errorf("unsupported binary format (not PE or ELF)")
}

// ValidatePE checks if the binary is a valid PE (Windows) executable
func ValidatePE(binaryPath string) error {
	file, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot open binary: %v", err)
	}
	defer file.Close()

	// Read first two bytes to check PE magic number
	header := make([]byte, 2)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("cannot read binary header: %v", err)
	}

	// Check for PE (Windows) magic number: MZ (0x4D5A)
	if header[0] != 0x4D || header[1] != 0x5A {
		return fmt.Errorf("binary is not a valid PE (Windows) executable. Only PE binaries can have Authenticode signatures")
	}

	return nil
}
