# GopherDrive ☁️

A robust, highly concurrent, and scalable self-hosted cloud storage REST API built with Go.  
GopherDrive provides secure file storage, folder hierarchy management, and user authentication, adhering to the principles of Clean Architecture and 12-Factor App methodology.

## Key Features:

* **True Streaming Uploads**: Handles files up to 5TB without loading them into RAM or temp disk files, using Go's `io.Reader` and `multipart.Reader`.
* **S3 Compatible Storage**: Integrates with MinIO/AWS S3 using `aws-sdk-go-v2` and `transfermanager` for concurrent, chunked multipart uploads.
* **Metadata & Hierarchy**: Manages virtual folder structures and file metadata using PostgreSQL.
* **Secure Authentication**: JWT-based auth with short-lived Access Tokens and HttpOnly Refresh Tokens stored in Redis.
* **Transactional Consistency**: Manual rollbacks for S3 objects if database metadata insertion fails.
* **12-Factor Config**: Pure environment-variable-based configuration (`envconfig`), completely agnostic of file paths.
---  
## 🏗 Architecture & Technical Decisions

The codebase is divided into `core` (infrastructure), `domain` (business entities & invariants), and `features` (Auth, Storage), ensuring zero import cycles and loose coupling.
### 1. True Streaming via `io.Reader` (Zero-RAM Overhead)
Instead of using `r.ParseMultipartForm()` (which buffers data into memory/disk), the API uses `r.MultipartReader()`. The file part is extracted and passed directly to the S3 `transfermanager` as an `io.Reader`. The server acts purely as a network pipe, passing bytes from the client's TCP socket directly to S3.
### 2. On-the-fly Byte Counting
S3 requires the file size for metadata, but in a true streaming scenario, the size is unknown until the stream ends. To solve this, a custom `sizeCounter` wrapper implements the `io.Reader` interface. It counts bytes as they pass through to S3, providing 100% accurate file sizes for the PostgreSQL database without requiring a separate S3 `HeadObject` request.
### 3. S3 ↔ Database Consistency
If a file is successfully uploaded to S3, but the subsequent PostgreSQL `INSERT` fails (e.g., due to a lost connection), the system performs an immediate cleanup. It uses `context.WithoutCancel(ctx)` to ensure the S3 `DeleteObject` command executes successfully even if the client has already canceled the HTTP request.
### 4. Security & DDoS Protection:
* **Slowloris Protection**: Configured `ReadHeaderTimeout` in the HTTP server.
* **Memory Exhaustion Protection**: Implemented `io.LimitReader` when parsing small multipart fields (like `folder_id`) to prevent malicious actors from sending gigabytes of text into a form field.
* **Header Injection Protection**: Uses `url.PathEscape` for filenames in `Content-Disposition` headers during downloads.
---  
## 🛠 Tech Stack
* **Language**: Go 1.26+
* **HTTP Routing**: `go-chi/chi`
* **Database**: PostgreSQL (Driver: `jackc/pgx/v5`)
* **Migrations**: `golang-migrate`
* **Cache / Sessions**: Redis
* **Object Storage**: MinIO / AWS S3 (`aws-sdk-go-v2`)
* **Logging**: `log/slog` (Structured Logging)
* **Configuration**: `kelseyhightower/envconfig`
* **Deployment**: Docker & Docker Compose
---
## ⚙️️ Local Setup & Getting Started
### Prerequisites
* Docker & Docker Compose
* `make` utility (optional, for local development)

### Installation
1. **Clone the repository:**
    ```bash
   git clone https://github.com/nemokiit/GopherDrive.git
   cd GopherDrive
    ```
2. **Setup Environment Variables:**
   Copy the example environment file and fill in your secrets.
    ```bash
   cp .env.example .env
   ```
#### Option A: Fully Containerized (Easiest)
You can run the entire infrastructure, database migrations, and the Go application with a single command.
The application will automatically wait for the database to be healthy and migrations to complete before starting.

3. **Build and start all containers:**
   ```bash
   docker-compose up --build -d
   ```

The API will be available at `http://localhost:8080`.
#### Option B: Local Development Mode
If you want to run the Go code on your host machine for development/debugging, while keeping the infrastructure in Docker:

3. **Start the Infrastructure (PostgreSQL, Redis, MinIO):**
   ```bash
   make env-up
   ```
4. **Apply Database Migrations:**
   ```bash
   make migrate-up  
   ```
5. **Run the Application:**
   ```bash
   make run
   ```

*The API will be available at `http://localhost:8080`.*
*MinIO Console is available at `http://localhost:9001`.*

---
## 📡 API Endpoints Overview
### Authentication (`/user`)
* `POST /user/register` - Register a new account.
* `POST /user/login` - Authenticate and receive JWT.
* `POST /user/refresh` - Refresh access token via HttpOnly cookie.
### Storage (`/folder`)
* `POST /folder` - Create a new virtual folder.
* `GET /folder` - Get root folder contents.
* `GET /folder/{id}` - Get folder contents (sub-folders and files).
* `DELETE /folder/{id}` - Recursively delete a folder, its metadata, and batched removal of S3 objects.
### Files (`/file`)
* `POST /file` - Upload a file (Requires `multipart/form-data` with `folder_id` preceding the `file` part).
* `GET /file/{id}/download` - Stream a file directly from S3 to the client.
* `DELETE /file/{id}` - Delete a file from the database and S3.  
