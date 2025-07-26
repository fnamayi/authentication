# 🌐 Real-Time Forum - Modern Discussion Platform

A comprehensive, feature-rich real-time forum application built with Go and vanilla JavaScript. This project demonstrates modern web development practices with real-time communication, advanced messaging systems, and a beautiful dark-themed interface.

## 📖 Project Overview

This forum application provides a complete social platform where users can:
- Create accounts and manage profiles
- Post discussions with categories and images
- Engage in real-time private messaging
- Comment on posts with nested threading
- Like/dislike content and interact socially
- Track online users and receive live notifications

The project emphasizes clean architecture, real-time features, and modern UI/UX design principles.

## ✨ Key Features

### 🔐 Authentication & User Management
- **Multi-method Authentication**: Traditional registration and OAuth (Google, GitHub)
- **Comprehensive Profiles**: Nickname, age, gender, first/last name, email, avatar
- **Secure Sessions**: Cookie-based authentication with automatic logout
- **Profile Management**: Edit profiles and upload custom avatars

### 💬 Advanced Messaging System
- **Real-time Private Messaging**: Instant messaging between users with WebSocket
- **Message Timestamps**: Smart time formatting (now, 5m, Yesterday 3:45 PM)
- **Message Grouping**: Consecutive messages from same sender grouped visually
- **Date Separators**: Automatic date dividers for multi-day conversations
- **Online Status**: Real-time user presence indicators
- **Message History**: Persistent message storage with pagination
- **Unread Counts**: Track unread messages per conversation
- **Browser Notifications**: Desktop notifications for new messages

### 📝 Discussion Features
- **Rich Post Creation**: Title, content, categories, and image uploads
- **Nested Comments**: Multi-level comment threading with replies
- **Like/Dislike System**: Express opinions on posts and comments
- **Category Filtering**: Organize content by predefined categories
- **Real-time Updates**: Live updates for new posts and comments
- **Image Support**: Upload and display images in posts

### 🎨 Modern Interface
- **Dark Theme**: Beautiful dark UI that's easy on the eyes
- **Responsive Design**: Optimized for desktop, tablet, and mobile
- **Single Page Application**: Smooth navigation without page reloads
- **Component Architecture**: Modular, reusable frontend components
- **Real-time UI**: Interface updates automatically with new data
- **Smooth Animations**: Polished transitions and hover effects

## 🛠️ Technology Stack

### Backend Technologies
- **Language**: Go (Golang) 1.21+
- **Database**: SQLite3 with comprehensive schema
- **WebSocket**: Gorilla WebSocket for real-time communication
- **Authentication**: Session-based + OAuth 2.0
- **HTTP Router**: Standard Go net/http with custom routing
- **File Handling**: Multipart form data for image uploads

### Frontend Technologies
- **Languages**: HTML5, CSS3, Vanilla JavaScript (ES6+)
- **Architecture**: Single Page Application (SPA)
- **Styling**: CSS Grid, Flexbox, CSS Variables for theming
- **Real-time**: WebSocket API for live updates
- **Responsive**: Mobile-first responsive design
- **No Frameworks**: Pure vanilla JavaScript for maximum performance

### External Services
- **OAuth Providers**: Google OAuth 2.0, GitHub OAuth
- **File Storage**: Local filesystem for uploaded images
- **Browser APIs**: Notification API, Web Audio API for sounds

## 🏗️ Project Structure

```
real-time-forum/
├── 📁 frontend/static/          # Frontend assets
│   ├── 📁 css/                  # Stylesheets
│   │   ├── main.css            # Core styles and dark theme
│   │   └── messaging.css       # Messaging interface styles
│   ├── 📁 js/                   # JavaScript modules
│   │   ├── 📁 components/       # Reusable UI components
│   │   ├── 📁 pages/           # Page-specific controllers
│   │   │   ├── home.js         # Homepage functionality
│   │   │   ├── posts.js        # Post viewing and creation
│   │   │   ├── messages.js     # Messaging interface
│   │   │   ├── profile.js      # User profile management
│   │   │   └── create-post.js  # Post creation form
│   │   ├── api.js              # API client and HTTP requests
│   │   ├── router.js           # SPA routing system
│   │   ├── utils.js            # Utility functions
│   │   └── main.js             # Application initialization
│   └── index.html              # Single page application entry
├── 📁 internal/                 # Go backend packages
│   ├── 📁 auth/                # Authentication logic
│   │   ├── auth.go             # Session management
│   │   ├── google.go           # Google OAuth integration
│   │   └── github.go           # GitHub OAuth integration
│   ├── 📁 database/            # Database layer
│   │   └── database.go         # SQLite setup and schema
│   ├── 📁 handlers/            # HTTP request handlers
│   │   ├── auth.go             # Authentication endpoints
│   │   ├── posts.go            # Post CRUD operations
│   │   ├── comments.go         # Comment management
│   │   ├── messaging.go        # Private messaging API
│   │   ├── users.go            # User management
│   │   └── uploads.go          # File upload handling
│   ├── 📁 models/              # Data structures
│   │   └── models.go           # All data models and types
│   └── 📁 websocket/           # Real-time communication
│       └── websocket.go        # WebSocket hub and client management
├── 📁 tests/                   # Test files
│   ├── auth_test.go            # Authentication tests
│   ├── models_test.go          # Model validation tests
│   ├── utils_test.go           # Utility function tests
│   └── websocket_test.go       # WebSocket functionality tests
├── 📁 uploads/                 # User uploaded files
│   └── 📁 post-images/         # Post image storage
├── main.go                     # Application entry point
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── forum.db                    # SQLite database file
└── README.md                   # Project documentation
```

## 🏗️ Architecture

## 🗄️ Database Schema

The application uses SQLite with the following key tables:

### Core Tables
- **users**: User accounts with profile information
- **sessions**: User session management
- **google_auth** / **github_auth**: OAuth provider data
- **posts**: Forum posts with categories and content
- **comments**: Nested comments with parent-child relationships
- **likes**: Like/dislike tracking for posts and comments

### Messaging Tables
- **messages**: Private messages between users
- **conversations**: Conversation metadata and last message info
- **online_users**: Real-time user presence tracking

### Key Features
- **Foreign Key Constraints**: Ensure data integrity
- **Indexes**: Optimized for common queries
- **Timestamps**: Track creation and modification times
- **Soft Deletes**: Preserve data relationships

## 🚀 Getting Started

### Prerequisites
- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Modern Browser**: Chrome, Firefox, Safari, or Edge
- **Git**: For cloning the repository

### Installation

1. **Clone the Repository**
   ```bash
   git clone https://github.com/your-username/real-time-forum.git
   cd real-time-forum
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure OAuth (Optional)**

   For Google OAuth:
   ```bash
   export GOOGLE_CLIENT_ID="your-google-client-id"
   export GOOGLE_CLIENT_SECRET="your-google-client-secret"
   ```

   For GitHub OAuth:
   ```bash
   export GITHUB_CLIENT_ID="your-github-client-id"
   export GITHUB_CLIENT_SECRET="your-github-client-secret"
   ```

4. **Run the Application**
   ```bash
   # Default port (8080)
   go run main.go

   # Custom port
   PORT=8081 go run main.go
   ```

5. **Access the Forum**
   Open your browser and navigate to:
   - Default: `http://localhost:8080`
   - Custom port: `http://localhost:YOUR_PORT`

### First Time Setup
1. **Create an Account**: Register with email or use OAuth
2. **Complete Profile**: Add your information and avatar
3. **Explore**: Browse posts, create content, and start messaging!

## 🔧 Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `GOOGLE_CLIENT_ID`: Google OAuth client ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth client secret
- `GITHUB_CLIENT_ID`: GitHub OAuth client ID
- `GITHUB_CLIENT_SECRET`: GitHub OAuth client secret

### Database
- **File**: `forum.db` (created automatically)
- **Type**: SQLite3
- **Location**: Project root directory

## 📡 API Endpoints

### Authentication
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `POST /api/logout` - User logout
- `GET /api/user` - Get current user info
- `GET /auth/google` - Google OAuth login
- `GET /auth/github` - GitHub OAuth login

### Posts & Comments
- `GET /api/posts` - Get all posts
- `POST /api/posts` - Create new post
- `GET /api/posts/{id}` - Get specific post
- `GET /api/comments/{postId}` - Get post comments
- `POST /api/comment` - Create comment
- `POST /api/like` - Like/dislike post or comment

### Messaging
- `GET /api/conversations` - Get user conversations
- `GET /api/messages` - Get conversation messages
- `POST /api/messages/send` - Send new message
- `POST /api/messages/read` - Mark messages as read
- `GET /api/online-users` - Get online users

### WebSocket
- `WS /ws` - Real-time communication endpoint

## 🏛️ Architecture Details

### Backend Architecture
The Go backend follows a clean, modular architecture:

- **main.go**: Application entry point and server setup
- **internal/auth**: Authentication logic and OAuth integration
- **internal/database**: Database connection and schema management
- **internal/handlers**: HTTP request handlers for all endpoints
- **internal/models**: Data structures and business logic
- **internal/websocket**: Real-time WebSocket communication hub

### Frontend Architecture
The frontend is a Single Page Application built with vanilla JavaScript:

- **Component-Based**: Reusable UI components
- **Page Controllers**: Separate logic for each page/route
- **API Client**: Centralized HTTP request handling
- **Router**: Client-side routing for SPA navigation
- **Real-time Updates**: WebSocket integration for live features

### WebSocket System
Advanced WebSocket implementation featuring:

- **Connection Hub**: Centralized connection management
- **User Sessions**: Multiple sessions per user support
- **Message Broadcasting**: Targeted message delivery
- **Heartbeat Monitoring**: Connection health checking
- **Automatic Reconnection**: Robust connection handling

## 🎨 UI/UX Features

### Dark Theme
- **Consistent Color Scheme**: Carefully chosen dark colors
- **CSS Variables**: Easy theme customization
- **Eye-friendly**: Reduced eye strain for extended use
- **Modern Aesthetics**: Clean, professional appearance

### Responsive Design
- **Mobile-First**: Optimized for mobile devices
- **Flexible Layouts**: CSS Grid and Flexbox
- **Touch-Friendly**: Large touch targets on mobile
- **Cross-Browser**: Compatible with all modern browsers

### User Experience
- **Smooth Animations**: Polished transitions and effects
- **Real-time Feedback**: Instant visual feedback for actions
- **Intuitive Navigation**: Clear, logical user interface
- **Accessibility**: Semantic HTML and keyboard navigation

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./tests/...

# Run specific test file
go test ./tests/auth_test.go

# Run with verbose output
go test -v ./tests/...

# Run with coverage
go test -cover ./tests/...
```

### Test Coverage
- **Authentication**: Login, registration, OAuth flows
- **Models**: Data validation and serialization
- **WebSocket**: Connection management and messaging
- **Utilities**: Helper functions and common operations

## 🚀 Deployment

### Production Considerations
1. **Environment Variables**: Set all required OAuth credentials
2. **Database**: Consider PostgreSQL for production scale
3. **Static Files**: Use CDN for better performance
4. **HTTPS**: Enable SSL/TLS for secure communication
5. **Process Management**: Use systemd or similar for service management

### Docker Deployment (Optional)
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o forum .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/forum .
COPY --from=builder /app/frontend ./frontend
EXPOSE 8080
CMD ["./forum"]
```

## 🤝 Contributing

1. **Fork the Repository**
2. **Create Feature Branch**: `git checkout -b feature/amazing-feature`
3. **Commit Changes**: `git commit -m 'Add amazing feature'`
4. **Push to Branch**: `git push origin feature/amazing-feature`
5. **Open Pull Request**

### Development Guidelines
- Follow Go conventions and best practices
- Write tests for new features
- Update documentation as needed
- Use meaningful commit messages
- Ensure responsive design for UI changes

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Community**: For excellent documentation and libraries
- **Gorilla WebSocket**: For robust WebSocket implementation
- **Modern Web Standards**: For enabling rich client-side applications
- **Open Source Community**: For inspiration and best practices

## 📞 Support

If you encounter any issues or have questions:

1. **Check Documentation**: Review this README and code comments
2. **Search Issues**: Look for existing GitHub issues
3. **Create Issue**: Open a new issue with detailed information
4. **Community**: Engage with other developers in discussions

---

**Built with ❤️ using Go and Vanilla JavaScript**