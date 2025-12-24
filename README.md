# SmartCalc

A powerful, multi-purpose calculator application with support for mathematical expressions, date/time calculations, network/IP operations, and more.
<img width="945" height="1027" alt="Screenshot 2025-12-18 at 15 26 31" src="https://github.com/user-attachments/assets/3ab30b3c-336d-4413-917c-f0e0eaa6ade4" />

## Features

### Basic Math & Currency
- Standard arithmetic operations with proper operator precedence
- Percentage calculations with smart context (e.g., `$100 - 20%`)
- Currency formatting with thousands separators
- Scientific functions (sin, cos, tan, sqrt, log, etc.)
- Line references to use previous results (`\1`, `\2`, etc.)

### Comparison Expressions
- Compare values with `>`, `<`, `>=`, `<=`, `==`, `!=`
- Results displayed as `true` or `false`

### Number Base Conversions
- Convert between decimal, hexadecimal, octal, and binary
- Supports input in any base format

### Date & Time Calculations
- Current time: `now`, `today()`
- Time in different cities: `now in Seattle`, `now in Kiev`
- Date arithmetic: `today() + 30 days`, `now - 1 week`
- Date difference: `19/01/22 - now` (shows years, months, weeks, days, hours, minutes)
- Duration conversion: `861.5 hours in days`
- Time zone conversion: `6:00 am Seattle in Kiev`
- Date ranges: `Dec 6 till March 11`
- Time arithmetic with timezone: `12 am PST - 3 hours`

### Network/IP Calculations
- Subnet information: `10.100.0.0/24`
- Split network by count: `10.100.0.0/16 / 4 subnets` or `10.100.0.0/16 / 4 networks`
- Split by host count: `10.100.0.0/28 / 16 hosts`
- Subnet mask: `mask for /24`, `wildcard for /24`
- IP range check: `is 10.100.0.50 in 10.100.0.0/24`
- DNS lookup: `dig google.com`, `nslookup github.com` (shows CNAME chain, A/AAAA, MX, NS, TXT records)
- WHOIS lookup: `whois google.com` (shows registrar, dates, name servers)
- IP geolocation: `geoip 8.8.8.8`, `ip lookup 8.8.8.8` (shows location, ISP, coordinates, timezone)

### SSL Certificate Decoder
- Decode certificates: `cert decode https://google.com` or `ssl decode example.com`
- Test certificates: `cert test https://expired.badssl.com`
- Shows subject, issuer, validity, SANs, key usage
- Displays certificate chain as tree

### Unit Conversions
- Length: `5 miles in km`, `100 cm to inches`
- Weight: `10 kg in lbs`, `5 oz to grams`
- Temperature: `100 f to c`, `25 celsius to fahrenheit`
- Volume: `5 gallons in liters`, `2 cups to ml`
- Data (SI, base 1000): `1234567 bytes to mb`, `500 mb in gb`, `1 tb to gb`
- Data (IEC, base 1024): `1234567 bytes to mib`, `1024 mib to gib`, `1 tib to gib`
- Speed: `60 mph to kph`
- Area: `1 acre to sqft`, `100 sqm to sqft`

### Color Conversions
- Hex to RGB: `#FF5733 to rgb`, `#FFF to rgb`
- Hex to HSL: `#FF5733 to hsl`
- RGB to Hex: `rgb(255, 87, 51) to hex`
- RGB to HSL: `rgb(255, 0, 0) to hsl`
- HSL to RGB: `hsl(0, 100%, 50%) to rgb`
- HSL to Hex: `hsl(240, 100%, 50%) to hex`

### Percentage Calculations
- What is X% of Y: `what is 15% of 200`
- What percent is X of Y: `50 is what % of 200`
- Increase/decrease: `increase 100 by 20%`, `decrease 500 by 15%`
- Percent change: `percent change from 50 to 75`
- Tip calculator: `tip 20% on $85.50`
- Bill splitting: `$150 split 4 ways with 18% tip`

### Financial Calculations
- Loan payments: `loan $250000 at 6.5% for 30 years`
- Mortgage: `mortgage $350000 at 7% for 30 years`
- Compound interest: `$10000 at 5% for 10 years compounded monthly`
- Simple interest: `simple interest $5000 at 3% for 2 years`
- Investment growth: `invest $1000 at 7% for 20 years`

### Statistics
- Average: `avg(10, 20, 30, 40)` or `mean(1, 2, 3, 4, 5)`
- Median: `median(1, 2, 3, 4, 100)`
- Sum: `sum(10, 20, 30)`
- Min/Max: `min(5, 10, 3)`, `max(5, 10, 3)`
- Standard deviation: `stddev(2, 4, 4, 4, 5, 5, 7, 9)`
- Variance: `variance(1, 2, 3, 4, 5)`
- Count: `count(1, 2, 3, 4, 5)`

### Programmer Utilities
- Bitwise operations: `0xFF AND 0x0F`, `0xF0 OR 0x0F`, `0xFF XOR 0x0F`
- Bit shifts: `1 << 8`, `256 >> 4`
- ASCII/Char: `ascii A`, `char 65`
- ASCII Table: `ascii table` (displays full ASCII table)
- UUID generation: `uuid`
- Hash functions: `md5 hello`, `sha256 hello`
- Base64 encoding: `base64 encode hello world`, `base64 decode SGVsbG8gd29ybGQ=`
- Password generator: `pwgen`, `pwgen -c 20` (custom length), `pwgen -h` (hyphenated)

### Regex Tester
- Basic match: `regex /hello/ test "hello world"`
- Multiple matches: `regex /\d+/ test "a1b2c3"`
- Capture groups: `regex /(\w+)@(\w+)\.(\w+)/ test "user@example.com"`
- Case insensitive: `regex /(?i)hello/ test "HELLO World"`
- Word boundary: `regex /\bword\b/ test "a word here"`
- Matching parts are highlighted with `«»` markers
- Captured groups are displayed in multi-line output

### Unix Permissions
- Chmod octal to symbolic: `chmod 755`, `chmod 644`, `chmod 4755`
- Chmod symbolic to octal: `chmod rwxr-xr-x`, `chmod rwx r-x r-x`
- Umask calculator: `umask 022`, `umask 077`
- Special bits: setuid (`4xxx`), setgid (`2xxx`), sticky (`1xxx`)
- Permission conversions: `755 to symbolic`, `rwxr-xr-x to octal`

### JWT Decoder
- Decode JWT tokens: `jwt decode <token>` or `jwt <token>`
- Shows header (algorithm, type)
- Shows payload with all claims
- Converts timestamps (exp, iat, nbf) to human-readable format
- Shows token expiration status (valid/expired)

### Physical & Mathematical Constants
- Mathematical: `pi`, `e`, `phi`, `golden ratio`
- Physical: `speed of light`, `gravity`, `avogadro`, `planck`
- Lookup: `value of pi`, `value of speed of light`

## Examples

```
# Basic Math
10 + 20 * 3 = 70
$100 - 20% = $80.00
sin(45) + cos(30) = 1.57

# Line References
100 = 100
\1 * 2 = 200

# Comparisons
25 > 2.5 = true
100 >= 100 = true
5 != 3 = true

# Base Conversions
255 in hex = 0xFF
0xFF in dec = 255
25 in bin = 0b11001
0b11001 in oct = 0o31

# Date & Time
now = 2025-12-18 15:04:32 PST
now in Kiev = 2025-12-19 01:04:32 EET
today() + 30 days = 2026-01-17
19/01/22 - now = 3 years 10 months 4 weeks 1 day 14 hours 13 min
12 am PST - 3 hours = 2025-12-17 21:00 PST

# Network/IP
10.100.0.0/24 = 
> Network: 10.100.0.0/24
> Hosts: 254
> Range: 10.100.0.1 - 10.100.0.254
> Mask: 255.255.255.0

10.100.0.0/16 / 4 networks = 
> 1: 10.100.0.0/18 (16382 hosts)
> 2: 10.100.64.0/18 (16382 hosts)
> 3: 10.100.128.0/18 (16382 hosts)
> 4: 10.100.192.0/18 (16382 hosts)

mask for /24 = 255.255.255.0
wildcard for /24 = 0.0.0.255
is 10.100.0.50 in 10.100.0.0/24 = yes

# DNS Lookup
dig google.com =
> DNS Lookup: google.com
> A Records:
>   142.250.80.46
> MX Records:
>   smtp.google.com (priority: 10)

# WHOIS Lookup
whois google.com =
> WHOIS: google.com
> Registrar: MarkMonitor Inc.
> Created: 1997-09-15T04:00:00Z
> Expires: 2028-09-14T04:00:00Z

# IP Geolocation
geoip 8.8.8.8 =
> Location: Mountain View, California, United States
> ISP: Google LLC
> Coords: 37.4056, -122.0775
> Timezone: America/Los_Angeles

# SSL Certificate
cert decode https://google.com =
> Subject: *.google.com
> Issuer: GTS CA 1C3
> Status: ✓ Valid (expires in 60 days)
> Certificate Chain: 3 certificates
> 🔐 GTS Root R1 (root)
>    └── GTS CA 1C3 (intermediate)
>        └── *.google.com (leaf)

# Unit Conversions
5 miles in km = 8.05 km
100 f to c = 37.78°C
10 kg in lbs = 22.05 lbs
500 mb in gb = 0.49 GB

# Percentage Calculations
what is 15% of 200 = 30
50 is what % of 200 = 25.00%
increase 100 by 20% = 120
tip 20% on $85.50 = Tip: $17.10, Total: $102.60

# Financial Calculations
loan $250000 at 6.5% for 30 years = Monthly: $1580.17
> Total: $568,861.22
> Interest: $318,861.22

# Statistics
avg(10, 20, 30, 40) = 25
median(1, 2, 3, 4, 100) = 3
stddev(2, 4, 4, 4, 5, 5, 7, 9) = 2

# Programmer Utilities
0xFF AND 0x0F = 15 (0xF)
1 << 8 = 256 (0x100)
ascii A = 65 (0x41)
uuid = a1b2c3d4-e5f6-7890-abcd-ef1234567890
base64 encode hello world = aGVsbG8gd29ybGQ=
base64 decode SGVsbG8gd29ybGQ= = hello world

# Physical Constants
pi = 3.141592654
speed of light = 2.99792458e+08 m/s
gravity = 9.80665 m/s²

# Regex Tester
regex /hello/ test "hello world" = match: «hello» world
regex /\d+/ test "a1b2c3" = 3 matches: a«1»b«2»c«3»
regex /(\w+)@(\w+)/ test "user@domain" = match: «user@domain»
> Groups:
>   [1]: "user"
>   [2]: "domain"

# Unix Permissions
chmod 755 = rwxr-xr-x
chmod 644 = rw-r--r--
chmod rwxr-xr-x = 755
chmod 4755 = rwsr-xr-x (setuid)
chmod 1777 = rwxrwxrwt (sticky)
umask 022 = files: 644 (rw-r--r--), directories: 755 (rwxr-xr-x)

# JWT Decoder
jwt decode eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig =
> Header:
>   {"alg": "HS256"}
> Payload:
>   {"sub": "1234567890"}
> Signature: sig
```

## Installation

### Download Pre-built Packages

Ready-to-use packages for all major platforms are available on the [GitHub Releases](https://github.com/vpoluyaktov/smartcalc/releases) page:

- **Windows**: `SmartCalc-windows-amd64.exe`
- **macOS (Intel)**: `SmartCalc-darwin-amd64.app.zip`
- **macOS (Apple Silicon)**: `SmartCalc-darwin-arm64.app.zip`
- **Linux**: `SmartCalc-linux-amd64`

### Build from Source

#### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or later
- [Node.js](https://nodejs.org/) 18 or later
- [Wails](https://wails.io/) v2

#### Install Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### Platform-specific Dependencies

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
- Install [MSYS2](https://www.msys2.org/) or use WSL

#### Build Steps

1. Clone the repository:
```bash
git clone https://github.com/vpoluyaktov/smartcalc.git
cd smartcalc
```

2. Install frontend dependencies:
```bash
cd frontend
npm install
cd ..
```

3. Build the application:
```bash
wails build
```

The built application will be in the `build/bin` directory.

#### Development Mode

To run in development mode with hot reload:
```bash
wails dev
```

## Usage Tips

- Press **Enter** at the end of a line to auto-append `=` and evaluate
- Use **Ctrl+C** to copy with line references resolved to actual values
- Use **Ctrl+V** to paste directly
- Check the **Snippets** menu for example expressions
- Lines starting with `#` are treated as comments
- Use `\1`, `\2`, etc. to reference results from previous lines

## License

MIT License - see [LICENSE](LICENSE) for details.
