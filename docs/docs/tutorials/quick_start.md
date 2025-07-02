---
title: "Clone a Project - Quick Start"
sidebar_label: "Clone a Project"
sidebar_position: 2
hide_table_of_contents: false

tags:
  - Getting Started
  - Clone
  - Existing Project
  - Rill Developer
---

# Clone a Project - Quick Start

This guide will help you get started with an existing Rill project by cloning it from a repository and setting it up locally.

## Prerequisites

Before you begin, make sure you have:

- **Rill CLI** installed ([Installation Guide](/tutorials/installation))
- **Git** installed
- **Access to the project repository** (GitHub, GitLab, etc.)
- **Required credentials** for data sources (if any)

## Step 1: Clone the Repository

### From GitHub

```bash
# Clone the repository
git clone https://github.com/username/rill-project.git
cd rill-project

# Or clone a specific branch
git clone -b main https://github.com/username/rill-project.git
cd rill-project
```

### From GitLab

```bash
# Clone from GitLab
git clone https://gitlab.com/username/rill-project.git
cd rill-project
```

### From Private Repository

```bash
# Using SSH (if you have SSH keys set up)
git clone git@github.com:username/rill-project.git
cd rill-project

# Using HTTPS with personal access token
git clone https://username:token@github.com/username/rill-project.git
cd rill-project
```

## Step 2: Explore the Project Structure

A typical Rill project contains:

```
rill-project/
├── rill.yaml              # Project configuration
├── sources/               # Data source definitions
│   ├── database.yaml      # Database connections
│   ├── api.yaml          # API endpoints
│   └── files.yaml        # File-based sources
├── models/                # SQL transformations
│   ├── staging/          # Staging models
│   ├── marts/            # Business logic models
│   └── metrics/          # Metric definitions
├── dashboards/           # Dashboard configurations
│   └── main_dashboard.yaml
├── alerts/               # Alert definitions
├── .env                  # Environment variables (not in git)
├── .gitignore           # Git ignore rules
└── README.md            # Project documentation
```

## Step 3: Set Up Environment Variables

Most Rill projects require environment variables for data source connections:

### Create Environment File

```bash
# Copy the example environment file (if it exists)
cp .env.example .env

# Or create a new one
touch .env
```

### Configure Credentials

Edit the `.env` file with your credentials:

```bash
# Database connections
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=mydb
DATABASE_USER=myuser
DATABASE_PASSWORD=mypassword

# API keys
API_KEY=your_api_key_here
API_SECRET=your_api_secret_here

# Cloud storage
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1

# Google Cloud
GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account.json
```

### Load Environment Variables

```bash
# Load environment variables
source .env

# Or use a tool like direnv
echo "source .env" >> .envrc
direnv allow
```

## Step 4: Install Dependencies

### Python Dependencies (if any)

```bash
# Install Python dependencies
pip install -r requirements.txt

# Or using pipenv
pipenv install

# Or using poetry
poetry install
```

### Node.js Dependencies (if any)

```bash
# Install Node.js dependencies
npm install

# Or using yarn
yarn install
```

## Step 5: Configure Data Sources

### Check Data Source Configuration

Review the data sources in the `sources/` directory:

```yaml
# sources/database.yaml
type: postgres
host: {{ .env.DATABASE_HOST }}
port: {{ .env.DATABASE_PORT }}
database: {{ .env.DATABASE_NAME }}
username: {{ .env.DATABASE_USER }}
password: {{ .env.DATABASE_PASSWORD }}
```

### Update Connection Details

Modify the source files with your specific connection details:

```yaml
# Example: Update database connection
type: postgres
host: your-database-host.com
port: 5432
database: your_database
username: your_username
password: your_password
```

## Step 6: Start Rill Developer

### Start the Development Server

```bash
# Start Rill Developer
rill dev
```

This will:
- Start the web UI at `http://localhost:8080`
- Process all data sources and models
- Show any errors or warnings

### Check for Issues

Look for common issues in the terminal output:

```
[INFO] Processing source: database
[ERROR] Connection failed: authentication failed
[WARN] Model 'daily_sales' has no data
```

## Step 7: Explore the Project

### Data Explorer

1. **Navigate to Data Explorer** in the UI
2. **Check Data Sources** - verify connections are working
3. **Preview Data** - examine the raw data
4. **Review Models** - understand the transformations

### Dashboard

1. **Open Dashboards** section
2. **View Existing Dashboards** - see what's already built
3. **Test Interactivity** - try filters and drill-downs
4. **Check Data Freshness** - verify data is up to date

## Common Issues and Solutions

### **Connection Errors**

#### Database Connection Failed
```bash
# Check if database is accessible
psql -h hostname -p port -U username -d database

# Verify credentials in .env file
cat .env | grep DATABASE
```

#### API Connection Issues
```bash
# Test API endpoint
curl -H "Authorization: Bearer $API_KEY" https://api.example.com/data

# Check API key format and permissions
```

### **Data Issues**

#### No Data in Models
```sql
-- Check if source has data
SELECT COUNT(*) FROM {{ ref("source_name") }}

-- Verify date filters
SELECT MIN(date), MAX(date) FROM {{ ref("source_name") }}
```

#### Schema Mismatches
```bash
# Check source schema
rill source describe source_name

# Compare with model expectations
rill model describe model_name
```

### **Environment Issues**

#### Missing Environment Variables
```bash
# Check which variables are missing
rill dev --dry-run

# Set required variables
export MISSING_VAR=value
```

#### Permission Issues
```bash
# Check file permissions
ls -la .env
chmod 600 .env

# Check directory permissions
ls -la sources/
```

## Advanced Setup

### **Multiple Environments**

Set up different configurations for dev/staging/prod:

```bash
# Development
rill dev --env dev

# Staging
rill dev --env staging

# Production
rill dev --env prod
```

### **Custom Ports**

```bash
# Use different port
rill dev --port 8081

# Specify host
rill dev --host 0.0.0.0 --port 8080
```

### **Debug Mode**

```bash
# Enable debug logging
rill dev --debug

# Show SQL queries
rill dev --show-sql
```

## Next Steps

### **Make Changes**

1. **Modify Models** - update SQL transformations
2. **Add Data Sources** - connect new data
3. **Create Dashboards** - build visualizations
4. **Add Alerts** - set up monitoring

### **Version Control**

```bash
# Check project status
git status

# Make your changes
git add .
git commit -m "Add new dashboard"

# Push to repository
git push origin main
```

### **Deploy Changes**

```bash
# Deploy to staging
rill deploy --env staging

# Deploy to production
rill deploy --env prod
```

## Project-Specific Setup

### **Check Project Documentation**

Always read the project's `README.md` for specific setup instructions:

```bash
# View project documentation
cat README.md

# Check for setup scripts
ls -la scripts/
```

### **Run Setup Scripts**

Many projects include setup scripts:

```bash
# Run setup script
./scripts/setup.sh

# Or using make
make setup

# Or using npm
npm run setup
```

### **Verify Project Requirements**

Check if the project has specific requirements:

```bash
# Check Rill version
rill --version

# Check Python version
python --version

# Check Node.js version
node --version
```

## Getting Help

### **Project-Specific Issues**

- **Check Issues** on the repository
- **Read Documentation** in the project
- **Contact Maintainers** via GitHub/GitLab

### **General Rill Issues**

- **Documentation**: [docs.rilldata.com](https://docs.rilldata.com)
- **Community**: [community.rilldata.com](https://community.rilldata.com)
- **GitHub**: [github.com/rilldata/rill](https://github.com/rilldata/rill)

## Summary

You've successfully:
- ✅ Cloned an existing Rill project
- ✅ Set up environment variables
- ✅ Configured data sources
- ✅ Started Rill Developer
- ✅ Explored the project structure

Now you're ready to explore, modify, and contribute to the project! 🚀 