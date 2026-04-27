# 🦊 fox-vault

**Simple. Practical. Reliable.**

[![Release](https://img.shields.io/github/v/release/foxhackerzdevs/fox-vault)](https://github.com/foxhackerzdevs/fox-vault/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/foxhackerzdevs/fox-vault)](https://goreportcard.com/report/github.com/foxhackerzdevs/fox-vault)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 📌 Introduction

`fox-vault` is a minimalist CLI tool for file encryption using modern cryptography.

No complex configurations. No cloud sync. Just secure, local encryption.

---

## ❓ Why fox-vault?

* No insecure modes or configuration footguns
* No external dependencies or services
* Uses modern, audited cryptographic primitives
* Designed for local, deliberate workflows

---

## ✨ Features

* **Hard to Misuse:** XChaCha20-Poly1305 + Argon2id
* **Tamper Evident:** AEAD prevents silent corruption
* **Memory Safe:** Best-effort zeroing of sensitive data
* **Zero Config:** Single static binary
* **Safe Defaults:** Refuses to overwrite files
* **Flexible Decryption:** Safe (`unlock`) and destructive (`burn`) modes

---

## 📦 Installation

### Pre-built Binaries

👉 [https://github.com/foxhackerzdevs/fox-vault/releases](https://github.com/foxhackerzdevs/fox-vault/releases)

---

### Verify Checksums

```bash
sha256sum -c fox-v1.0.1-linux-amd64.sha256
```

---

### Install via Go

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

### 🔐 Encrypt

```bash
./fox -m lock -f secret.txt
```

Creates:

```
secret.txt.fox
```

---

### 🔓 Decrypt (safe)

```bash
./fox -m unlock -f secret.txt.fox
```

* Prints decrypted content to stdout
* Does **NOT** delete the file

Recommended:

```bash
./fox -m unlock -f file.fox | less
./fox -m unlock -f file.fox > output.txt
```

---

### 🔥 Decrypt + Delete (burn)

```bash
./fox -m burn -f secret.txt.fox
```

* Decrypts file
* Prints content
* Prompts before deletion
* Deletes encrypted file after confirmation

⚠️ **This operation is destructive**

---

### 📦 Version

```bash
./fox -v
```

---

## 📦 Exit Codes

* `0` Success
* `1` General error
* `2` Invalid arguments
* `3` Decryption failed

---

## 🔒 Security Specs

* **KDF:** Argon2id (`Time=3`, `Memory=256MB`, `Threads=4`)
* **Encryption:** XChaCha20-Poly1305
* **Nonce:** 24 bytes

**File Format:**

```
[1-byte Version][16-byte Salt][24-byte Nonce][Ciphertext]
```

* **Password Input:** Secure terminal input
* **File Permissions:** `0600`

---

## ⚠️ What fox-vault is NOT

* Not for plausible deniability
* Not for hiding filenames
* Not for very large files (RAM-bound, recommended <1GB)
* Not a password manager
* No password recovery — lost password = lost data

---

## ⚠️ Secure Deletion Warning

The `burn` mode uses overwrite + delete.

❗ On SSDs and modern filesystems, secure deletion is **not guaranteed**.

---

## 🏗 Building from Source

```bash
go mod tidy
go build -trimpath -ldflags="-s -w -X main.version=dev" -o fox main.go
```

---

## 🔁 Reproducible Builds

Uses:

* `-trimpath`
* stripped binaries (`-s -w`)

---

## 🗺 Roadmap

* [ ] Replace `burn` with `--delete` flag
* [ ] Streaming support (>1GB files)
* [ ] Output file support (`-o output.txt`)

---

## ⚖️ License

MIT License — see [LICENSE](LICENSE)

---

## ⚠️ Disclaimer

This software is provided **"as is"**.

* No warranty
* No recovery
* No guarantees

You are responsible for your data and passwords.

---

## 👥 Contributors

Built by [https://github.com/foxhackerzdevs](https://github.com/foxhackerzdevs)
Pull requests welcome.
