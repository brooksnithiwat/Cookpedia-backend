# Cookpedia-Backend

## Tech Stack
- **Language:** Go  
- **Framework:** Echo  
- **Database:** PostgreSQL  

## Installation
1. **Clone the Repository**  
```bash
git clone https://github.com/brooksnithiwat/cookpedia.git
cd cookpedia
```
2. **Install Go AND Verify Installation**  
```
go version
```
3. **Create .env File**  
```
DB_USER=your_db_username
DB_PASSWORD=your_password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cookpedia
GOOGLE_REDIRECT_URL=your_google_redirect_url
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
JWT_SECRET=your_jwt_secret
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_anon_keys
SUPABASE_BUCKET_NAME=your_supabase_bucket_name
```
4. **Install Dependencies**  
```
go mod tidy
```
5. **Run the Application**  
```
go run main.go
```





