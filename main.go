package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/term"
)

const (
	fileExt = ".fox"
	saltLen = 16
	keyLen = 32
	formatVer = 0x01 // First byte of file = version
	// Argon2id params - tune so this takes ~500ms on your machine
	argonTime = 3
	argonMemory = 256 * 1024 // 256 MB
	argonThreads = 4
)

var (
	version = "dev" // overwritten by -ldflags at build time
	
	ErrInvalidFile = errors.New("invalid or corrupted file")
	ErrWrongPassword = errors.New("decryption failed: wrong password or corrupted file")
	ErrShortFile = errors.New("file too short")
	ErrInvalidArgs = errors.New("invalid arguments")
)

// Exit codes per README
const (
	ExitSuccess = 0
	ExitGeneral = 1
	ExitUsage = 2
	ExitDecrypt = 3
)

func deriveKey(password []byte, salt []byte) []byte {
	defer func() { // Clear password from memory
		for i := range password {
			password[i] = 0
		}
	}()
	return argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, keyLen)
}

func readPassword(prompt string) ([]byte, error) {
	fmt.Fprint(os.Stderr, prompt)
	pass, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err!= nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if len(pass) == 0 {
		return nil, errors.New("password cannot be empty")
	}
	return pass, nil
}

func encrypt(filename string) error {
	data, err := os.ReadFile(filename)
	if err!= nil {
		return fmt.Errorf("read source failed: %w", err)
	}

	outName := filename + fileExt
	if _, err := os.Stat(outName); err == nil {
		return fmt.Errorf("output file %s already exists", outName)
	}

	password, err := readPassword("Set password: ")
	if err!= nil {
		return err
	}
	confirm, err := readPassword("Confirm password: ")
	if err!= nil {
		return err
	}
	if string(password)!= string(confirm) {
		for i := range password { password[i] = 0 }
		for i := range confirm { confirm[i] = 0 }
		return errors.New("passwords do not match")
	}
	for i := range confirm { confirm[i] = 0 }

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err!= nil {
		return fmt.Errorf("salt generation failed: %w", err)
	}

	key := deriveKey(password, salt)
	defer func() { for i := range key { key[i] = 0 } }()

	aead, err := chacha20poly1305.NewX(key)
	if err!= nil {
		return fmt.Errorf("cipher init failed: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err!= nil {
		return fmt.Errorf("nonce generation failed: %w", err)
	}

	// Format: version | salt | nonce | ciphertext
	output := make([]byte, 1+len(salt)+len(nonce))
	output[0] = formatVer
	copy(output[1:], salt)
	copy(output[1+len(salt):], nonce)

	encrypted := aead.Seal(output, nonce, data, nil)

	if err := os.WriteFile(outName, encrypted, 0600); err!= nil {
		return fmt.Errorf("write failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✔ Encrypted: %s\n", outName)
	return nil
}

func decrypt(filename string) error {
	if len(filename) < len(fileExt) || filename[len(filename)-len(fileExt):]!= fileExt {
		return fmt.Errorf("%w: not a.fox file", ErrInvalidArgs)
	}

	raw, err := os.ReadFile(filename)
	if err!= nil {
		return fmt.Errorf("read failed: %w", err)
	}

	if len(raw) < 1+saltLen+chacha20poly1305.NonceSizeX {
		return ErrShortFile
	}

	if raw[0]!= formatVer {
		return fmt.Errorf("%w: unsupported version 0x%x", ErrInvalidFile, raw[0])
	}

	salt := raw[1 : 1+saltLen]
	nonce := raw[1+saltLen : 1+saltLen+chacha20poly1305.NonceSizeX]
	ciphertext := raw[1+saltLen+chacha20poly1305.NonceSizeX:]

	password, err := readPassword("Password: ")
	if err!= nil {
		return err
	}

	key := deriveKey(password, salt)
	defer func() { for i := range key { key[i] = 0 } }()

	aead, err := chacha20poly1305.NewX(key)
	if err!= nil {
		return fmt.Errorf("cipher init failed: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err!= nil {
		return ErrWrongPassword
	}
	defer func() { for i := range plaintext { plaintext[i] = 0 } }()

	// Just print to stdout. Pipe to less, cat, etc if needed
	os.Stdout.Write(plaintext)
	return nil
}

func main() {
	// Handle -v before flag.Parse so it works without -f
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("fox-vault %s\n", version)
		os.Exit(ExitSuccess)
	}

	mode := flag.String("m", "", "Mode: lock or view")
	file := flag.String("f", "", "Target file")
	flag.Parse()

	if *file == "" || *mode == "" {
		fmt.Fprintln(os.Stderr, "Usage: fox -m [lock|view] -f <filename>")
		fmt.Fprintln(os.Stderr, " fox -v")
		os.Exit(ExitUsage)
	}

	var err error
	switch *mode {
	case "lock":
		err = encrypt(*file)
	case "view":
		err = decrypt(*file)
	default:
		fmt.Fprintln(os.Stderr, "Invalid mode. Use -m lock or -m view")
		os.Exit(ExitUsage)
	}

	if err!= nil {
		fmt.Fprintf(os.Stderr, "✘ Error: %v\n", err)
		switch {
		case errors.Is(err, ErrWrongPassword):
			os.Exit(ExitDecrypt)
		case errors.Is(err, ErrInvalidArgs):
			os.Exit(ExitUsage)
		default:
			os.Exit(ExitGeneral)
		}
	}
}
