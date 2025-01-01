# Lisk Automation Project

## Overview

**Lisk Automation Project** is a Go-based project designed to automate interactions with blockchain applications, including decentralised exchanges (DEX) and lending protocols. It has a modular architecture to support multiple accounts, transaction management and interaction with smart contracts.

---

## Features

- **Multi-Account Automation**: Concurrently processes multiple blockchain accounts.
- **DEX Integration**: Handles swaps, liquidity management, and pool interactions.
- **Lending Protocols**: Supports borrowing, lending, and collateral management.
- **Modular Design**: Easily extendable with new modules.
- **Comprehensive Logging**: Tracks transactions and provides debugging information.
- **Top Cheker**: Checking accounts in the leaderboard. Logging the number of points, place in the table and time of the last update
- **Task performer**: Marks all available assignments complete, provided they are completed in advance. Also - daily task
- **Version control system**: Automatic checking of the software version relevance and registration of warnings in case of an update, as well as a link to the new release

---
## Installation

### Requirements

- **Go** (Version 1.22 or newer)
- Git (for cloning the repository)
- Optional: `make` for simplified build and run commands.

### Steps

1. Clone the repository:
```bash
git clone https://github.com/ssq0-0/Lisk.git
cd lisk
go mod download
go build -o Lisk ./core/main.go   
```
2. Run the application:

```bash
./Lisk
```

3. Or run main.go:
```bash
cd lisk
go run ./core/main.go
```
---

### Wallets (`wallets.csv`)

This section defines the wallets used by the software. Each wallet is described by the following fields:

- **`privatekey`**: The p-k of your wallet.
---
### Proxy (`proxy.txt`)

This section defines the proxy used by the program. Each proxy is described by the following fields:

- **`http://user:pass@ip:port`**:
---

### Modules (`modules`)

- Oku. Dex swaps.
- Ionic. Luquidity proocol.
- Relay. Bridge from other L2 chains to LISK.
- Top Checker.
- Task Performer.
- Daily checker.
- Version control.

### For additional assistance or troubleshooting, refer to the official documentation or reach out via [support channel](https://t.me/cheifssq).
