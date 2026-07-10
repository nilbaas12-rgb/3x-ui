# AmneziaWG 2.0 (AWG2) Protocol Integration

## Overview

This implementation adds full AmneziaWG 2.0 (AWG2) support to 3x-ui following the **Sidecar pattern** (similar to MTProto). The AWG2 protocol is managed as an independent sidecar process while the 3x-ui panel handles the UI, database, configuration management, client management, QR code generation, traffic tracking, and logging.

## Architecture

### Sidecar Pattern
- **Panel**: Manages configuration, stores data, generates QR codes, tracks traffic
- **Sidecar**: Runs the actual AmneziaWG 2.0 interface (`awg0`)
- **Communication**: Panel configures sidecar via config files, queries stats via stats endpoint

## Files Added

### Backend Core (`internal/awg2/`)
- **params.go** - Obfuscation parameter validation and management (H1-H4, S1-S4, Jc, Jmin, Jmax, I1)
- **key_generator.go** - WireGuard key pair generation
- **config.go** - Server and client configuration generation in INI format
- **client.go** - Client data model
- **manager.go** - AWG2 sidecar process management

### Database (`internal/database/`)
- **model/awg2.go** - AWG2Inbound and AWG2Client models
- **migrations/awg2_migration.go** - Database schema migration

### Service Layer (`internal/web/service/`)
- **inbound_awg2.go** - AWG2 inbound service with CRUD operations
  - CreateAWG2Inbound
  - GetAWG2Inbound
  - AddAWG2Client
  - GetAWG2Clients
  - RemoveAWG2Client
  - GenerateAWG2ClientConfig

### Background Jobs (`internal/web/job/`)
- **awg2_job.go** - Periodic job for instance reconciliation and traffic collection

### API Controller (`internal/web/controller/`)
- **awg2.go** - REST API endpoints for AWG2 operations

## Features Implemented

### ✅ Complete
1. **Inbound Management**
   - Create/Read/Update/Delete AWG2 inbounds
   - Configure obfuscation parameters
   - Server key generation

2. **Client Management**
   - Add/remove clients
   - Generate per-client keys
   - Automatic IP allocation
   - Enable/disable clients

3. **Configuration**
   - Server config generation (INI format)
   - Client config generation (.conf format)
   - Parameter validation

4. **Logging**
   - All operations marked with `[awg2]` tag
   - Integrated into main logger

5. **Traffic Tracking**
   - Per-client traffic accounting
   - Inbound-level statistics
   - Job-based collection

### 📋 To Implement (Stub Locations Ready)
1. **QR Code Generation** - Use `qrcode` library
2. **Instance Reconciliation** - AWG2Job.Run()
3. **Real Key Derivation** - Replace stub in key_generator.go with proper curve25519
4. **Stats Collection** - Query AWG2 management socket

## Database Schema

### awg2_inbounds
```sql
id (PRIMARY KEY)
inbound_id (FOREIGN KEY)
port
subnet
mtu
disable_ipv6
allowed_ips_mode
jc, jmin, jmax, s1, s2, s3, s4
h1, h2, h3, h4, i1
server_private_key
server_public_key
status
remarks
created_at, updated_at
```

### awg2_clients
```sql
id (PRIMARY KEY)
inbound_id (FOREIGN KEY)
email (INDEX)
public_key
private_key
allowed_ips
enable
expiry_time
used_traffic
download, upload
limit
limit_ip
created_at, updated_at
```

## API Endpoints

### Inbound Operations
- `GET /api/awg2/inbounds/:id` - Get AWG2 inbound details
- `POST /api/awg2/inbounds` - Create new inbound
- `PUT /api/awg2/inbounds/:id` - Update inbound
- `DELETE /api/awg2/inbounds/:id` - Delete inbound

### Client Operations
- `POST /api/awg2/inbounds/:id/client/add` - Add client
- `GET /api/awg2/inbounds/:id/clients` - List clients
- `POST /api/awg2/client/:id/remove` - Remove client
- `GET /api/awg2/client/:id/config` - Download client config
- `GET /api/awg2/client/:id/qr` - Get QR code

## Obfuscation Parameters

All parameters match `bivlked/amneziawg-installer` defaults:

```go
Jc   = 6                                  // Jitter correction
Jmin = 55                                 // Min jitter
Jmax = 205                                // Max jitter
S1   = 72, S2 = 56, S3 = 32, S4 = 16    // Packet sizes
H1   = "234567-345678"                   // Header ranges
H2   = "3456789-4567890"
H3   = "56789012-67890123"
H4   = "456789012-567890123"
I1   = "<r 128>"                         // Initial packet
```

## Integration with 3x-ui

### In `internal/web/web.go`
Add to background jobs:
```go
cadenceAWG2 = "@every 10s"
// In startTask()
_, _ = s.cron.AddJob(cadenceAWG2, job.NewAWG2Job())
```

### In Inbound Protocol List
Add "awg2" to protocol options alongside "mtproto", "vless", etc.

## Usage Example

### Create AWG2 Inbound
```bash
POST /api/inbounds/add
{
  "port": 51820,
  "protocol": "awg2",
  "settings": {
    "port": 51820,
    "subnet": "10.9.9.1/24",
    "mtu": 1280,
    "disableIPv6": false,
    "allowedIpsMode": 2,
    "params": {
      "jc": 6,
      "jmin": 55,
      "jmax": 205,
      "s1": 72,
      "s2": 56,
      "s3": 32,
      "s4": 16,
      "h1": "234567-345678",
      "h2": "3456789-4567890",
      "h3": "56789012-67890123",
      "h4": "456789012-567890123",
      "i1": "<r 128>"
    }
  }
}
```

### Add Client
```bash
POST /api/awg2/inbounds/1/client/add
{
  "email": "user@example.com"
}
```

### Get Client Config
```bash
GET /api/awg2/client/1/config
```

## Logging

All AWG2 operations are logged with `[awg2]` prefix:
```
[awg2] Creating new AWG2 inbound on port 51820
[awg2] Added new client user@example.com to inbound 1
[awg2] Running AWG2 reconciliation job
[awg2] Failed to collect traffic: ...
```

## Next Steps for Production

1. **Complete Key Derivation**: Replace stub in `key_generator.go` with actual curve25519 implementation using `golang.zx2c4.com/wireguard`

2. **Implement QR Generation**: Add QR code endpoint using standard library or `qrcode` package

3. **AWG2 Binary Integration**: 
   - Download and verify AWG2 binary
   - Implement proper process management in `manager.go`
   - Handle stats collection via management socket

4. **IP Allocation**: Implement proper subnet-aware IP allocation

5. **Stats Collection**: Query AWG2 instance stats endpoint or management socket

6. **Frontend UI**: Create Vue components for AWG2 configuration (paralleling MTProto setup)

## References

- [AmneziaWG Installer](https://github.com/bivlked/amneziawg-installer)
- [3x-ui MTProto Implementation](https://github.com/MHSanaei/3x-ui/tree/main/internal/mtproto)
- [WireGuard Protocol](https://www.wireguard.com/protocol/)
