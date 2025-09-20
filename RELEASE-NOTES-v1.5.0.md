# MarkGo Engine v1.5.0 - Content Creator Experience & Code Quality

**Release Date:** September 20, 2025
**Tag:** `v1.5.0`
**Status:** ✅ Production Ready

---

## 🚀 Major New Feature: Live Preview Service

This release introduces a **real-time preview service** for draft articles, significantly enhancing the developer experience with instant feedback during content creation.

## ✨ Key Features

### 🔄 Real-Time Preview System
- **NEW**: Live preview service for draft articles with WebSocket-based updates
- **NEW**: Real-time file watching with automatic browser reload
- **NEW**: Session-based preview management with configurable timeouts
- **NEW**: Multi-client preview support for collaborative editing
- **NEW**: Secure WebSocket origin validation for production safety

### 🛡️ Security & Reliability Improvements
- **FIXED**: Critical concurrency race condition in session management
- **ENHANCED**: WebSocket origin checking with proper host validation
- **IMPROVED**: Interface-based dependency injection for better testability
- **SECURED**: Preview sessions with auto-generated authentication tokens

### 🔧 Developer Experience Enhancements
- **ADDED**: Comprehensive Preview API documentation
- **UPDATED**: Configuration guide with new preview service options
- **IMPROVED**: Error handling and logging throughout preview system
- **ADDED**: Real-time session statistics and monitoring

### 🏆 Code Quality Excellence
- **ACHIEVED**: Zero lint warnings across entire codebase (300+ issues resolved)
- **OPTIMIZED**: Slice preallocation for improved memory performance
- **STANDARDIZED**: Function comments and documentation formatting
- **ENHANCED**: Security patterns with proper exemptions and error handling
- **IMPROVED**: Build and test reliability with comprehensive checks

## 📊 Technical Improvements

### New Preview Service Architecture
- **WebSocket Integration**: Real-time communication for instant updates
- **File System Watching**: Automatic detection of article changes
- **Session Management**: Configurable timeouts and session limits
- **Origin Security**: Comprehensive CORS and origin validation
- **Performance Optimized**: Minimal overhead with efficient goroutine usage

### API Additions
- `POST /api/preview/sessions` - Create preview sessions
- `GET /api/preview/sessions` - List active sessions
- `DELETE /api/preview/sessions/{id}` - Terminate sessions
- `GET /api/preview/ws/{id}` - WebSocket connection endpoint
- `GET /preview/{id}` - Serve preview pages

### Configuration Options
- `PREVIEW_ENABLED` - Enable/disable preview service
- `PREVIEW_PORT` - WebSocket port configuration
- `PREVIEW_BASE_URL` - Custom preview base URL
- `PREVIEW_MAX_SESSIONS` - Maximum concurrent sessions
- `PREVIEW_SESSION_TIMEOUT` - Session timeout duration

## 🔄 Breaking Changes

**None** - This is a backward-compatible release. All existing functionality remains unchanged.

## 🛠️ Installation & Upgrade

### New Installation
```bash
# Download latest release
wget https://github.com/vnykmshr/markgo/releases/download/v1.5.0/markgo-v1.5.0-linux-amd64.tar.gz
tar -xzf markgo-v1.5.0-linux-amd64.tar.gz

# Initialize with preview support
./markgo init --quick
```

### Upgrade from v1.4.x
```bash
# Stop existing service
sudo systemctl stop markgo

# Replace binary
sudo cp markgo-v1.5.0 /usr/local/bin/markgo

# Update configuration (optional - preview disabled by default)
echo "PREVIEW_ENABLED=true" >> .env

# Start service
sudo systemctl start markgo
```

## 📝 Preview Service Usage

### Enable Preview Service
```bash
# In your .env file
PREVIEW_ENABLED=true
PREVIEW_PORT=8081
PREVIEW_MAX_SESSIONS=10
PREVIEW_SESSION_TIMEOUT=30m
```

### Create Preview Session
```bash
curl -X POST http://localhost:3000/api/preview/sessions \
  -H "Content-Type: application/json" \
  -d '{"article_slug": "my-draft-article"}'
```

### Real-Time Editing Workflow
1. Create a draft article in your articles directory
2. Create a preview session via API or admin interface
3. Open the preview URL in your browser
4. Edit the article file - changes appear instantly in browser
5. WebSocket automatically reloads content on file changes

## 🔧 Migration Notes

### Existing Deployments
- **✅ Zero downtime upgrade** - Preview service is opt-in
- **✅ All existing APIs unchanged**
- **✅ Configuration backward compatible**
- **✅ No database migrations required**

### New Configuration
Preview service is **disabled by default**. To enable:
1. Set `PREVIEW_ENABLED=true` in your environment
2. Optionally configure preview-specific settings
3. Restart the service

## 📚 Documentation Updates

- **[Preview API Documentation](docs/api.md#preview-api-endpoints)** - Complete API reference
- **[Configuration Guide](docs/configuration.md#preview-service-configuration)** - New preview settings
- **[System Overview](docs/system-overview.md)** - Architecture with preview service

## 🐛 Bug Fixes & Security

### Critical Fixes
- **Security**: Fixed WebSocket origin validation vulnerability
- **Concurrency**: Resolved race condition in session access patterns
- **Architecture**: Improved service coupling with interface dependencies

### Performance Improvements
- Optimized WebSocket message broadcasting
- Reduced memory allocation in session management
- Enhanced file watching efficiency

## 🎯 What's Next?

### v1.6.0 Preview
- Enhanced preview UI with live editing capabilities
- Template hot-reload integration
- Multi-user collaborative editing features
- Advanced preview customization options

## 💡 Developer Workflow Enhancement

### Before v1.5.0
```bash
# Edit article
vim articles/my-post.md

# Refresh browser manually
# Repeat cycle...
```

### After v1.5.0
```bash
# Create preview session
curl -X POST localhost:3000/api/preview/sessions -d '{"article_slug":"my-post"}'

# Edit article
vim articles/my-post.md
# Browser automatically reloads with changes!
```

## 📦 Release Assets

### Multi-Platform Binaries

| Platform | Architecture | File | Size |
|----------|-------------|------|------|
| Linux | x86_64 | `markgo-v1.5.0-linux-amd64.tar.gz` | ~9.2MB |
| Linux | ARM64 | `markgo-v1.5.0-linux-arm64.tar.gz` | ~8.1MB |
| macOS | Intel | `markgo-v1.5.0-darwin-amd64.tar.gz` | ~9.3MB |
| macOS | Apple Silicon | `markgo-v1.5.0-darwin-arm64.tar.gz` | ~8.4MB |
| Windows | x86_64 | `markgo-v1.5.0-windows-amd64.exe.zip` | ~9.4MB |
| Windows | ARM64 | `markgo-v1.5.0-windows-arm64.exe.zip` | ~8.2MB |

### Verification
All binaries include SHA256 checksums for integrity verification.

## 🌟 Community Impact

### Developer Experience Improvements
- **Instant Feedback**: Real-time preview eliminates refresh cycles
- **Professional Workflow**: WebSocket-based updates match modern editors
- **Collaborative Ready**: Multi-client preview support
- **Production Safe**: Security-first design with proper validation

### Technical Excellence
- **Zero Breaking Changes**: Seamless upgrade path
- **Comprehensive Testing**: All new features thoroughly tested
- **Documentation Complete**: Full API and configuration documentation
- **Security Audited**: WebSocket security and origin validation

---

**🎉 MarkGo Engine v1.5.0 delivers professional-grade live preview capabilities!**

*Built with ❤️ and Go for the ultimate developer experience.*