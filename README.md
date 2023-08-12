## Face Recognition Microservices Project Documentation

This documentation provides an overview of the Face Recognition Microservices Project, including details about the project's architecture, API endpoints, setup instructions, and usage guidelines.

### Table of Contents

1. [Introduction](#introduction)
2. [Architecture](#architecture)
3. [API Endpoints](#api-endpoints)
4. [Getting Started](#getting-started)
5. [Running the Project](#running-the-project)
6. [Accessing Microservices](#accessing-microservices)
7. [Stopping the Project](#stopping-the-project)

---

### Introduction

The Face Recognition Microservices Project demonstrates the usage of Docker, Docker Compose, FastAPI (Python), and Gin (Golang) to create two microservices that perform face recognition tasks. The project enables communication between microservices and utilizes the face-recognition library for encoding and comparing face images.

### Architecture

The project consists of the following components:

1. **Microservice 1 (Python/FastAPI):**
   - Provides an API for encoding and comparing face images.
   - Endpoint: `/face-recognition/`

2. **Microservice 2 (Golang/Gin):**
   - Orchestrates communication between Microservice 1 and a local MongoDB container.
   - Endpoint: `/face-recognition/`

3. **Local MongoDB Container:**
   - Used for data persistence and storage.
   - Accessible at: `mongodb://localhost:27017`

### API Endpoints

#### Microservice 1 (Python/FastAPI)

**Endpoint:** `/face-recognition/`
**Method:** POST

**Request:**
- Content-Type: `multipart/form-data`
- Parameters:
  - `file1`: Face image file (JPEG or PNG format)
  - `file2`: Face image file (JPEG or PNG format)

**Response:**
- Content-Type: `application/json`
- Body: JSON object containing a boolean value which tells whether the faces in the two images matched or not.
  ```json
  {
    "matched": true
  }
  ```

#### Microservice 2 (Golang/Gin)

**Endpoint:** `/face-recognition/`
**Method:** POST

**Request:**
- Content-Type: `multipart/form-data`
- Parameters:
  - `file1`: Face image file (JPEG or PNG format)
  - `file2`: Face image file (JPEG or PNG format)

**Response:**
- Content-Type: `application/json`
- Body: JSON object indicating the success of recording the transaction and the matching result of the images.
  ```json
  {
    "message": "Transaction recorded and saved images",
    "matched": true
  }
  ```

### Getting Started

Follow these steps to set up and run the Face Recognition Microservices Project on your local machine.

1. Clone this repository:
   ```bash
   git clone https://github.com/SudeepKumarS/face-recognition.git
   ```

2. Navigate to the project directory:
   ```bash
   cd face-recognition
   ```

### Running the Project

1. Install Docker and Docker Compose if not already installed.

2. Create a directory named `data` in the project root for MongoDB data persistence. This is ignored by git as it is added in the git ignore file.

### Build and Start Containers

```bash
docker-compose up -d
```

### Accessing Microservices

- Microservice 1 (Python/FastAPI) will be accessible at: `http://localhost:8000`
- Microservice 2 (Golang/Gin) will be accessible at: `http://localhost:8080`
- Mongo (Local) will be accessible at: `http://localhost:27017` 

### Stopping the Project

To stop and remove the containers:

```bash
docker-compose down
```
