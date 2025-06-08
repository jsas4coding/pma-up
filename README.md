# phpMyAdmin Updater (pma-up)

![CI](https://github.com/jsas4coding/pma-up/actions/workflows/release.yml/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.24.4-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
![GitHub release (latest by tag)](https://img.shields.io/github/v/release/jsas4coding/pma-up)
[![codecov](https://codecov.io/gh/jsas4coding/pma-up/branch/main/graph/badge.svg?token=36JSSXXHB3)](https://codecov.io/gh/jsas4coding/pma-up)

**phpMyAdmin Updater — CLI tool for fully automated phpMyAdmin updates**

This project automates the full update process of a phpMyAdmin installation, downloading the latest release, safely replacing the current installation, preserving configuration files, and creating backups.

> ⚠ **Project Notice**
> This repository is primarily for personal/internal use.
> Pull Requests and Issues from external contributors are not being accepted.
> Feel free to use it as-is or fork for your own usage.

---

## Features

- Automatically fetches the latest phpMyAdmin release.
- Verifies version file directly from phpMyAdmin servers.
- Downloads and extracts the latest zip archive.
- Backs up existing installation before upgrade.
- Preserves your existing `config.inc.php` file.
- Fully automated with detailed logging.
- Built with paranoid error checking.
- Designed for cron-based unattended updates.

---

## Installation

### 1️⃣ Download prebuilt binary

Prebuilt releases are available at:
[GitHub Releases](https://github.com/jsas4coding/pma-up/releases)

Choose the binary appropriate for your platform, and place it somewhere in your `$PATH`.

Example (Linux amd64):

```bash
wget https://github.com/jsas4coding/pma-up/releases/download/vX.Y.Z/pma-up_X.Y.Z_linux_amd64.tar.gz
tar -xzf pma-up_X.Y.Z_linux_amd64.tar.gz
sudo mv pma-up /usr/local/bin/
```

### 2️⃣ Build locally (optional)

If you prefer to build from source:

```bash
git clone https://github.com/jsas4coding/pma-up.git
cd pma-up
make build
```

This will generate a local `pma-up` binary.

---

## Usage

```bash
pma-up <phpmyadmin_path> <config_file_path>
```

Example:

```bash
pma-up /var/www/html/phpmyadmin /path/to/config.inc.php
```

The tool will:

- Create a backup directory with timestamp:
  `/var/www/html/phpmyadmin_backup_YYYYMMDDHHMMSS`
- Download the latest release.
- Extract and replace safely.
- Restore your existing `config.inc.php`.

---

## Automating with crontab

To schedule periodic updates automatically:

```bash
crontab -e
```

Add a line similar to:

```bash
0 3 * * 0 /usr/local/bin/pma-up /var/www/html/phpmyadmin /path/to/config.inc.php >> /var/log/pma-up.log 2>&1
```

- Runs every Sunday at 3AM.
- Logs output to `/var/log/pma-up.log`.

✅ Always verify functionality manually before automating.

---

## Testing

You can run full tests locally:

```bash
make test     # Unit tests
make e2e      # End-to-end tests
make lint     # Linter check
```

---

## Test Coverage Philosophy

The **phpMyAdmin Updater** project applies paranoid-grade testing strategy:

- ✅ Functional flows fully tested.
- ✅ Error handling fully validated.
- ✅ Full linter compliance (`golangci-lint clean`).
- ✅ Coverage reports currently show ~65%-70% due to inherent Go tooling limitations.

### Why is coverage not 100%?

The following conditions in Go are technically difficult to cover via tests:

- `defer Close()` errors are nearly impossible to force under normal conditions.
- `file.Open()` failures cannot be simulated easily without dependency injection.
- `scanner.Err()` in `bufio.Scanner` requires custom Reader injection not feasible in pure unit tests.
- These code paths remain safe but untriggered.

We prefer honest functional coverage rather than artificially inflating coverage via aggressive mocking.

### Coverage Visualization

![Coverage Sunburst](https://codecov.io/gh/jsas4coding/pma-up/graphs/sunburst.svg?token=36JSSXXHB3)

You can explore the current detailed coverage report [here](https://codecov.io/gh/jsas4coding/pma-up).

### Reference

- [Go coverage: known uncovered patterns](https://github.com/golang/go/issues/18835)
- [The cost of 100% test coverage in Go](https://peter.bourgon.org/blog/2017/06/09/test-coverage.html)

---

## License

This project is licensed under the MIT License.

It is provided "as is", without warranty of any kind.

The repository is public, but not actively maintained as a community-driven project.

Use at your own risk and discretion.
