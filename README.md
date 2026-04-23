# 🦊 fox-vault

**Simple. Practical. Reliable.**

[![Release](https://img.shields.io/github/v/release/foxhackerzdevs/fox-vault)](https://github.com/foxhackerzdevs/fox-vault/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/foxhackerzdevs/fox-vault)](https://goreportcard.com/report/github.com/foxhackerzdevs/fox-vault)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 📌 Introduction

`fox-vault` is a minimalist CLI tool for file encryption. It follows the philosophy of building tools that are hard to misuse and easy to trust.

No complex configurations. No cloud sync. Just industry-standard cryptography.

---

## ❓ Why fox-vault?

* No insecure modes or configuration footguns
* No external dependencies or services
* Uses modern, audited cryptographic primitives by default
* Designed for local, deliberate encryption workflows

---

## ✨ Features

* **Hard to Misuse:** Uses **XChaCha20-Poly1305** and **Argon2id** by default—no algorithm selection needed
* **Tamper Evident:** AEAD ensures decryption fails if data is modified or password is incorrect
* **Memory Safe:** Best-effort zeroing of sensitive data after use
* **Zero Config:** Single static binary
* **Practical UX:** Password masking + confirmation
* **Safe by Default:** Refuses to overwrite existing files

---

## 📦 Installation

### Pre-built Binaries

Download the latest release from:
👉 [https://github.com/foxhackerzdevs/fox-vault/releases](https://github.com/foxhackerzdevs/fox-vault/releases)

### Verify Checksums

Always verify downloads before running:

```bash
sha256sum -c fox-v1.0.0-linux-amd64.sha256
```

Checksums are published alongside releases.

---

### Install via Go

Requires Go 1.21+

```bash
go install github.com/foxhackerzdevs/fox-vault@latest
```

---

### Build Locally

```bash
git clone https://github.com/foxhackerzdevs/fox-vault
cd fox-vault
go build -trimpath -ldflags="-s -w" -o fox main.go
```

---

## 🛠 Usage

### Encrypt a file

Creates `secret.txt.fox`. Refuses to overwrite existing files.

```bash
./fox -m lock -f secret.txt
```

---

### View (decrypt) a file

Prints decrypted content to stdout:

```bash
./fox -m view -f secret.txt.fox
```

Recommended usage:

```bash
./fox -m view -f secret.txt.fox | less
./fox -m view -f secret.txt.fox > decrypted.txt
```

---

### Check version

```bash
./fox -v
```

---

## 📦 Exit Codes

* `0` Success
* `1` General error
* `2` Invalid arguments
* `3` Decryption failed (wrong password or corrupted file)

---

## 🔒 Security Specs

* **KDF:** Argon2id (`Time=3`, `Memory=256MB`, `Threads=4`)
* **Encryption:** XChaCha20-Poly1305 (24-byte nonce)
* **File Format:**
  `[1-byte Version] [16-byte Salt] [24-byte Nonce] [Ciphertext]`
* **Password Input:** Secure input via terminal (never CLI args)
* **File Permissions:** `0600`

---

## ⚠️ What fox-vault is NOT

* Not for plausible deniability (files are obviously encrypted)
* Not for hiding filenames (`secret.txt.fox` still reveals name)
* Not for very large files (entire file is loaded into memory; practical limit depends on available RAM, recommended <1GB)
* Not a replacement for full-disk encryption
* Not a password manager
* No password recovery — lost password = lost data

---

## 🏗 Building from Source

```bash
go get golang.org/x/crypto@latest
go get golang.org/x/term@latest
go build -trimpath -ldflags="-s -w -X main.version=dev" -o fox main.go
```

---

## 🔁 Reproducible Builds

Builds use `-trimpath` and stripped binaries to improve reproducibility.

---

## 🗺 Roadmap

* [ ] `-delete` flag to securely wipe original file
* [ ] Streaming support for large files (>1GB)
* [ ] Output file support:
  `fox -m unlock -f file.fox -o output`

---

## ⚖️ License

MIT License — see [LICENSE](LICENSE)

---

## ⚠️ Disclaimer

This software is provided **"as is"**, without warranty of any kind.

You are responsible for managing your passwords securely.
There are no backdoors—if you lose your password, your data is permanently unrecoverable.

---

## 👥 Contributors

Built by [https://github.com/foxhackerzdevs](https://github.com/foxhackerzdevs)
Pull requests are welcome.

