# Lisk Automation Project

## Overview

**Lisk Automation Project** is a Go-based project designed to automate interactions with blockchain applications, including decentralised exchanges (DEX) and lending protocols. It has a modular architecture to support multiple accounts, transaction management and interaction with smart contracts.

---

## Features

- **Multi-Account Automation**: Concurrently processes multiple blockchain accounts.
- **DEX Integration**: Handles swaps, liquidity management, and pool interactions. Dynamic swap to ETH if you suddenly run out of native coin to pay for gas
- **Ionic Lending Protocols**: Supports borrowing, lending, and collateral management.
- **Balance Sheet Verification**. The balance check module goes through all tokens in the network and writes all balances to a csv file
- **WRAP_UNWRAP**. The module is required to accumulate the number of transactions due to high gas. The cheapest option to get the number of transactions.  
- **Modular Design**: Easily extendable with new modules.
- **Comprehensive Logging**: Tracks transactions and provides debugging information.
- **Top Cheker**: Checking accounts in the leaderboard. Logging the number of points, place in the table and time of the last update
- **Task performer**: Marks all available assignments complete, provided they are completed in advance. Also - daily task
- **Version control system**: Automatic checking of the software version relevance and registration of warnings in case of an update, as well as a link to the new release
- **Debugging erroneous accounts**: If a critical error (no balance on the wallet) occurs during the execution of the programme, this functionality will move the wallet from the general list to a separate file for erroneous wallets, so that you can configure it afterwards and run it without the others. 
- **Statistics**:  Account statistics functionality: general statistics file with successful actions/portal points/rank/last update date

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
- **Setup config.json(time and count actions)**

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
### Config (`config.json`)

1. Customise your configuration. There are two fields you can operate on: ‘actions_count’ and ‘max_actions_time’. The first one is responsible for the number of actions in case of Oku & WRAP_UNWRAP, and the second one for the programme execution time. 

Full instructions for setting can be found in the same file, with detailed comments on each parameter. Initially, the basic configuration with average values is set up as follows

2. Balance
It is important to note that the minimum balances to work with the software: 
- +- 0.35$ in ETH
- 0.1 USDT/USDC

**If the balance is insufficient for the commission, any token (usdt/usdc) will be automatically exchanged to ETH. Works only in case of exchanges on oku**
---

### Modules (`modules`)

- Wraper. Module for WRAP/UNWRAP operations. 
- Oku. Random dex swaps.
- Ionic. Supply, repay + withdraw, borrow.
- Relay. Bridge ETH from other L2 chains to LISK.
- Top Checker. Makes a request to the platform and checks your rank+place+date of last updated information.
- Task Performer. Collects points for completed tasks on the platform.
- Daily checker. Makes a daily check on the platform.
- balance checker. Check balance in all token in LISK chain.
- Version control.

### For additional assistance or troubleshooting, refer to the official documentation or reach out via [support channel](https://t.me/cheifssq).
