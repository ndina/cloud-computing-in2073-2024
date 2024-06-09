
# Cloud Computing Course Project at TUM

This repository contains the project for the Cloud Computing course at the Technical University of Munich (TUM). The project demonstrates the use of Docker and Docker Compose to build, deploy, and manage microservices in a cloud environment.

## Project Structure

The project is organized into the following directories:

```
/my_project
├── api
│   └── books
│       ├── get
│       │   ├── main.go
│       │   └── Dockerfile
│       ├── post
│       │   ├── main.go
│       │   └── Dockerfile
│       ├── put
│       │   ├── main.go
│       │   └── Dockerfile
│       └── delete
│           ├── main.go
│           └── Dockerfile
├── web
│   ├── main.go
│   ├── Dockerfile
│   ├── views
│   │   ├── index.html
│   │   ├── authors.html
│   │   ├── years.html
│   │   └── search-bar.html
│   └── css
│       └── styles.css
├── nginx
│   └── nginx.conf
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Services

### NGINX

The NGINX service acts as a reverse proxy, routing traffic to the appropriate microservice based on the HTTP request method.

### Web Service

The Web Service serves the web application, providing HTML pages for the user interface.

### API Services

#### GET Service

Handles HTTP GET requests to retrieve book information.

#### POST Service

Handles HTTP POST requests to create new book entries.

#### PUT Service

Handles HTTP PUT requests to update existing book entries.

#### DELETE Service

Handles HTTP DELETE requests to remove book entries.

### MongoDB

A MongoDB instance is used to store book information.

## Setup and Running the Project

### Prerequisites

- Docker
- Docker Compose

### Steps to Build and Run the Project

1. **Clone the Repository**

```sh
git clone <repository-url>
cd my_project
```

2. **Build and Start the Containers**

```sh
docker-compose up --build
```

This command will build the Docker images for each service and start the containers.

3. **Access the Application**

- NGINX: [http://localhost](http://localhost)
- Web Service: [http://localhost](http://localhost)
- API Endpoints:
  - `GET /api/books`
  - `POST /api/books`
  - `PUT /api/books`
  - `DELETE /api/books/:id`

You can use tools like `curl` or Postman to test the API endpoints.

### Testing the API

Example using `curl` to test the GET endpoint:

```sh
curl -X GET http://localhost/api/books
```

## Configuration

### NGINX Configuration

The NGINX configuration file is located at `nginx/nginx.conf` and routes traffic based on the request methods.

```nginx
events {}

http {
    server {
        listen 80;

        location /api/books {
            if ($request_method = GET) {
                proxy_pass http://get_service:80;
            }
            if ($request_method = POST) {
                proxy_pass http://post_service:80;
            }
            if ($request_method = PUT) {
                proxy_pass http://put_service:80;
            }
            if ($request_method = DELETE) {
                proxy_pass http://delete_service:80;
            }
        }

        location / {
            proxy_pass http://web_service:80;
        }
    }
}
```

### Environment Variables

Each service connects to MongoDB using the `MONGO_URI` environment variable, which is set in the `docker-compose.yml` file:

```yaml
environment:
  - MONGO_URI=mongodb://mongodb:27017
```

## Contributing

If you wish to contribute to this project, please fork the repository and submit a pull request with your changes.

## License

This project is licensed under the MIT License.

## Contact

For any inquiries, please contact the course instructor or the project maintainer.

---

This README file provides an overview of the project, setup instructions, and relevant configurations. Feel free to modify it to better suit your needs.
