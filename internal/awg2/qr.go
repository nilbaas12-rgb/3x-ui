package awg2

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// QRConfig represents QR code configuration
type QRConfig struct {
	Config string // WireGuard config content
	Email  string // client identifier
	Size   int    // output size in pixels
}

// GenerateQRCode generates a QR code from client configuration
func GenerateQRCode(cc string, size int) ([]byte, error) {
	if size == 0 {
		size = 400
	}

	// Generate QR code
	qrCode, err := qr.Encode(cc, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR: %w", err)
	}

	// Scale to desired size
	scaled, err := barcode.Scale(qrCode, size, size)
	if err != nil {
		return nil, fmt.Errorf("failed to scale QR: %w", err)
	}

	// Convert to PNG
	return generatePNG(scaled)
}

// GenerateQRCodeToFile generates QR code and saves to file
func GenerateQRCodeToFile(cc, outputPath string, size int) error {
	data, err := GenerateQRCode(cc, size)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outputPath, data, 0644)
}

// generatePNG converts barcode to PNG bytes
func generatePNG(img barcode.Barcode) ([]byte, error) {
	// Create temporary file
	tmpfile, err := ioutil.TempFile("", "qr-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpfile.Name())

	// Write PNG
	if err := barcode.WriteFile(tmpfile.Name(), "png", img, 0); err != nil {
		return nil, err
	}

	// Read and return
	return ioutil.ReadFile(tmpfile.Name())
}

// ConfigToQRText converts config to QR-encodable text
func ConfigToQRText(cc string) string {
	// WireGuard format is already QR-friendly
	return cc
}
