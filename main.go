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
	fileExt       = ".fox"
	saltLen       = 16
	keyLen        = 32
	formatVer     = 0x01
	argonTime     = 3
	argonMemory   = 256 * 1024
	argonThreads  = 4
)

var (
	version = "1.0.0"

	ErrInvalidFile   = errors.New("invalid or corrupted file")
	ErrWrongPassword = errors.New("decryption failed: wrong password or corrupted file")
	ErrShortFile     = errors.New("file too short")
	ErrInvalidArgs   = errors.New("invalid arguments")
)

const (
	ExitSuccess = 0
	ExitGeneral = 1
	ExitUsage   = 2
	ExitDecrypt = 3
)

// shred overwrites the file with random data before deleting it
func shred(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	for pass := 0; pass < 3; pass++ {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}
		buff := make([]byte, 4096)
		for i := int64(0); i < info.Size(); i += 4096 {
			if _, err := io.ReadFull(rand.Reader, buff); err != nil {
				return err
			}
			if _, err := f.Write(buff); err != nil {
				return err
			}
		}
		f.Sync()
	}
	f.Close()
	return os.Remove(filename)
}

func deriveKey(password []byte, salt []byte) []byte {
	defer func() {
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
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if len(pass) == 0 {
		return nil, errors.New("password cannot be empty")
	}
	return pass, nil
}

func encrypt(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read source failed: %w", err)
	}

	outName := filename + fileExt
	if _, err := os.Stat(outName); err == nil {
		return fmt.Errorf("output file %s already exists", outName)
	}

	password, err := readPassword("Set password: ")
	if err != nil {
		return err
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}
	if string(password) != string(confirm) {
		return errors.New("passwords do not match")
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("salt generation failed: %w", err)
	}

	key := deriveKey(password, salt)
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return fmt.Errorf("cipher init failed: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("nonce generation failed: %w", err)
	}

	output := make([]byte, 1+len(salt)+len(nonce))
	output[0] = formatVer
	copy(output[1:], salt)
	copy(output[1+len(salt):], nonce)

	encrypted := aead.Seal(output, nonce, data, nil)

	if err := os.WriteFile(outName, encrypted, 0600); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✔ Encrypted: %s\n", outName)
	return nil
}

func decrypt(filename string, burn bool) error {
	if len(filename) < len(fileExt) || filename[len(filename)-len(fileExt):] != fileExt {
		return fmt.Errorf("%w: not a .fox file", ErrInvalidArgs)
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	if len(raw) < 1+saltLen+chacha20poly1305.NonceSizeX {
		return ErrShortFile
	}

	if raw[0] != formatVer {
		return fmt.Errorf("%w: unsupported version 0x%x", ErrInvalidFile, raw[0])
	}

	salt := raw[1 : 1+saltLen]
	nonce := raw[1+saltLen : 1+saltLen+chacha20poly1305.NonceSizeX]
	ciphertext := raw[1+saltLen+chacha20poly1305.NonceSizeX:]

	password, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	key := deriveKey(password, salt)
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return fmt.Errorf("cipher init failed: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ErrWrongPassword
	}
	defer func() {
		for i := range plaintext {
			plaintext[i] = 0
		}
	}()

	os.Stdout.Write(plaintext)

	if burn {
		fmt.Fprint(os.Stderr, "\n🔥 Delete encrypted file after decrypt? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			return errors.New("burn cancelled")
		}

		fmt.Fprintln(os.Stderr, "🔥 Burning file...")
		return shred(filename)
	}

	return nil
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("fox-vault %s\n", version)
		os.Exit(ExitSuccess)
	}

	mode := flag.String("m", "", "Mode: lock, unlock, or burn")
	file := flag.String("f", "", "Target file")
	flag.Parse()

	if *file == "" || *mode == "" {
		fmt.Fprintln(os.Stderr, "Usage: fox -m [lock|unlock|burn] -f <filename>")
		os.Exit(ExitUsage)
	}

	var err error
	switch *mode {
	case "lock":
		err = encrypt(*file)
	case "unlock":
		err = decrypt(*file, false)
	case "burn":
		err = decrypt(*file, true)
	default:
		fmt.Fprintln(os.Stderr, "Invalid mode. Use -m lock, unlock, or burn")
		os.Exit(ExitUsage)
	}

	if err != nil {
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